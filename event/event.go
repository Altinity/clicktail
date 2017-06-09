// Package event contains the struct used to pass events between parsers and
// the libhoney module.
package event

import "time"

// Event is a single log event
type Event struct {
	// Timestamp is the time of the event (may be different from current time)
	Timestamp time.Time
	// SampleRate is the rate at which this event is sampled. If it is positive,
	// the event should be sent with that sample rate.  If it is -1 the event
	// should be dropped instead of getting sent. Zero value should be treated as
	// unset and the event sent with a sample rate of 1
	SampleRate int
	// Data is a map[string]interface{} containing key/value pairs for all the
	// metrics to submit in this event
	Data map[string]interface{}
}
