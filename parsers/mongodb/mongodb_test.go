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
	time_string1 := "2010-01-02T12:34:56.000Z"
	t1, _ := time.Parse(iso8601UTCTimeFormat, time_string1)

	locks_string1 := "locks:{ Global: { acquireCount: { r: 2, w: 2 } }, Database: { acquireCount: { w: 2 } }, Collection: { acquireCount: { w: 1 } }, oplog: { acquireCount: { w: 1 } } }"
	locks1 := map[string]interface{}{
		"Global": map[string]interface{}{
			"acquireCount": map[string]interface{}{
				"r": float64(2),
				"w": float64(2),
			},
		},
		"Database": map[string]interface{}{
			"acquireCount": map[string]interface{}{
				"w": float64(2),
			},
		},
		"Collection": map[string]interface{}{
			"acquireCount": map[string]interface{}{
				"w": float64(1),
			},
		},
		"oplog": map[string]interface{}{
			"acquireCount": map[string]interface{}{
				"w": float64(1),
			},
		},
	}
	tlm := []testLineMaps{
		{
			line: time_string1 + " I CONTROL [conn123456789] git version fooooooo",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"severity":      "informational",
					"component":     "CONTROL",
					"context":       "conn123456789",
					"message":       "git version fooooooo",
					"read_or_write": "",
				},
			},
		},

		{
			line: time_string1 + " I QUERY    [context12345] query database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"severity":              "informational",
					"component":             "QUERY",
					"context":               "context12345",
					"operation":             "query",
					"namespace":             "database.collection:Stuff",
					"database":              "database",         // decomposed from namespace
					"collection":            "collection:Stuff", // decomposed from namespace
					"query":                 map[string]interface{}{},
					"normalized_query":      "{  }",
					"locks":                 locks1,
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
					"read_or_write":         "read",
				},
			},
		},

		{
			line: time_string1 + " I WRITE    [context12345] insert database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"severity":              "informational",
					"component":             "WRITE",
					"context":               "context12345",
					"operation":             "insert",
					"namespace":             "database.collection:Stuff",
					"database":              "database",         // decomposed from namespace
					"collection":            "collection:Stuff", // decomposed from namespace
					"query":                 map[string]interface{}{},
					"normalized_query":      "{  }",
					"locks":                 locks1,
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
					"read_or_write":         "write",
				},
			},
		},

		{
			line: time_string1 + " I COMMAND    [context12345] command database.$cmd command: insert { } " + locks_string1 + " 0ms",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"severity":              "informational",
					"component":             "COMMAND",
					"context":               "context12345",
					"operation":             "command",
					"namespace":             "database.$cmd",
					"database":              "database", // decomposed from namespace
					"collection":            "$cmd",     // decomposed from namespace
					"command_type":          "insert",
					"command":               map[string]interface{}{},
					"locks":                 locks1,
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
					"read_or_write":         "write",
				},
			},
		},
	}
	m := &Parser{
		conf:       Options{},
		nower:      &FakeNower{},
		lineParser: &MongoLineParser{},
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
