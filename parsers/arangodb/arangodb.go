// Package arangodb is a parser for ArangoDB logs
package arangodb

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
)

const numParsers = 20

const (
	iso8601UTCTimeFormat   = "2006-01-02T15:04:05Z"
	iso8601LocalTimeFormat = "2006-01-02T15:04:05"

	timestampFieldName  = "timestamp"
	pidFieldName        = "pid"
	logLevelFieldName   = "logLevel"
	logTopicFieldName   = "logTopic"
	idFieldName         = "id"
	sourceIPFieldName   = "sourceIP"
	methodFieldName     = "method"
	protocolFieldName   = "protocol"
	resCodeFieldName    = "responseCode"
	reqBodyLenFieldName = "reqBodyLen"
	resBodyLenFieldName = "resBodyLen"
	fullURLFieldName    = "fullURL"
	totalTimeFieldName  = "totalTime"
)

var timestampFormats = []string{
	iso8601UTCTimeFormat,
	iso8601LocalTimeFormat,
}

// Options type for line parser, so far there are none.
type Options struct {
}

// Parser for log lines.
type Parser struct {
	conf       Options
	lineParser parsers.LineParser
}

// ArangoLineParser is a LineParser for ArangoDB log files.
type ArangoLineParser struct {
}

func firstWord(line *string) (word string, abort bool) {
	var pos = strings.IndexByte(*line, ' ')
	if pos < 0 {
		return "", true
	}
	word = (*line)[:pos]
	*line = (*line)[pos+1:]
	abort = false
	return
}

func removeBrackets(word string) string {
	var l = len(word)
	if l < 2 {
		return word
	}
	if word[0] == '(' && word[l-1] == ')' {
		return word[1 : l-1]
	}
	if word[0] == '[' && word[l-1] == ']' {
		return word[1 : l-1]
	}
	if word[0] == '{' && word[l-1] == '}' {
		return word[1 : l-1]
	}
	return word
}

func removeQuotes(word string) string {
	if len(word) == 0 {
		return word
	}
	if word[0] == '"' {
		word = word[1:]
	}
	if len(word) > 0 && word[len(word)-1] == '"' {
		word = word[:len(word)-1]
	}
	return word
}

// ParseLine method for an ArangoLineParser implementing LineParser.
func (m *ArangoLineParser) ParseLine(line string) (_ map[string]interface{}, err error) {
	// Do the actual work here, we look for log lines in the log topic "requests",
	// there are two types, one is a DEBUG line (could be switched off) containing
	// the request body, the other is the INFO line marking the end of the
	// request.
	var v = make(map[string]interface{})
	err = errors.New("Line is not a request log line.")
	var abort bool
	var s string

	v[timestampFieldName], abort = firstWord(&line)
	if abort {
		return
	}

	s, abort = firstWord(&line)
	if abort {
		return
	}
	v[pidFieldName] = removeBrackets(s)

	v[logLevelFieldName], abort = firstWord(&line)
	if abort {
		return
	}

	s, abort = firstWord(&line)
	if abort {
		return
	}
	v[logTopicFieldName] = s

	if s != "{requests}" {
		return
	}

	var fields = strings.Split(line, ",")
	if v[logLevelFieldName] == "DEBUG" {
		if len(fields) != 6 {
			return
		}
		v[idFieldName] = removeQuotes(fields[1])
		v[sourceIPFieldName] = removeQuotes(fields[2])
		v[methodFieldName] = removeQuotes(fields[3])
		v[protocolFieldName] = removeQuotes(fields[4])
		v[fullURLFieldName] = removeQuotes(fields[5])
	} else {
		if len(fields) != 10 {
			return
		}
		v[idFieldName] = removeQuotes(fields[1])
		v[sourceIPFieldName] = removeQuotes(fields[2])
		v[methodFieldName] = removeQuotes(fields[3])
		v[protocolFieldName] = removeQuotes(fields[4])
		v[resCodeFieldName], _ = strconv.ParseInt(fields[5], 10, 32)
		v[reqBodyLenFieldName], _ = strconv.ParseInt(fields[6], 10, 64)
		v[resBodyLenFieldName], _ = strconv.ParseInt(fields[7], 10, 64)
		v[fullURLFieldName] = removeQuotes(fields[8])
		v[totalTimeFieldName], _ = strconv.ParseFloat(fields[9], 64)
	}
	return v, nil
}

// Init method for parser object.
func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)
	p.lineParser = &ArangoLineParser{}
	return nil
}

// ProcessLines method for Parser.
func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	wg := sync.WaitGroup{}
	for i := 0; i < numParsers; i++ {
		wg.Add(1)
		go func() {
			for line := range lines {
				line = strings.TrimSpace(line)
				// take care of any headers on the line
				var prefixFields map[string]string
				if prefixRegex != nil {
					var prefix string
					prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
					line = strings.TrimPrefix(line, prefix)
				}

				values, err := p.lineParser.ParseLine(line)
				// we get a bunch of errors from the parser on ArangoDB logs, skip em
				if err == nil {
					timestamp, err := p.parseTimestamp(values)
					if err != nil {
						logSkipped(line, "couldn't parse logline timestamp, skipping")
						continue
					}

					// merge the prefix fields and the parsed line contents
					for k, v := range prefixFields {
						values[k] = v
					}

					logrus.WithFields(logrus.Fields{
						"line":   line,
						"values": values,
					}).Debug("Successfully parsed line")

					// we'll be putting the timestamp in the Event
					// itself, no need to also have it in the Data
					delete(values, timestampFieldName)

					send <- event.Event{
						Timestamp: timestamp,
						Data:      values,
					}
				} else {
					logSkipped(line, "logline didn't parse, skipping.")
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	logrus.Debug("lines channel is closed, ending arangodb processor")
}

func (p *Parser) parseTimestamp(values map[string]interface{}) (time.Time, error) {
	timestampValue, ok := values[timestampFieldName].(string)
	if ok {
		var err error
		for _, f := range timestampFormats {
			var timestamp time.Time
			timestamp, err = httime.Parse(f, timestampValue)
			if err == nil {
				return timestamp, nil
			}
		}
		return time.Time{}, err
	}

	return time.Time{}, errors.New("timestamp missing from logline")
}

func logSkipped(line string, msg string) {
	logrus.WithFields(logrus.Fields{"line": line}).Debugln(msg)
}
