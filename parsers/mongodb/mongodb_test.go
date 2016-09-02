package mongodb

import (
	"reflect"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
)

type testLineMaps struct {
	line string
	ev   event.Event
}

func TestProcessLines(t *testing.T) {
	t1, _ := time.Parse(iso8601UTCTimeFormat, "2010-01-02T12:34:56.000Z")
	tlm := []testLineMaps{
		{
			line: "2010-01-02T12:34:56.000Z I CONTROL [conn123456789] git version fooooooo",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"severity":  "informational",
					"component": "CONTROL",
					"context":   "conn123456789",
					"message":   "git version fooooooo",
				},
			},
		},
	}
	m := &Parser{
		conf:            Options{},
		nower:           &FakeNower{},
		timeStampFormat: iso8601UTCTimeFormat,
		lineParser:      &MongoLineParser{},
	}
	lines := make(chan string)
	send := make(chan event.Event)
	// prep the incoming channel with test lines for the processor
	go func() {
		for _, pair := range tlm {
			lines <- pair.line
		}
		close(lines)
	}()
	// spin up the processor to process our test lines
	go m.ProcessLines(lines, send)
	for _, pair := range tlm {
		ev := <-send

		equals := true
		if len(ev.Data) != len(pair.ev.Data) {
			equals = false
		}
		if equals && ev.Timestamp != pair.ev.Timestamp {
			equals = false
		}
		for k := range ev.Data {
			if !reflect.DeepEqual(ev.Data[k], pair.ev.Data[k]) {
				equals = false
				break
			}
		}

		if !equals {
			t.Fatalf("line ev didn't match up for %s. Expected %+v, actual: %+v",
				// pair.line, string(pair.ev.Blob), string(resp.Blob))
				pair.line, pair.ev, ev)
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(iso8601UTCTimeFormat, "2010-10-02T12:34:56.000Z")
	return fakeTime
}
