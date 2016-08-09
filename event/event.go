// Package event contains the struct used to pass events between parsers and
// the libhoney module.
package event

import "time"

// Event is a single log event
type Event struct {
	// Timestamp is the time of the event (may be different from current time)
	Timestamp time.Time
	// Data is a map[string]interface{} containing key/value pairs for all the
	// metrics to submit in this event
	Data map[string]interface{}
}
