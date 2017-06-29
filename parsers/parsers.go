// Package parsers provides an interface for different log parsing engines.
//
// Each module in here takes care of a specific log type, providing
// any necessary or relevant smarts for that style of logs.
package parsers

import "github.com/honeycombio/honeytail/event"

type Parser interface {
	// Init does any initialization necessary for the module
	Init(options interface{}) error
	// ProcessLines consumes log lines from the lines channel and sends log events
	// to the send channel. prefixRegex, if not nil, will be stripped from the
	// line prior to parsing. Any named groups will be added to the event.
	ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *ExtRegexp)
}

type LineParser interface {
	ParseLine(line string) (map[string]interface{}, error)
}
