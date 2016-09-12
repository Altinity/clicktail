// Package mysql parses the mysql slow query log
package mysql

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"

	"github.com/honeycombio/honeytail/event"
)

// 3 sample log entries
//
// # Time: 2016-04-01T00:31:09.817887Z
// # User@Host: root[root] @ localhost []  Id:   233
// # Query_time: 0.008393  Lock_time: 0.000154 Rows_sent: 1  Rows_examined: 357
// SET timestamp=1459470669;
// show status like 'Uptime';
//
// # Time: 2016-04-01T00:31:09.853523Z
// # User@Host: root[root] @ localhost []  Id:   233
// # Query_time: 0.020424  Lock_time: 0.000147 Rows_sent: 494  Rows_examined: 494
// SET timestamp=1459470669;
// SHOW /*innotop*/ GLOBAL VARIABLES;
//
// # Time: 2016-04-01T00:31:09.856726Z
// # User@Host: root[root] @ localhost []  Id:   233
// # Query_time: 0.000021  Lock_time: 0.000000 Rows_sent: 494  Rows_examined: 494
// SET timestamp=1459470669;
// # administrator command: Ping;
//
// We should ignore the administrator command entry; the stats it presents (eg rows_sent)
// are actually for the previous command

var (
	reTime       = myRegexp{regexp.MustCompile("^# Time: (?P<time>[^ ]+)Z *$")}
	reAdminPing  = myRegexp{regexp.MustCompile("^# administrator command: Ping; *$")}
	reUser       = myRegexp{regexp.MustCompile("^# User@Host: (?P<user>[^#]+) @ (?P<host>[^#]+).*$")}
	reQueryStats = myRegexp{regexp.MustCompile("^# Query_time: (?P<queryTime>[0-9.]+) *Lock_time: (?P<lockTime>[0-9.]+) *Rows_sent: (?P<rowsSent>[0-9]+) *Rows_examined: (?P<rowsExamined>[0-9]+) *$")}
	reSetTime    = myRegexp{regexp.MustCompile("^SET timestamp=(?P<unixTime>[0-9]+);$")}
	reQuery      = myRegexp{regexp.MustCompile("^(?P<query>[^#]*).*$")}
	reUse        = myRegexp{regexp.MustCompile("^(?i)use ")}

	rdsStr  = "rds"
	ec2Str  = "ec2"
	selfStr = "self"
)

const timeFormat = "2006-01-02T15:04:05.000000"

type Options struct {
	Host          string `long:"host" description:"MySQL host in the format (address:port)"`
	User          string `long:"user" description:"MySQL username"`
	Pass          string `long:"pass" description:"MySQL password"`
	QueryInterval uint   `long:"interval" description:"interval for querying the MySQL DB in seconds" default:"30"`
}

type Parser struct {
	conf       Options
	wg         sync.WaitGroup
	nower      Nower
	hostedOn   *string
	readOnly   *bool
	replicaLag *int64
	role       *string
}

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (n *RealNower) Now() time.Time {
	return time.Now().UTC()
}

// rawEvent is the unparsed multi-line entry representing a single slow query
type rawEvent struct {
	lines []string
}

// slowQuery represents the structured form of a query from the slow query log
type SlowQuery struct {
	Timestamp       time.Time `json:"time"`
	UnixTime        int       `json:"unixtime"`
	User            string    `json:"user"`
	Client          string    `json:"client"`
	QueryTime       float64   `json:"query_time"`
	LockTime        float64   `json:"lock_time"`
	RowsSent        int       `json:"rows_sent"`
	RowsExamined    int       `json:"rows_examined"`
	Query           string    `json:"query,omitempty"`
	NormalizedQuery string    `json:"normalized_query,omitempty"`
	DB              string    `json:"db,omitempty"`
	HostedOn        *string   `json:"hosted_on,omitempty"`
	ReadOnly        *bool     `json:"read_only,omitempty"`
	ReplicaLag      *int64    `json:"replica_lag,omitempty"`
	Role            *string   `json:"role,omitempty"`
	ClientIP        string    `json:"client_ip,omitempty"`
	skipQuery       bool
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)
	p.nower = &RealNower{}
	if p.conf.Host != "" {
		url := fmt.Sprintf("%s:%s@tcp(%s)/", p.conf.User, p.conf.Pass, p.conf.Host)
		db, err := sql.Open("mysql", url)
		if err != nil {
			logrus.Fatal(err)
		}

		// run one time queries
		hostedOn, err := getHostedOn(db)
		if err != nil {
			logrus.WithError(err).Warn("failed to get host env")
		}
		p.hostedOn = hostedOn

		role, err := getRole(db)
		if err != nil {
			logrus.WithError(err).Warn("failed to get role")
		}
		p.role = role

		// update hostedOn and readOnly every <n> seconds
		go func() {
			defer db.Close()
			ticker := time.NewTicker(time.Second * time.Duration(p.conf.QueryInterval))
			for _ = range ticker.C {
				readOnly, err := getReadOnly(db)
				if err != nil {
					logrus.WithError(err).Warn("failed to get read-only state")
				}
				p.readOnly = readOnly

				replicaLag, err := getReplicaLag(db)
				if err != nil {
					logrus.WithError(err).Warn("failed to get replica lag")
				}
				p.replicaLag = replicaLag
			}
		}()
	}
	return nil
}

func getReadOnly(db *sql.DB) (*bool, error) {
	rows, err := db.Query("SELECT @@global.read_only;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var value bool

	rows.Next()
	if err := rows.Scan(&value); err != nil {
		logrus.WithError(err).Warn("failed to get read-only state")
		return nil, err
	}
	if err := rows.Err(); err != nil {
		logrus.WithError(err).Warn("failed to get read-only state")
		return nil, err
	}
	return &value, nil
}

func getHostedOn(db *sql.DB) (*string, error) {
	rows, err := db.Query("SHOW GLOBAL VARIABLES WHERE Variable_name = 'basedir';")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var varName, value string

	for rows.Next() {
		if err := rows.Scan(&varName, &value); err != nil {
			return nil, err
		}
		if strings.HasPrefix(value, "/rdsdbbin/") {
			return &rdsStr, nil
		}
	}

	// TODO: implement ec2 detection

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func getReplicaLag(db *sql.DB) (*int64, error) {
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, err
	}

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	for i := range values {
		var v sql.RawBytes
		values[i] = &v
	}

	err = rows.Scan(values...)
	if err != nil {
		return nil, err
	}

	for i, name := range columns {
		bp := values[i].(*sql.RawBytes)
		vs := string(*bp)
		if name == "Seconds_Behind_Master" {
			vi, err := strconv.ParseInt(vs, 10, 64)
			if err != nil {
				return nil, err
			}
			return &vi, nil
		}
	}
	return nil, nil
}

func getRole(db *sql.DB) (*string, error) {
	var res string
	rows, err := db.Query("SHOW MASTER STATUS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		res = "secondary"
		return nil, err
	}
	res = "primary"
	return &res, nil
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event) {
	// start up a goroutine to handle grouped sets of lines
	rawEvents := make(chan rawEvent)
	var wg sync.WaitGroup
	p.wg = wg
	defer p.wg.Wait()
	p.wg.Add(1)
	go p.handleEvents(rawEvents, send)

	// flag to indicate when we've got a complete event to send
	var sendEvent bool
	groupedLines := make([]string, 0, 5)
	for line := range lines {
		if strings.HasPrefix(line, "# Time: ") {
			// we've started a new event. Send the previous one.
			sendEvent = true
		}
		if sendEvent {
			sendEvent = false
			rawEvents <- rawEvent{lines: groupedLines}
			groupedLines = make([]string, 0, 5)
		}
		groupedLines = append(groupedLines, line)
	}
	if len(groupedLines) != 0 {
		rawEvents <- rawEvent{lines: groupedLines}
	}
	logrus.Debug("lines channel is closed, ending mysql processor")
	close(rawEvents)
}

func (p *Parser) handleEvents(rawEvents <-chan rawEvent, send chan<- event.Event) {
	defer p.wg.Done()
	for rawE := range rawEvents {
		sq := p.handleEvent(rawE)
		sq.HostedOn = p.hostedOn
		sq.ReadOnly = p.readOnly
		sq.ReplicaLag = p.replicaLag
		sq.Role = p.role
		ev, err := p.processSlowQuery(sq)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":       err,
				"slowQuery": sq,
			}).Debug("skipping query")
			continue
		}
		send <- ev
	}
	logrus.Debug("done with mysql handleEvents")
}

// handle a single event
func (p *Parser) handleEvent(rawE rawEvent) SlowQuery {
	var err error
	sq := SlowQuery{}
	logrus.WithFields(logrus.Fields{
		"rawE": rawE,
	}).Debug("About to parse event")
	for _, line := range rawE.lines {
		// parse each line and populate the SlowQuery object
		switch {
		case reTime.MatchString(line):
			matchGroups := reTime.FindStringSubmatchMap(line)
			sq.Timestamp, err = time.Parse(timeFormat, matchGroups["time"])
			if err != nil {
				sq.Timestamp = p.nower.Now()
			}
		case reAdminPing.MatchString(line):
			// this evetn is an administrative ping and we should
			// ignore the entire event
			logrus.WithFields(logrus.Fields{
				"line":  line,
				"event": rawE,
			}).Debug("readmin ping detected; skipping this event")
			sq.skipQuery = true
		case reUser.MatchString(line):
			matchGroups := reUser.FindStringSubmatchMap(line)
			sq.User = strings.Split(matchGroups["user"], "[")[0]
			hostAndIP := strings.Split(matchGroups["host"], " ")
			sq.Client = hostAndIP[0]
			sq.ClientIP = hostAndIP[1][1 : len(hostAndIP[1])-1]
		case reQueryStats.MatchString(line):
			matchGroups := reQueryStats.FindStringSubmatchMap(line)
			sq.QueryTime, err = strconv.ParseFloat(matchGroups["queryTime"], 64)
			sq.LockTime, err = strconv.ParseFloat(matchGroups["lockTime"], 64)
			sq.RowsSent, err = strconv.Atoi(matchGroups["rowsSent"])
			sq.RowsExamined, err = strconv.Atoi(matchGroups["rowsExamined"])
		case reUse.FindString(line) != "":
			db := strings.TrimPrefix(line, reUse.FindString(line))
			db = strings.TrimRight(db, ";")
			db = strings.Trim(db, "`")
			sq.DB = db
			// Use this line as the query unless, if a real query follows it will be replaced.
			sq.Query = line
		case reSetTime.MatchString(line):
			matchGroups := reSetTime.FindStringSubmatchMap(line)
			sq.UnixTime, err = strconv.Atoi(matchGroups["unixTime"])
		case reQuery.MatchString(line):
			matchGroups := reQuery.FindStringSubmatchMap(line)
			sq.Query = matchGroups["query"]
			sq.NormalizedQuery = ""
		default:
			// unknown row; log and skip
			logrus.WithFields(logrus.Fields{
				"line": line,
			}).Debug("No regex match for line in the middle of a query. skipping")
		}
	}
	// We always need a timestamp; if we never found the line, use Now()
	if sq.Timestamp.IsZero() {
		sq.Timestamp = p.nower.Now()
	}
	return sq
}

// custom error to indicate empty query
type emptyQueryError struct {
	err string
}

func (e *emptyQueryError) Error() string {
	e.err = "skipped slow query"
	return e.err
}

func (p *Parser) processSlowQuery(sq SlowQuery) (event.Event, error) {
	// if we didn't match any lines at all, skip the query
	if sq == (SlowQuery{}) {
		sq.skipQuery = true
	}
	// OK, we've collected all the lines, send in the event
	if !sq.skipQuery {
		return event.Event{
			Timestamp: sq.Timestamp,
			Data:      sq.mapify(),
		}, nil
	}
	// we're skipping this query
	return event.Event{}, &emptyQueryError{}
}

func (s SlowQuery) mapify() map[string]interface{} {
	mapped := map[string]interface{}{
		"time":             s.Timestamp,
		"unixtime":         s.UnixTime,
		"user":             s.User,
		"client":           s.Client,
		"client_ip":        s.ClientIP,
		"query_time":       s.QueryTime,
		"lock_time":        s.LockTime,
		"rows_sent":        s.RowsSent,
		"rows_examined":    s.RowsExamined,
		"query":            s.Query,
		"normalized_query": s.NormalizedQuery,
	}
	if s.HostedOn != nil {
		mapped["hosted_on"] = *s.HostedOn
	}
	if s.ReadOnly != nil {
		mapped["read_only"] = *s.ReadOnly
	}
	if s.ReplicaLag != nil {
		mapped["replica_lag"] = *s.ReplicaLag
	}
	if s.Role != nil {
		mapped["role"] = *s.Role
	}
	return mapped
}
