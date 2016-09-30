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
	"github.com/honeycombio/mysqltools/query/normalizer"

	"github.com/honeycombio/honeytail/event"
)

// See mysql_test for example log entries
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

const (
	rdsStr  = "rds"
	ec2Str  = "ec2"
	selfStr = "self"

	// Event attributes
	userKey            = "user"
	clientKey          = "client"
	clientIPKey        = "client_ip"
	queryTimeKey       = "query_time"
	lockTimeKey        = "lock_time"
	rowsSentKey        = "rows_sent"
	rowsExaminedKey    = "rows_examined"
	databaseKey        = "database"
	queryKey           = "query"
	normalizedQueryKey = "normalized_query"
	statementKey       = "statement"
	tablesKey          = "tables"
	// Event attributes that apply to the host as a whole
	hostedOnKey   = "hosted_on"
	readOnlyKey   = "read_only"
	replicaLagKey = "replica_lag"
	roleKey       = "role"
)

var (
	reTime       = myRegexp{regexp.MustCompile("^# Time: (?P<time>[^ ]+)Z *$")}
	reAdminPing  = myRegexp{regexp.MustCompile("^# administrator command: Ping; *$")}
	reUser       = myRegexp{regexp.MustCompile("^# User@Host: (?P<user>[^#]+) @ (?P<host>[^#]+).*$")}
	reQueryStats = myRegexp{regexp.MustCompile("^# Query_time: (?P<queryTime>[0-9.]+) *Lock_time: (?P<lockTime>[0-9.]+) *Rows_sent: (?P<rowsSent>[0-9]+) *Rows_examined: (?P<rowsExamined>[0-9]+) *$")}
	reSetTime    = myRegexp{regexp.MustCompile("^SET timestamp=(?P<unixTime>[0-9]+);$")}
	reQuery      = myRegexp{regexp.MustCompile("^(?P<query>[^#]*).*$")}
	reUse        = myRegexp{regexp.MustCompile("^(?i)use ")}
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
	hostedOn   string
	readOnly   *bool
	replicaLag *int64
	role       *string
	normalizer *normalizer.Parser
}

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (n *RealNower) Now() time.Time {
	return time.Now().UTC()
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)
	p.nower = &RealNower{}
	p.normalizer = &normalizer.Parser{}
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

func getHostedOn(db *sql.DB) (string, error) {
	rows, err := db.Query("SHOW GLOBAL VARIABLES WHERE Variable_name = 'basedir';")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var varName, value string

	for rows.Next() {
		if err := rows.Scan(&varName, &value); err != nil {
			return "", err
		}
		if strings.HasPrefix(value, "/rdsdbbin/") {
			return rdsStr, nil
		}
	}

	// TODO: implement ec2 detection

	if err := rows.Err(); err != nil {
		return "", err
	}
	return "", nil
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
	rawEvents := make(chan []string)
	var wg sync.WaitGroup
	p.wg = wg
	defer p.wg.Wait()
	p.wg.Add(1)
	go p.handleEvents(rawEvents, send)

	// flag to indicate when we've got a complete event to send
	var sendEvent bool
	groupedLines := make([]string, 0, 5)
	for line := range lines {
		if strings.HasPrefix(line, "# ") && len(groupedLines) > 0 {
			// we've started a new event. Send the previous one.
			sendEvent = true
		}
		if sendEvent {
			sendEvent = false
			rawEvents <- groupedLines
			groupedLines = make([]string, 0, 5)
		}
		groupedLines = append(groupedLines, line)
	}
	if len(groupedLines) != 0 {
		rawEvents <- groupedLines
	}
	logrus.Debug("lines channel is closed, ending mysql processor")
	close(rawEvents)
}

func (p *Parser) handleEvents(rawEvents <-chan []string, send chan<- event.Event) {
	defer p.wg.Done()
	for rawE := range rawEvents {
		sq, timestamp := p.handleEvent(rawE)
		if len(sq) == 0 {
			continue
		}
		if p.hostedOn != "" {
			sq[hostedOnKey] = p.hostedOn
		}
		if p.readOnly != nil {
			sq[readOnlyKey] = *p.readOnly
		}
		if p.replicaLag != nil {
			sq[replicaLagKey] = *p.replicaLag
		}
		if p.role != nil {
			sq[roleKey] = *p.role
		}
		send <- event.Event{
			Timestamp: timestamp,
			Data:      sq,
		}
	}
	logrus.Debug("done with mysql handleEvents")
}

// Parse a set of MySQL log lines that seem to represent a single event and
// return a struct of extracted data as well as the highest-resolution timestamp
// available.
func (p *Parser) handleEvent(rawE []string) (map[string]interface{}, time.Time) {
	sq := map[string]interface{}{}
	var timeFromComment time.Time
	var timeFromSet int64
	query := ""
	for _, line := range rawE {
		// parse each line and populate the map of attributes
		switch {
		case reTime.MatchString(line):
			matchGroups := reTime.FindStringSubmatchMap(line)
			timeFromComment, _ = time.Parse(timeFormat, matchGroups["time"])
		case reAdminPing.MatchString(line):
			// this event is an administrative ping and we should
			// ignore the entire event
			logrus.WithFields(logrus.Fields{
				"line":  line,
				"event": rawE,
			}).Debug("readmin ping detected; skipping this event")
			return nil, time.Time{}
		case reUser.MatchString(line):
			query = ""
			matchGroups := reUser.FindStringSubmatchMap(line)
			sq[userKey] = strings.Split(matchGroups["user"], "[")[0]
			hostAndIP := strings.Split(matchGroups["host"], " ")
			sq[clientKey] = hostAndIP[0]
			sq[clientIPKey] = hostAndIP[1][1 : len(hostAndIP[1])-1]
		case reQueryStats.MatchString(line):
			query = ""
			matchGroups := reQueryStats.FindStringSubmatchMap(line)
			if queryTime, err := strconv.ParseFloat(matchGroups["queryTime"], 64); err == nil {
				sq[queryTimeKey] = queryTime
			}
			if lockTime, err := strconv.ParseFloat(matchGroups["lockTime"], 64); err == nil {
				sq[lockTimeKey] = lockTime
			}
			if rowsSent, err := strconv.Atoi(matchGroups["rowsSent"]); err == nil {
				sq[rowsSentKey] = rowsSent
			}
			if rowsExamined, err := strconv.Atoi(matchGroups["rowsExamined"]); err == nil {
				sq[rowsExaminedKey] = rowsExamined
			}
		case reUse.FindString(line) != "":
			query = ""
			db := strings.TrimPrefix(line, reUse.FindString(line))
			db = strings.TrimRight(db, ";")
			db = strings.Trim(db, "`")
			sq[databaseKey] = db
			// Use this line as the query/normalized_query unless, if a real query follows it will be replaced.
			sq[queryKey] = strings.TrimRight(line, ";")
			sq[normalizedQueryKey] = sq[queryKey]
		case reSetTime.MatchString(line):
			query = ""
			matchGroups := reSetTime.FindStringSubmatchMap(line)
			timeFromSet, _ = strconv.ParseInt(matchGroups["unixTime"], 10, 64)
		case reQuery.MatchString(line):
			matchGroups := reQuery.FindStringSubmatchMap(line)
			query = query + " " + matchGroups["query"]
			if strings.HasSuffix(query, ";") {
				q := strings.TrimSpace(strings.TrimSuffix(query, ";"))
				sq[queryKey] = q
				sq[normalizedQueryKey] = p.normalizer.NormalizeQuery(q)
				if len(p.normalizer.LastTables) > 0 {
					sq[tablesKey] = strings.Join(p.normalizer.LastTables, " ")
				}
				sq[statementKey] = p.normalizer.LastStatement
				query = ""
			}

		default:
			// unknown row; log and skip
			logrus.WithFields(logrus.Fields{
				"line": line,
			}).Debug("No regex match for line in the middle of a query. skipping")
		}
	}

	// We always need a timestamp.
	//
	// timeFromComment may include millisecond resolution but doesn't include
	//   time zone.
	// timeFromSet is a UNIX timestamp and thus more reliable, but also (thus)
	//   doesn't contain millisecond resolution.
	//
	// In the best case (we have both), we combine the two; in the worst case (we
	//   have neither) we fall back to "now."
	combinedTime := p.nower.Now()
	if !timeFromComment.IsZero() && timeFromSet > 0 {
		nanos := time.Duration(timeFromComment.Nanosecond())
		combinedTime = time.Unix(timeFromSet, 0).Add(nanos)
	} else if !timeFromComment.IsZero() {
		combinedTime = timeFromComment // cross our fingers that UTC is ok
	} else if timeFromSet > 0 {
		combinedTime = time.Unix(timeFromSet, 0)
	}

	return sq, combinedTime
}

// custom error to indicate empty query
type emptyQueryError struct {
	err string
}

func (e *emptyQueryError) Error() string {
	e.err = "skipped slow query"
	return e.err
}
