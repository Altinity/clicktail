// Package mongodb is a parser for mongodb logs
package mongodb

import (
	"math/rand"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeytail/event"
	"github.com/tmc/mongologtools/parser"
)

type Options struct {
}

type Parser struct {
	conf       Options
	lineParser LineParser
	nower      Nower
}

type LineParser interface {
	ParseLogLine(line string) (map[string]interface{}, error)
}

type MongoLineParser struct {
}

func (m *MongoLineParser) ParseLogLine(line string) (map[string]interface{}, error) {
	return parser.ParseLogLine(line)
}

func (p *Parser) Init(_ interface{}) error {
	p.nower = &RealNower{}
	p.lineParser = &MongoLineParser{}
	return nil
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event) {
	for line := range lines {
		values, err := p.lineParser.ParseLogLine(line)
		// we get a bunch of errors from the parser on mongo logs, skip em
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"line":   line,
				"values": values,
			}).Debug("Successfully parsed line")
			// for each entry, make a json blob with key/value pairs for each value map
			e := event.Event{
				Timestamp: randomTime(p.nower),
				Data:      values,
			}
			send <- e
		} else {
			logrus.WithFields(logrus.Fields{
				"line": line,
			}).Debug("logline didn't parse, skipping.")
		}
	}
	logrus.Debug("lines channel is closed, ending mongo processor")
}

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (r *RealNower) Now() time.Time {
	return time.Now().UTC()
}

func randomTime(n Nower) time.Time {
	return n.Now().Add(-1 * time.Duration(rand.Int63n(604800)) * time.Second)
}
