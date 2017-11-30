// Package postgresql contains code for parsing PostgreSQL slow query logs.
//
// The Postgres slow query format
// ------------------------------
//
// In general, Postgres logs consist of a prefix, a level, and a message:
//
// 2017-11-06 19:20:32 UTC [11534-2] LOG:  autovacuum launcher shutting down
// |<-------------prefix----------->|level|<----------message-------------->|
//
// 2017-11-07 01:43:39 UTC [3542-7] postgres@test LOG:  duration: 15.577 ms  statement: SELECT * FROM test;
// |<-------------prefix------------------------>|level|<-------------------message---------------------->|
//
// The format of the configuration prefix is configurable as `log_line_prefix` in postgresql.conf
// using the following format specifiers:
//
//   %a = application name
//   %u = user name
//   %d = database name
//   %r = remote host and port
//   %h = remote host
//   %p = process ID
//   %t = timestamp without milliseconds
//   %m = timestamp with milliseconds
//   %i = command tag
//   %e = SQL state
//   %c = session ID
//   %l = session line number
//   %s = session start timestamp
//   %v = virtual transaction ID
//   %x = transaction ID (0 if none)
//   %q = stop here in non-session
//        processes
//   %% = '%'
//
// For example, the prefix format for the lines above is:
// %t [%p-%l] %q%u@%d
// We currently require users to pass the prefix format as a parser option.
//
// Slow query logs specifically have the following format:
// 2017-11-07 01:43:39 UTC [3542-7] postgres@test LOG:  duration: 15.577 ms  statement: SELECT * FROM test;
// |<-------------prefix------------------------>|<----------header-------------------->|<-----query----->|
//
// For convenience, we call everything after the prefix but before the actual query string the "header".
//
// The query may span multiple lines; continuations are indented. For example:
//
// 2017-11-07 01:43:39 UTC [3542-7] postgres@test LOG:  duration: 15.577 ms  statement: SELECT * FROM test
//		WHERE id=1;
package postgresql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/mysqltools/query/normalizer"
)

const (
	// Regex string that matches timestamps in log
	timestampRe   = `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}[.0-9]* [A-Z]+`
	defaultPrefix = "%t [%p-%l] %u@%d"
	// Regex string that matches the header of a slow query log line
	slowQueryHeader = `\s*(?P<level>[A-Z0-9]+):\s+duration: (?P<duration>[0-9\.]+) ms\s+statement: `
)

var slowQueryHeaderRegex = &parsers.ExtRegexp{regexp.MustCompile(slowQueryHeader)}

// prefixField represents a specific format specifier in the log_line_prefix string
// (see module comment for details).
type prefixField struct {
	Name    string
	Pattern string
}

func (p *prefixField) ReString() string {
	return fmt.Sprintf("(?P<%s>%s)", p.Name, p.Pattern)
}

var prefixValues = map[string]prefixField{
	"%a": prefixField{Name: "application", Pattern: "\\S+"},
	"%u": prefixField{Name: "user", Pattern: "\\S+"},
	"%d": prefixField{Name: "database", Pattern: "\\S+"},
	"%r": prefixField{Name: "host_port", Pattern: "\\S+"},
	"%h": prefixField{Name: "host", Pattern: "\\S+"},
	"%p": prefixField{Name: "pid", Pattern: "\\d+"},
	"%t": prefixField{Name: "timestamp", Pattern: timestampRe},
	"%m": prefixField{Name: "timestamp_millis", Pattern: timestampRe},
	"%n": prefixField{Name: "timestamp_unix", Pattern: "\\d+"},
	"%i": prefixField{Name: "command_tag", Pattern: "\\S+"},
	"%e": prefixField{Name: "sql_state", Pattern: "\\S+"},
	"%c": prefixField{Name: "session_id", Pattern: "\\d+"},
	"%l": prefixField{Name: "session_line_number", Pattern: "\\d+"},
	"%s": prefixField{Name: "session_start", Pattern: timestampRe},
	"%v": prefixField{Name: "virtual_transaction_id", Pattern: "\\S+"},
	"%x": prefixField{Name: "transaction_id", Pattern: "\\S+"},
}

type Options struct {
	LogLinePrefix string `long:"log_line_prefix" description:"Format string for PostgreSQL log line prefix"`
}

type Parser struct {
	// regex to match the log_line_prefix format specified by the user
	pgPrefixRegex *parsers.ExtRegexp
}

func (p *Parser) Init(options interface{}) (err error) {
	var logLinePrefixFormat string
	conf, ok := options.(*Options)
	if !ok || conf.LogLinePrefix == "" {
		logLinePrefixFormat = defaultPrefix
	} else {
		logLinePrefixFormat = conf.LogLinePrefix
	}
	p.pgPrefixRegex, err = buildPrefixRegexp(logLinePrefixFormat)
	return err
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	rawEvents := make(chan []string)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go p.handleEvents(rawEvents, send, wg)
	var groupedLines []string
	for line := range lines {
		if prefixRegex != nil {
			// This is the "global" prefix regex as specified by the
			// --log_prefix option, for stripping prefixes added by syslog or
			// the like. It's unlikely that it'll actually be set by consumers
			// of database logs. Don't confuse this with p.pgPrefixRegex, which
			// is a compiled regex for parsing the postgres-specific line
			// prefix.
			var prefix string
			prefix = prefixRegex.FindString(line)
			line = strings.TrimPrefix(line, prefix)
		}
		if !isContinuationLine(line) && len(groupedLines) > 0 {
			// If the line we just parsed is the start of a new log statement,
			// send off the previously accumulated group.
			rawEvents <- groupedLines
			groupedLines = make([]string, 0, 1)
		}
		groupedLines = append(groupedLines, line)
	}

	rawEvents <- groupedLines
	close(rawEvents)
	wg.Wait()
}

// handleEvents receives sets of grouped log lines, each representing a single
// log statement. It attempts to parse them, and sends the events it constructs
// down the send channel.
func (p *Parser) handleEvents(rawEvents <-chan []string, send chan<- event.Event, wg *sync.WaitGroup) {
	defer wg.Done()
	// TODO: spin up a group of goroutines to do this
	for rawEvent := range rawEvents {
		ev := p.handleEvent(rawEvent)
		if ev != nil {
			send <- *ev
		}
	}
}

// handleEvent takes a single grouped log statement (an array of lines) and attempts to parse it.
// It returns a pointer to an Event if successful, and nil if not.
func (p *Parser) handleEvent(rawEvent []string) *event.Event {
	normalizer := normalizer.Parser{}
	if len(rawEvent) == 0 {
		return nil
	}
	firstLine := rawEvent[0]

	// First, try to parse the prefix
	match, suffix, generalMeta := parsePrefix(p.pgPrefixRegex, firstLine)
	if !match {
		// Note: this may be noisy when debug logging is turned on, since the
		// postgres general log contains lots of other statements as well.
		logrus.WithField("line", firstLine).Debug("Log line prefix didn't match expected format")
		return nil
	}

	ev := &event.Event{
		Data: make(map[string]interface{}, 0),
	}

	addFieldsToEvent(generalMeta, ev)

	// Now, parse the slow query header
	match, query, slowQueryMeta := parsePrefix(slowQueryHeaderRegex, suffix)

	if !match {
		logrus.WithField("line", firstLine).Debug("didn't find slow query header, skipping line")
		return nil
	}

	if rawDuration, ok := slowQueryMeta["duration"]; ok {
		duration, _ := strconv.ParseFloat(rawDuration, 64)
		ev.Data["duration"] = duration
	} else {
		logrus.WithField("query", query).Debug("Failed to find query duration in log line")
	}

	// Finally, concatenate the remaining text to form the query, and attempt to
	// normalize it.
	for _, line := range rawEvent[1:] {
		query += " " + strings.TrimLeft(line, " \t")
	}
	normalizedQuery := normalizer.NormalizeQuery(query)

	ev.Data["query"] = query
	ev.Data["normalized_query"] = normalizedQuery
	if len(normalizer.LastTables) > 0 {
		ev.Data["tables"] = strings.Join(normalizer.LastTables, " ")
	}
	if len(normalizer.LastComments) > 0 {
		ev.Data["comments"] = "/* " + strings.Join(normalizer.LastComments, " */ /* ") + " */"
	}

	return ev
}

func isContinuationLine(line string) bool {
	return strings.HasPrefix(line, "\t")
}

// addFieldsToEvent takes a map of key-value metadata extracted from a log
// line, and adds them to the given event. It'll convert values to integer
// types where possible, and try to populate the event's timestamp.
func addFieldsToEvent(fields map[string]string, ev *event.Event) {
	for k, v := range fields {
		// Try to convert values to integer types where sensible, and extract
		// timestamp for event
		switch k {
		case "session_id", "pid", "session_line_number":
			if typed, err := strconv.Atoi(v); err == nil {
				ev.Data[k] = typed
			} else {
				ev.Data[k] = v
			}
		case "timestamp", "timestamp_millis":
			if timestamp, err := time.Parse("2006-01-02 15:04:05.999 MST", v); err == nil {
				ev.Timestamp = timestamp
			} else {
				logrus.WithField("timestamp", v).WithError(err).Debug("Error parsing query timestamp")
			}
		case "timestamp_unix":
			if typed, err := strconv.Atoi(v); err == nil {
				// Convert millisecond-resolution Unix timestamp to time.Time
				// object
				timestamp := time.Unix(int64(typed/1000), int64((1000*1000)*(typed%1000))).UTC()
				ev.Timestamp = timestamp
			} else {
				logrus.WithField("timestamp", v).WithError(err).Debug("Error parsing query timestamp")
			}
		default:
			ev.Data[k] = v
		}
	}
}

func parsePrefix(re *parsers.ExtRegexp, line string) (matched bool, suffix string, fields map[string]string) {
	prefix, fields := re.FindStringSubmatchMap(line)
	if prefix == "" {
		return false, "", nil
	}
	return true, line[len(prefix):], fields
}

func buildPrefixRegexp(prefixFormat string) (*parsers.ExtRegexp, error) {
	prefixFormat = strings.Replace(prefixFormat, "%%", "%", -1)
	// The %q format specifier means "if this log line isn't part of a session,
	// stop here." The slow query logs that we care about always come from
	// sessions, so ignore this.
	prefixFormat = strings.Replace(prefixFormat, "%q", "", -1)
	prefixFormat = regexp.QuoteMeta(prefixFormat)
	for k, v := range prefixValues {
		prefixFormat = strings.Replace(prefixFormat, k, v.ReString(), -1)
	}

	re, err := regexp.Compile(prefixFormat)
	if err != nil {
		return nil, err
	}
	return &parsers.ExtRegexp{re}, nil
}
