// Package nginx consumes nginx logs
package nginx

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/gonx"
	flag "github.com/jessevdk/go-flags"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

const (
	commonLogFormatTimeLayout = "02/Jan/2006:15:04:05 -0700"
	iso8601TimeLayout         = "2006-01-02T15:04:05-07:00"
)

type Options struct {
	ConfigFile    flag.Filename `long:"conf" description:"Path to Nginx config file"`
	LogFormatName string        `long:"format" description:"Log format name to look for in the Nginx config file"`
}

type Parser struct {
	conf       Options
	lineParser LineParser
	nower      Nower
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
	n.nower = &RealNower{}
	return nil
}

type LineParser interface {
	ParseLine(line string) (map[string]string, error)
}

type GonxLineParser struct {
	parser *gonx.Parser
}

func (g *GonxLineParser) ParseLine(line string) (map[string]string, error) {
	gonxEvent, err := g.parser.ParseString(line)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"logline": line,
		}).Debug("failed to parse nginx log line")
		return nil, err
	}
	return gonxEvent.Fields, nil
}

func (n *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	// parse lines one by one
	for line := range lines {
		logrus.WithFields(logrus.Fields{
			"line": line,
		}).Debug("Attempting to process nginx log line")

		// take care of any headers on the line
		var prefixFields map[string]string
		if prefixRegex != nil {
			var prefix string
			prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
			line = strings.TrimPrefix(line, prefix)
		}

		parsedLine, err := n.lineParser.ParseLine(line)
		if err != nil {
			continue
		}
		// merge the prefix fields and the parsed line contents
		for k, v := range prefixFields {
			parsedLine[k] = v
		}
		// typedEvent, err := typeifyEvent(nginxEvent)
		typedEvent, err := typeifyParsedLine(parsedLine)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"line":  line,
				"event": parsedLine,
			}).Debug("failed to typeify event")
			continue
		}
		timestamp := getTimestamp(n.nower, typedEvent)

		e := event.Event{
			Timestamp: timestamp,
			Data:      typedEvent,
		}
		send <- e
	}
	logrus.Debug("lines channel is closed, ending nginx processor")
}

// typeifyParsedLine attempts to cast numbers in the event to floats or ints
func typeifyParsedLine(pl map[string]string) (map[string]interface{}, error) {
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
	return msi, nil
}

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (r *RealNower) Now() time.Time {
	return time.Now().UTC()
}

// tries to extract a timestamp from the log line
func getTimestamp(nower Nower, evMap map[string]interface{}) time.Time {
	var timestamp time.Time
	var err error
	defer delete(evMap, "time_local")
	defer delete(evMap, "time_iso8601")
	if val, ok := evMap["time_local"]; ok {
		rawTime, found := val.(string)
		if !found {
			// unable to parse string. log and return Now()
			logrus.WithFields(logrus.Fields{
				"expected_time": val,
			}).Debug("unable to coerce expected time to string")
			return nower.Now()
		}
		timestamp, err = time.Parse(commonLogFormatTimeLayout, rawTime)
		if err != nil {
			timestamp = nower.Now()
		}
	} else if val, ok := evMap["time_iso8601"]; ok {
		rawTime, found := val.(string)
		if !found {
			// unable to parse string. log and return Now()
			logrus.WithFields(logrus.Fields{
				"expected_time": val,
			}).Debug("unable to coerce expected time to string")
			return nower.Now()
		}
		timestamp, err = time.Parse(iso8601TimeLayout, rawTime)
		if err != nil {
			timestamp = nower.Now()
		}
	} else {
		timestamp = nower.Now()
	}
	return timestamp
}
