// Package mysqlaudit consumes mysql audit logs
package mysqlaudit

import (
	"regexp"
	"strings"
	"sync"
	"encoding/json"
	"time"


	"github.com/Sirupsen/logrus"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
)

var (
	reTime = parsers.ExtRegexp{regexp.MustCompile("^[0-9]+_(?P<ts>[^ ]+)$")}
)

type Options struct {
	TimeFieldName     string `long:"timefield" description:"Name of the field that contains a timestamp"`
	TimeFieldFormat   string `long:"format" description:"Format of the timestamp found in timefield (supports strftime and Golang time formats)"`
	FilterRegex       string `long:"filter_regex" description:"a regular expression that will filter the input stream and only parse lines that match"`
	InvertFilter      bool   `long:"invert_filter" description:"change the filter_regex to only process lines that do *not* match"`
	SyslogIdent       string `long:"syslog_ident" description:"If the log is collected via syslog. Handy if you got multiple database servers sending log to a central syslog server"`

	NumParsers int `hidden:"true" description:"number of keyval parsers to spin up"`
}

type Parser struct {
	conf        Options
	lineParser  parsers.LineParser
	filterRegex *regexp.Regexp

	warnedAboutTime bool
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)
	if p.conf.FilterRegex != "" {
		var err error
		if p.filterRegex, err = regexp.Compile(p.conf.FilterRegex); err != nil {
			return err
		}
	}

	p.lineParser = &AuditLineParser{}
	return nil
}

type AuditLineParser struct {
}

func (j *AuditLineParser) ParseLine(line string) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})
	var err error
    var data map[string]interface{}

    if err := json.Unmarshal([]byte(line), &data); err != nil {
        logrus.Debug("boom")
    } else {
        parsed = data["audit_record"].(map[string]interface{})
    }

	return parsed, err
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	wg := sync.WaitGroup{}
	numParsers := 1
	if p.conf.NumParsers > 0 {
		numParsers = p.conf.NumParsers
	}
	for i := 0; i < numParsers; i++ {
		wg.Add(1)
		go func() {
			for line := range lines {
				line = strings.TrimSpace(line)
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Debug("Attempting to process keyval log line")

				// if matching regex is set, filter lines here
				if p.filterRegex != nil {
					matched := p.filterRegex.MatchString(line)
					// if both are true or both are false, skip. else continue
					if matched == p.conf.InvertFilter {
						logrus.WithFields(logrus.Fields{
							"line":    line,
							"matched": matched,
						}).Debug("skipping line due to FilterMatch.")
						continue
					}
				}

				// take care of any headers on the line
				var prefixFields map[string]string
				if prefixRegex != nil {
					var prefix string
					prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
					line = strings.TrimPrefix(line, prefix)
				}
				//take care of syslog headers if SyslogIdent is defined
				var hostname string
				if p.conf.SyslogIdent != "" {
					//regexp matching the first word of the line which is the hostname
					re := regexp.MustCompile(`^[^\s]*\s`)
					hostname  = re.FindString(line)
					if hostname != "" {
						//Strip away hostname from line
						line = strings.TrimPrefix(line, hostname)
                    				//strip away syslog identifier from line
                    				line = strings.TrimPrefix(line, p.conf.SyslogIdent)
					}
				}
				parsedLine, err := p.lineParser.ParseLine(line)
				//if a hostname is found it is applied to parsedLine
				if hostname != "" { 
                    			parsedLine["dbserver"] = hostname
               			}
				if err != nil {
					// skip lines that won't parse
					logrus.WithFields(logrus.Fields{
						"line":  line,
						"error": err,
					}).Debug("skipping line; failed to parse.")
					continue
				}
				if len(parsedLine) == 0 {
					// skip empty lines, as determined by the parser
					logrus.WithFields(logrus.Fields{
						"line":  line,
						"error": err,
					}).Debug("skipping line; no key/val pairs found.")
					continue
				}
				if allEmpty(parsedLine) {
					// skip events for which all fields are the empty string, because that's
					// probably broken
					logrus.WithFields(logrus.Fields{
						"line":  line,
						"error": err,
					}).Debug("skipping line; all values are the empty string.")
					continue
				}
				// merge the prefix fields and the parsed line contents
				for k, v := range prefixFields {
					parsedLine[k] = v
				}

				// look for the timestamp in any of the prefix fields or regular content
				//timestamp := httime.GetTimestamp(parsedLine, "timestamp", p.conf.TimeFieldFormat)
				timestamp := p.getTimestamp(parsedLine)

				delete(parsedLine, "timestamp")

				// send an event to Transmission
				e := event.Event{
					Timestamp: timestamp,
					Data:      parsedLine,
				}
				send <- e
			}
			wg.Done()
		}()
	}
	wg.Wait()
	logrus.Debug("lines channel is closed, ending keyval processor")
}

// allEmpty returns true if all values in the map are the empty string
// TODO move this into the main honeytail loop instead of the keyval parser
func allEmpty(pl map[string]interface{}) bool {
	for _, v := range pl {
		vStr, ok := v.(string)
		if !ok {
			// wouldn't coerce to string, so it must have something that's not an
			// empty string
			return false
		}
		if vStr != "" {
			return false
		}
	}
	// we've gone through the entire map and every field value has matched ""
	return true
}

// tries to extract a timestamp from the log line
func (n *Parser) getTimestamp(evMap map[string]interface{}) time.Time {

    /*_, mg := reTime.FindStringSubmatchMap(evMap["record"].(string))

    if mg == nil {
        return httime.Now()
    }

    logrus.Debug("xxx")
    logrus.Debug(mg["ts"])

	timestamp, _ := httime.Parse("2006-01-02T15:04:05", mg["ts"])*/

	timestamp, _ := httime.Parse("2006-01-02T15:04:05 MST", evMap["timestamp"].(string))

    logrus.Debug(timestamp.Format(time.RFC3339))

	return timestamp
}
