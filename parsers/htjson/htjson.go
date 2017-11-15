// Package htjson (honeytail-json, renamed to not conflict with the json module)
// parses logs that are one json blob per line.
package htjson

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
)

type Options struct {
	TimeFieldName   string `long:"timefield" description:"Name of the field that contains a timestamp"`
	TimeFieldFormat string `long:"format" description:"Format of the timestamp found in timefield (supports strftime and Golang time formats)"`

	NumParsers int `hidden:"true" description:"number of htjson parsers to spin up"`
}

type Parser struct {
	conf       Options
	lineParser parsers.LineParser

	warnedAboutTime bool
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)

	p.lineParser = &JSONLineParser{}
	return nil
}

type JSONLineParser struct {
}

// ParseLine will unmarshal the thing it read in to detect errors in the JSON
// (by failing to parse) and give us an object that can be mutated by the
// various filters honeytail might apply.
func (j *JSONLineParser) ParseLine(line string) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})
	err := json.Unmarshal([]byte(line), &parsed)
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
				}).Debug("Attempting to process json log line")

				// take care of any headers on the line
				var prefixFields map[string]string
				if prefixRegex != nil {
					var prefix string
					prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
					line = strings.TrimPrefix(line, prefix)
				}

				parsedLine, err := p.lineParser.ParseLine(line)

				if err != nil {
					// skip lines that won't parse
					logrus.WithFields(logrus.Fields{
						"line": line,
					}).Debug("skipping line; failed to parse.")
					continue
				}
				timestamp := httime.GetTimestamp(parsedLine, p.conf.TimeFieldName, p.conf.TimeFieldFormat)

				// merge the prefix fields and the parsed line contents
				for k, v := range prefixFields {
					parsedLine[k] = v
				}

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
	logrus.Debug("lines channel is closed, ending json processor")
}
