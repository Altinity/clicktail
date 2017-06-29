// Package nginx consumes nginx logs
package nginx

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/gonx"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
)

const (
	commonLogFormatTimeLayout = "02/Jan/2006:15:04:05 -0700"
	iso8601TimeLayout         = "2006-01-02T15:04:05-07:00"
)

type Options struct {
	ConfigFile      string `long:"conf" description:"Path to Nginx config file"`
	LogFormatName   string `long:"format" description:"Log format name to look for in the Nginx config file"`
	TimeFieldName   string `long:"timefield" description:"Name of the field that contains a timestamp"`
	TimeFieldFormat string `long:"time_format" description:"Timestamp format to use (strftime and Golang time.Parse supported)"`

	NumParsers int `hidden:"true" description:"number of nginx parsers to spin up"`
}

type Parser struct {
	conf       Options
	lineParser parsers.LineParser
}

func (n *Parser) Init(options interface{}) error {
	n.conf = *options.(*Options)

	// Verify we've got our config, find our format
	nginxConfig, err := os.Open(string(n.conf.ConfigFile))
	if err != nil {
		return err
	}
	defer nginxConfig.Close()
	// get the nginx log format from the config file
	// get a nginx log parser
	parser, err := gonx.NewNginxParser(nginxConfig, n.conf.LogFormatName)
	if err != nil {
		return err
	}
	gonxParser := &GonxLineParser{
		parser: parser,
	}
	n.lineParser = gonxParser
	return nil
}

type GonxLineParser struct {
	parser *gonx.Parser
}

func (g *GonxLineParser) ParseLine(line string) (map[string]interface{}, error) {
	gonxEvent, err := g.parser.ParseString(line)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"logline": line,
		}).Debug("failed to parse nginx log line")
		return nil, err
	}
	return typeifyParsedLine(gonxEvent.Fields), nil
}

func (n *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	// parse lines one by one
	wg := sync.WaitGroup{}
	for i := 0; i < n.conf.NumParsers; i++ {
		wg.Add(1)
		go func() {
			for line := range lines {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Debug("Attempting to process nginx log line")

				// take care of any headers on the line
				var prefixFields map[string]interface{}
				if prefixRegex != nil {
					var prefix string
					prefix, fields := prefixRegex.FindStringSubmatchMap(line)
					line = strings.TrimPrefix(line, prefix)
					prefixFields = typeifyParsedLine(fields)
				}

				parsedLine, err := n.lineParser.ParseLine(line)
				if err != nil {
					continue
				}
				// merge the prefix fields and the parsed line contents
				for k, v := range prefixFields {
					parsedLine[k] = v
				}
				timestamp := n.getTimestamp(parsedLine)

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
	logrus.Debug("lines channel is closed, ending nginx processor")
}

// typeifyParsedLine attempts to cast numbers in the event to floats or ints
func typeifyParsedLine(pl map[string]string) map[string]interface{} {
	// try to convert numbers, if possible
	msi := make(map[string]interface{}, len(pl))
	for k, v := range pl {
		switch {
		case strings.Contains(v, "."):
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				msi[k] = f
				continue
			}
		case v == "-":
			// no value, don't set a "-" string
			continue
		default:
			i, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				msi[k] = i
				continue
			}
		}
		msi[k] = v
	}
	return msi
}

// tries to extract a timestamp from the log line
func (n *Parser) getTimestamp(evMap map[string]interface{}) time.Time {
	var (
		setBothFieldsMsg = "Timestamp format and field must both be set to be used, one was not. Using current time instead."
	)

	// Custom (user-defined) timestamp field/format takes priority over the
	// default parsing behavior. Try that first.
	if n.conf.TimeFieldFormat != "" || n.conf.TimeFieldName != "" {
		if n.conf.TimeFieldFormat == "" || n.conf.TimeFieldName == "" {
			logrus.Debug(setBothFieldsMsg)
			return httime.Now()
		}
		return httime.GetTimestamp(evMap, n.conf.TimeFieldName, n.conf.TimeFieldFormat)
	}

	if _, ok := evMap["time_local"]; ok {
		return httime.GetTimestamp(evMap, "time_local", commonLogFormatTimeLayout)
	}

	if _, ok := evMap["time_iso8601"]; ok {
		return httime.GetTimestamp(evMap, "time_iso8601", iso8601TimeLayout)
	}

	return httime.GetTimestamp(evMap, "", "")
}
