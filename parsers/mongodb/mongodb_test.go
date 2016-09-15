package mongodb

import (
	"reflect"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
)

const (
	CONTROL             = "2010-01-02T12:34:56.000Z I CONTROL [conn123456789] git version fooooooo"
	UBUNTU_3_2_9_INSERT = `2016-09-14T23:39:23.450+0000 I COMMAND  [conn68] command protecteddb.comedy command: insert { insert: "comedy", documents: [ { _id: ObjectId('57d9dfab9fc9998ce4e0c072'), name: "Bill & Ted's Excellent Adventure", year: 1989.0, extrakey: "benisawesome" } ], ordered: true } ninserted:1 keyUpdates:0 writeConflicts:0 numYields:0 reslen:25 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } protocol:op_command 0ms`
)

var (
	T1                          = time.Date(2010, time.January, 2, 12, 34, 56, 0, time.UTC)
	UBUNTU_3_2_9_INSERT_TIME, _ = time.Parse(iso8601LocalTimeFormat, "2016-09-14T23:39:23.450+0000")
)

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
	tlm := []struct {
		line         string
		expectedTime time.Time
		expectedData map[string]interface{} // Intentionally not comprehensive
	}{
		{
			line:         CONTROL,
			expectedTime: T1,
			expectedData: map[string]interface{}{
				"severity":  "informational",
				"component": "CONTROL",
				"context":   "conn123456789",
				"message":   "git version fooooooo",
			},
		},

		{
			line:         time_string1 + " I QUERY    [context12345] query database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			expectedTime: t1,
			expectedData: map[string]interface{}{
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
				"read_or_write":         "read",
				"global_read_lock":      float64(2),
				"global_write_lock":     float64(2),
				"database_write_lock":   float64(2),
				"collection_write_lock": float64(1),
				"oplog_write_lock":      float64(1),
				"duration_ms":           float64(0),
			},
		},

		{
			line:         time_string1 + " I WRITE    [context12345] insert database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			expectedTime: t1,
			expectedData: map[string]interface{}{
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

		{
			line:         time_string1 + " I COMMAND    [context12345] command database.$cmd command: insert { } " + locks_string1 + " 0ms",
			expectedTime: t1,
			expectedData: map[string]interface{}{
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

		{
			line:         UBUNTU_3_2_9_INSERT,
			expectedTime: UBUNTU_3_2_9_INSERT_TIME,
			expectedData: map[string]interface{}{
				"read_or_write": "write",
				"namespace":     "protecteddb.comedy",
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

		if ev.Timestamp.UnixNano() != pair.expectedTime.UnixNano() {
			t.Errorf("Parsed timestamp didn't match up for %s.\n  Expected: %+v\n  Actual: %+v",
				pair.line, pair.expectedTime, ev.Timestamp)
		}

		var missing []string
		for k := range pair.expectedData {
			if _, ok := ev.Data[k]; !ok {
				missing = append(missing, k)
			} else if !reflect.DeepEqual(ev.Data[k], pair.expectedData[k]) {
				t.Errorf("  Parsed data value %s didn't match up for %s.\n  Expected: %+v\n  Actual: %+v",
					k, pair.line, pair.expectedData[k], ev.Data[k])
			}
		}
		if missing != nil {
			t.Errorf("  Parsed data was missing keys for line: %s\n  Missing: %+v", pair.line, missing)
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(iso8601UTCTimeFormat, "2010-10-02T12:34:56.000Z")
	return fakeTime
}
