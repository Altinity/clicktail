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
	UBUNTU_3_2_9_FIND   = `2016-09-15T00:01:55.387+0000 I COMMAND [conn93] command protecteddb.comedy command: find { find: "comedy", filter: { $where: "this.year > 2000" } } planSummary: COLLSCAN keysExamined:0 docsExamined:5 cursorExhausted:1 keyUpdates:0 writeConflicts:0 numYields:0 nreturned:2 reslen:245 locks:{ Global: { acquireCount: { r: 7, w: 1 } }, Database: { acquireCount: { r: 3, w: 1 } }, Collection: { acquireCount: { r: 3, w: 1 } }, Metadata: { acquireCount: { W: 1 } } } protocol:op_command 29ms`
	UBUNTU_3_2_9_UPDATE = `2016-09-14T23:36:36.793+0000 I WRITE [conn61] update protecteddb.comedy query: { name: "Hulk" } update: { $unset: { cast: 1.0 } } keysExamined:0 docsExamined:4 nMatched:0 nModified:0 keyUpdates:0 writeConflicts:0 numYields:0 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } 0ms`
	UBUNTU_2_6_FIND     = `2016-09-15T02:38:10.395-0400 [conn1579035] query starfruit_production.users query: { $query: { altemails: { $in: [ "REDACTED@domain.org" ] } }, $orderby: { _id: 1 } } planSummary: IXSCAN { _id: 1 } ntoskip:0 nscanned:67439 nscannedObjects:67439 keyUpdates:0 numYields:1 locks(micros) r:114782 nreturned:0 reslen:20 105ms`
	UBUNTU_2_4_FIND     = `Tue Sep 13 21:10:33.961 [TTLMonitor] query btest.system.indexes query: { expireAfterSeconds: { $exists: true } } ntoreturn:0 ntoskip:0 nscanned:1 keyUpdates:0 locks(micros) r:60 nreturned:0 reslen:20 0ms`
	OSX_3_2_9_AGGREGATE = `2016-09-14T14:46:13.879-0700 I COMMAND [conn1] command testtest.zips command: aggregate { aggregate: "zips", pipeline: [ { $group: { _id: "$state", totalPop: { $sum: "$pop" } } }, { $match: { totalPop: { $gte: 10000000.0 } } } ], cursor: {} } keyUpdates:0 writeConflicts:0 numYields:229 reslen:342 locks:{ Global: { acquireCount: { r: 466 } }, Database: { acquireCount: { r: 233 } }, Collection: { acquireCount: { r: 233 } } } protocol:op_command 34ms`
	HEARTBEAT           = `Sun Sep 18 07:20:03.246 [conn123456789] command admin.$cmd command: replSetHeartbeat { replSetHeartbeat: "replica-set-here", from: "host:port" } ntoreturn:1 keyUpdates:0 numYields:0  reslen:100 0ms`
	NESTED_QUOTES       = `2016-09-20T14:55:06.189-0400 [conn2915444] update namespace.collection query: { _id: ObjectId('51abe5b6c') } update: { $set: { recent_data: [ { id: ObjectId('57e1860'), msg: "Nietzsche said "For what constitutes the tremendous historical uniqueness of that Persian is just the opposite of this."" } ] } } nscanned:1 nscannedObjects:1 nMatched:1 nModified:1 keyUpdates:0 numYields:0 locks(micros) w:393 0ms`
	T1_STRING           = "2010-01-02T12:34:56.000Z"
)

var (
	T1                          = time.Date(2010, time.January, 2, 12, 34, 56, 0, time.UTC)
	UBUNTU_3_2_9_INSERT_TIME, _ = time.Parse(iso8601LocalTimeFormat, "2016-09-14T23:39:23.450+0000")
	UBUNTU_3_2_9_FIND_TIME, _   = time.Parse(iso8601LocalTimeFormat, "2016-09-15T00:01:55.387+0000")
	UBUNTU_3_2_9_UPDATE_TIME, _ = time.Parse(iso8601LocalTimeFormat, "2016-09-14T23:36:36.793+0000")
	HEARTBEAT_TIME, _           = time.Parse(ctimeTimeFormat, "Sun Sep 18 07:20:03.246")
	UBUNTU_2_4_FIND_TIME, _     = time.Parse(ctimeTimeFormat, "Tue Sep 13 21:10:33.961")
	UBUNTU_2_6_FIND_TIME, _     = time.Parse(iso8601LocalTimeFormat, "2016-09-15T02:38:10.395-0400")
	OSX_3_2_9_AGGREGATE_TIME, _ = time.Parse(iso8601LocalTimeFormat, "2016-09-14T14:46:13.879-0700")
	NESTED_QUOTES_TIME, _       = time.Parse(iso8601LocalTimeFormat, "2016-09-20T14:55:06.189-0400")
)

type processed struct {
	time        time.Time
	includeData map[string]interface{}
	excludeKeys []string
}

func TestProcessLines(t *testing.T) {
	nower := &FakeNower{}
	locks_string1 := "locks:{ Global: { acquireCount: { r: 2, w: 2 } }, Database: { acquireCount: { w: 2 } }, Collection: { acquireCount: { w: 1 } }, oplog: { acquireCount: { w: 1 } } }"
	tlm := []struct {
		line     string
		expected processed
	}{
		{
			line: CONTROL,
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"severity":  "informational",
					"component": "CONTROL",
					"context":   "conn123456789",
					"message":   "git version fooooooo",
				},
			},
		},

		{
			line: T1_STRING + " I QUERY    [context12345] query database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"component":             "QUERY",
					"context":               "context12345",
					"operation":             "query",
					"namespace":             "database.collection:Stuff",
					"database":              "database",         // decomposed from namespace
					"collection":            "collection:Stuff", // decomposed from namespace
					"query":                 map[string]interface{}{},
					"normalized_query":      "{  }",
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
				},
			},
		},

		{
			line: T1_STRING + " I WRITE    [context12345] insert database.collection:Stuff query: {} " + locks_string1 + " 0ms",
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"component":             "WRITE",
					"context":               "context12345",
					"operation":             "insert",
					"namespace":             "database.collection:Stuff",
					"database":              "database",         // decomposed from namespace
					"collection":            "collection:Stuff", // decomposed from namespace
					"query":                 map[string]interface{}{},
					"normalized_query":      "{  }",
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
				},
				excludeKeys: []string{
					"database_read_lock",
				},
			},
		},

		{
			line: T1_STRING + " I COMMAND    [context12345] command database.$cmd command: insert { } " + locks_string1 + " 0ms",
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"component":             "COMMAND",
					"context":               "context12345",
					"operation":             "command",
					"namespace":             "database.$cmd",
					"database":              "database", // decomposed from namespace
					"collection":            "$cmd",     // decomposed from namespace
					"command_type":          "insert",
					"command":               map[string]interface{}{},
					"global_read_lock":      float64(2),
					"global_write_lock":     float64(2),
					"database_write_lock":   float64(2),
					"collection_write_lock": float64(1),
					"oplog_write_lock":      float64(1),
					"duration_ms":           float64(0),
				},
				excludeKeys: []string{
					"database_read_lock",
				},
			},
		},

		{
			line: UBUNTU_3_2_9_INSERT,
			expected: processed{
				time: UBUNTU_3_2_9_INSERT_TIME,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"operation":             "command",
					"command_type":          "insert",
					"namespace":             "protecteddb.comedy",
					"keyUpdates":            0.0,
					"reslen":                25.0,
					"global_read_lock":      1.0,
					"global_write_lock":     1.0,
					"database_write_lock":   1.0,
					"collection_write_lock": 1.0,
				},
				excludeKeys: []string{
					"oplog_write_lock",
					"collection_read_lock",
				},
			},
		},

		{
			line: UBUNTU_3_2_9_FIND,
			expected: processed{
				time: UBUNTU_3_2_9_FIND_TIME,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"operation":             "command",
					"command_type":          "find",
					"namespace":             "protecteddb.comedy",
					"keyUpdates":            0.0,
					"reslen":                245.0,
					"global_read_lock":      7.0,
					"global_write_lock":     1.0,
					"database_read_lock":    3.0,
					"database_write_lock":   1.0,
					"collection_read_lock":  3.0,
					"collection_write_lock": 1.0,
					"docsExamined":          5.0,
					"duration_ms":           29.0,
				},
				excludeKeys: []string{
					"oplog_write_lock",
				},
			},
		},

		{
			line: UBUNTU_3_2_9_UPDATE,
			expected: processed{
				time: UBUNTU_3_2_9_UPDATE_TIME,
				includeData: map[string]interface{}{
					"severity":              "informational",
					"operation":             "update",
					"namespace":             "protecteddb.comedy",
					"keyUpdates":            0.0,
					"global_read_lock":      1.0,
					"global_write_lock":     1.0,
					"database_write_lock":   1.0,
					"collection_write_lock": 1.0,
					"docsExamined":          4.0,
				},
				excludeKeys: []string{
					"database_read_lock",
					"oplog_write_lock",
				},
			},
		},

		{
			line: UBUNTU_2_6_FIND,
			expected: processed{
				time: UBUNTU_2_6_FIND_TIME,
				includeData: map[string]interface{}{
					"operation":         "query",
					"context":           "conn1579035",
					"namespace":         "starfruit_production.users",
					"read_lock_held_us": int64(114782),
					"duration_ms":       105.0,
					"reslen":            20.0,
					"nscanned":          67439.0,
					"database":          "starfruit_production",
					"collection":        "users",
				},
			},
		},

		{
			line: HEARTBEAT,
			expected: processed{
				time: HEARTBEAT_TIME.AddDate(nower.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"command_type": "replSetHeartbeat",
					"replica_set":  "replica-set-here",
				},
			},
		},

		{
			line: UBUNTU_2_4_FIND,
			expected: processed{
				time: UBUNTU_2_4_FIND_TIME.AddDate(nower.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"operation":         "query",
					"read_lock_held_us": int64(60),
					"replica_set":       "replica-set-here",
				},
			},
		},

		{
			line: OSX_3_2_9_AGGREGATE,
			expected: processed{
				time: OSX_3_2_9_AGGREGATE_TIME,
				includeData: map[string]interface{}{
					"duration_ms": 34.0,
					"replica_set": "replica-set-here",
				},
				excludeKeys: []string{},
			},
		},

		{
			line: NESTED_QUOTES,
			expected: processed{
				time: NESTED_QUOTES_TIME,
				includeData: map[string]interface{}{
					"duration_ms": 0.0,
				},
				excludeKeys: []string{},
			},
		},
	}
	m := &Parser{
		conf:       Options{},
		nower:      nower,
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

		if ev.Timestamp.UnixNano() != pair.expected.time.UnixNano() {
			t.Errorf("Parsed timestamp didn't match up for %s.\n  Expected: %+v\n  Actual: %+v",
				pair.line, pair.expected.time, ev.Timestamp)
		}

		var missing []string
		for k := range pair.expected.includeData {
			if _, ok := ev.Data[k]; !ok {
				missing = append(missing, k)
			} else if !reflect.DeepEqual(ev.Data[k], pair.expected.includeData[k]) {
				t.Errorf("  Parsed data value %s didn't match up for %s.\n  Expected: %+v\n  Actual: %+v",
					k, pair.line, pair.expected.includeData[k], ev.Data[k])
			}
		}
		if missing != nil {
			t.Errorf("  Parsed data was missing keys for line: %s\n  Missing: %+v\n  Parsed data: %+v",
				pair.line, missing, ev.Data)
		}
		for _, k := range pair.expected.excludeKeys {
			if _, ok := ev.Data[k]; ok {
				t.Errorf("  Parsed data included unexpected key %s for line: %s", k, pair.line)
			}
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(iso8601UTCTimeFormat, "2010-10-02T12:34:56.000Z")
	return fakeTime
}
