package mongodb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/httime/httimetest"
)

const (
	CONTROL                = "2010-01-02T12:34:56.000Z I CONTROL [conn123456789] git version fooooooo"
	UBUNTU_3_2_9_INSERT    = `2016-09-14T23:39:23.450+0000 I COMMAND  [conn68] command protecteddb.comedy command: insert { insert: "comedy", documents: [ { _id: ObjectId('57d9dfab9fc9998ce4e0c072'), name: "Bill & Ted's Excellent Adventure", year: 1989.0, extrakey: "benisawesome" } ], ordered: true } ninserted:1 keyUpdates:0 writeConflicts:0 numYields:0 reslen:25 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } protocol:op_command 0ms`
	UBUNTU_3_2_9_FIND      = `2016-09-15T00:01:55.387+0000 I COMMAND [conn93] command protecteddb.comedy command: find { find: "comedy", filter: { $where: "this.year > 2000" } } planSummary: COLLSCAN keysExamined:0 docsExamined:5 cursorExhausted:1 keyUpdates:0 writeConflicts:0 numYields:0 nreturned:2 reslen:245 locks:{ Global: { acquireCount: { r: 7, w: 1 } }, Database: { acquireCount: { r: 3, w: 1 } }, Collection: { acquireCount: { r: 3, w: 1 } }, Metadata: { acquireCount: { W: 1 } } } protocol:op_command 29ms`
	UBUNTU_3_2_9_UPDATE    = `2016-09-14T23:36:36.793+0000 I WRITE [conn61] update protecteddb.comedy query: { name: "Hulk" } update: { $unset: { cast: 1.0 } } keysExamined:0 docsExamined:4 nMatched:0 nModified:0 keyUpdates:0 writeConflicts:0 numYields:0 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } 0ms`
	UBUNTU_2_6_FIND        = `2016-09-15T02:38:10.395-0400 [conn1579035] query starfruit_production.users query: { $query: { altemails: { $in: [ "REDACTED@domain.org" ] } }, $orderby: { _id: 1 } } planSummary: IXSCAN { _id: 1 } ntoskip:0 nscanned:67439 nscannedObjects:67439 keyUpdates:0 numYields:1 locks(micros) r:114782 nreturned:0 reslen:20 105ms`
	UBUNTU_2_4_FIND        = `Tue Sep 13 21:10:33.961 [TTLMonitor] query btest.system.indexes query: { expireAfterSeconds: { $exists: true } } ntoreturn:0 ntoskip:0 nscanned:1 keyUpdates:0 locks(micros) r:60 nreturned:0 reslen:20 0ms`
	OSX_3_2_9_AGGREGATE    = `2016-09-14T14:46:13.879-0700 I COMMAND [conn1] command testtest.zips command: aggregate { aggregate: "zips", pipeline: [ { $group: { _id: "$state", totalPop: { $sum: "$pop" } } }, { $match: { totalPop: { $gte: 10000000.0 } } } ], cursor: {} } keyUpdates:0 writeConflicts:0 numYields:229 reslen:342 locks:{ Global: { acquireCount: { r: 466 } }, Database: { acquireCount: { r: 233 } }, Collection: { acquireCount: { r: 233 } } } protocol:op_command 34ms`
	HEARTBEAT              = `Sun Sep 18 07:20:03.246 [conn123456789] command admin.$cmd command: replSetHeartbeat { replSetHeartbeat: "replica-set-here", from: "host:port" } ntoreturn:1 keyUpdates:0 numYields:0  reslen:100 0ms`
	NESTED_QUOTES          = `2016-09-20T14:55:06.189-0400 [conn2915444] update namespace.collection query: { _id: ObjectId('51abe5b6c') } update: { $set: { recent_data: [ { id: ObjectId('57e1860'), msg: "Nietzsche said "For what constitutes the tremendous historical uniqueness of that Persian is just the opposite of this."" } ] } } nscanned:1 nscannedObjects:1 nMatched:1 nModified:1 keyUpdates:0 numYields:0 locks(micros) w:393 0ms`
	UBUNTU_3_0_KILLCURSORS = `2016-09-20T14:55:06.189-0400 I QUERY [conn924267662] killcursors keyUpdates:0 writeConflicts:0 numYields:0 locks:{ Global: { acquireCount: { r: 2 } }, Database: { acquireCount: { r: 1 } }, Collection: { acquireCount: { r: 1 } } } user_key_comparison_count:0 block_cache_hit_count:0 block_read_count:0 block_read_byte:0 internal_key_skipped_count:0 internal_delete_skipped_count:0 get_from_memtable_count:0 seek_on_memtable_count:0 seek_child_seek_count:0 0ms`
	T1_STRING              = "2010-01-02T12:34:56.000Z"
	MONGO_3_4_QUERY        = `2016-10-20T22:27:54.580+0000 I COMMAND [Balancer] command config.locks command: findAndModify { findAndModify: "locks", query: { _id: "balancer", state: 0 }, update: { $set: { ts: ObjectId('580944e96a82726bb4a8427f'), state: 2, who: "ConfigServer:Balancer", process: "ConfigServer", when: new Date(1477002473519), why: "CSRS Balancer" } }, upsert: true, new: true, writeConcern: { w: "majority", wtimeout: 15000 } } planSummary: IXSCAN { _id: 1 } update: { $set: { ts: ObjectId('580944e96a82726bb4a8427f'), state: 2, who: "ConfigServer:Balancer", process: "ConfigServer", when: new Date(1477002473519), why: "CSRS Balancer" } } keysExamined:0 docsExamined:0 nMatched:0 nModified:0 upsert:1 keysInserted:3 numYields:0 reslen:338 locks:{ Global: { acquireCount: { r: 2, w: 2 } }, Database: { acquireCount: { w: 2 }, acquireWaitCount: { w: 1 }, timeAcquiringMicros: { w: 9385 } }, Collection: { acquireCount: { w: 1 } }, Metadata: { acquireCount: { w: 1 }, acquireWaitCount: { w: 1 }, timeAcquiringMicros: { w: 18 } }, oplog: { acquireCount: { w: 1 } } } protocol:op_query 1061ms`
	MONGO_3_4_GETMORE      = `2016-10-20T22:28:01.785+0000 I COMMAND  [conn8] command TestDB.TestColl appName: "MongoDB Shell" command: getMore { getMore: 18862806827, collection: "TestColl" } originatingCommand: { find: "TestColl", projection: { _id: 0.0, Counter: 1.0 }, shardVersion: [ Timestamp 1000|0, ObjectId('580944efeaec999c2d8e0d5b') ] } planSummary: COLLSCAN cursorid:18862806827 keysExamined:0 docsExamined:59899 cursorExhausted:1 numYields:468 nreturned:59899 reslen:1726100 locks:{ Global: { acquireCount: { r: 938 } }, Database: { acquireCount: { r: 469 } }, Collection: { acquireCount: { r: 469 } } } protocol:op_command 120ms`
	MONGO_3_4_INIT         = `2016-10-20T21:13:03.294+0000 I CONTROL [initandlisten] options: { net: { port: 23511 }, nopreallocj: true, replication: { oplogSizeMB: 40, replSet: "test-configRS" }, setParameter: { enableTestCommands: "1", logComponentVerbosity: "{tracking:1}", numInitialSyncAttempts: "1", numInitialSyncConnectAttempts: "60", writePeriodicNoops: "false" }, sharding: { clusterRole: "configsvr" }, storage: { dbPath: "/data/db/job14/mongorunner/test-configRS-0", engine: "wiredTiger", journal: { enabled: true }, mmapv1: { preallocDataFiles: false, smallFiles: true }, wiredTiger: { engineConfig: { cacheSizeGB: 1.0 } } }}`
	MONGO_3_4_INDEX        = `2016-10-20T22:27:59.508+0000 I INDEX    [conn5] build index on: TestDB.TestColl properties: { v: 2, key: { Counter: 1.0 }, name: "Counter_1", ns: "TestDB.TestColl" }`
	MONGO_3_4_SHARDING     = `2016-10-20T22:27:59.516+0000 I SHARDING [conn1] about to log metadata event into changelog: { _id: "ip-10-69-189-27-2016-10-20T22:27:59.516+0000-580944efeaec999c2d8e0d58", server: "ip-10-69-189-27", clientAddr: "127.0.0.1:35756", time: new Date(1477002479516), what: "shardCollection.start", ns: "TestDB.TestColl", details: { shardKey: { Counter: 1.0 }, collection: "TestDB.TestColl", primary: "shard0000:ip-10-69-189-27:23760", initShards: [], numChunks: 1 } }`
	UPDATE_SIMPLE_COMMAND  = `Tue Sep 13 21:10:33.961 I COMMAND  [conn11896572] command data.$cmd command: update { update: "currentMood", updates: [ { q: { mood: "bright" }, u: { $set: { mood: "dark" } } } ], writeConcern: { getLastError: 1, w: 1 }, ordered: true } keyUpdates:0 writeConflicts:0 numYields:0 reslen:95 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } user_key_comparison_count:466 block_cache_hit_count:10 block_read_count:0 block_read_byte:0 internal_key_skipped_count:17 internal_delete_skipped_count:0 get_from_memtable_count:0 seek_on_memtable_count:2 seek_child_seek_count:12 0ms`
	UPDATE_COMMAND         = `Tue Sep 13 21:10:33.961 I COMMAND  [conn11896572] command data.$cmd command: update { update: "avengers", updates: [ { q: { hulkForm: "Bruce Banner" }, u: { $set: { hulkForm: "Big Green" }, $setOnInsert: { hulkForm: "Big Green" } }, upsert: true } ], writeConcern: { getLastError: 1, w: 1 }, ordered: true } keyUpdates:0 writeConflicts:0 numYields:0 reslen:95 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } user_key_comparison_count:466 block_cache_hit_count:10 block_read_count:0 block_read_byte:0 internal_key_skipped_count:17 internal_delete_skipped_count:0 get_from_memtable_count:0 seek_on_memtable_count:2 seek_child_seek_count:12 0ms`
	DELETE_SIMPLE_COMMAND  = `Tue Sep 13 21:10:33.961 I COMMAND  [conn11974626] command appdata387.$cmd command: delete { delete: "currentMood", deletes: [ { q: { mood: "bright" } } ], writeConcern: { getLastError: 1, w: 1 } } keyUpdates:0 writeConflicts:0 numYields:0 reslen:80 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } user_key_comparison_count:392 block_cache_hit_count:10 block_read_count:0 block_read_byte:0 internal_key_skipped_count:0 internal_delete_skipped_count:0 get_from_memtable_count:0 seek_on_memtable_count:2 seek_child_seek_count:12 0ms`
	DELETE_COMMAND         = `Tue Sep 13 21:10:33.961 I COMMAND  [conn11974626] command appdata387.$cmd command: delete { delete: "avengerMembers", deletes: [ { q: { hulkForm: "Big Green", issue: { $ne: 4 } }, limit: 1 } ], writeConcern: { getLastError: 1, w: 1 } } keyUpdates:0 writeConflicts:0 numYields:0 reslen:80 locks:{ Global: { acquireCount: { r: 1, w: 1 } }, Database: { acquireCount: { w: 1 } }, Collection: { acquireCount: { w: 1 } } } user_key_comparison_count:392 block_cache_hit_count:10 block_read_count:0 block_read_byte:0 internal_key_skipped_count:0 internal_delete_skipped_count:0 get_from_memtable_count:0 seek_on_memtable_count:2 seek_child_seek_count:12 0ms`
)

var (
	T1                             = time.Date(2010, time.January, 2, 12, 34, 56, 0, time.UTC)
	UBUNTU_3_2_9_INSERT_TIME, _    = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-14T23:39:23.450+0000", time.UTC)
	UBUNTU_3_2_9_FIND_TIME, _      = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-15T00:01:55.387+0000", time.UTC)
	UBUNTU_3_2_9_UPDATE_TIME, _    = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-14T23:36:36.793+0000", time.UTC)
	HEARTBEAT_TIME, _              = time.ParseInLocation(ctimeTimeFormat, "Sun Sep 18 07:20:03.246", time.UTC)
	UBUNTU_2_4_FIND_TIME, _        = time.ParseInLocation(ctimeTimeFormat, "Tue Sep 13 21:10:33.961", time.UTC)
	UBUNTU_2_6_FIND_TIME, _        = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-15T02:38:10.395-0400", time.UTC)
	OSX_3_2_9_AGGREGATE_TIME, _    = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-14T14:46:13.879-0700", time.UTC)
	NESTED_QUOTES_TIME, _          = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-20T14:55:06.189-0400", time.UTC)
	UBUNTU_3_0_KILLCURSORS_TIME, _ = time.ParseInLocation(iso8601LocalTimeFormat, "2016-09-20T14:55:06.189-0400", time.UTC)
	MONGO_3_4_QUERY_TIME, _        = time.ParseInLocation(iso8601LocalTimeFormat, "2016-10-20T22:27:54.580+0000", time.UTC)
	MONGO_3_4_GETMORE_TIME, _      = time.ParseInLocation(iso8601LocalTimeFormat, "2016-10-20T22:28:01.785+0000", time.UTC)
	MONGO_3_4_INIT_TIME, _         = time.ParseInLocation(iso8601LocalTimeFormat, "2016-10-20T21:13:03.294+0000", time.UTC)
	MONGO_3_4_INDEX_TIME, _        = time.ParseInLocation(iso8601LocalTimeFormat, "2016-10-20T22:27:59.508+0000", time.UTC)
	MONGO_3_4_SHARDING_TIME, _     = time.ParseInLocation(iso8601LocalTimeFormat, "2016-10-20T22:27:59.516+0000", time.UTC)
	UPDATE_COMMAND_TIME, _         = time.ParseInLocation(ctimeTimeFormat, "Tue Sep 13 21:10:33.961", time.UTC)
	DELETE_COMMAND_TIME, _         = time.ParseInLocation(ctimeTimeFormat, "Tue Sep 13 21:10:33.961", time.UTC)
)

func init() {
	fakeNow, _ := time.Parse(iso8601UTCTimeFormat, "2010-10-02T12:34:56.000Z")
	httime.DefaultNower = &httimetest.FakeNower{fakeNow}
}

type processed struct {
	time        time.Time
	includeData map[string]interface{}
	excludeKeys []string
}

func TestProcessLines(t *testing.T) {
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
				time: HEARTBEAT_TIME.AddDate(httime.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"command_type": "replSetHeartbeat",
					"replica_set":  "replica-set-here",
				},
			},
		},

		{
			line: UBUNTU_2_4_FIND,
			expected: processed{
				time: UBUNTU_2_4_FIND_TIME.AddDate(httime.Now().Year(), 0, 0),
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
		{
			line: UBUNTU_3_0_KILLCURSORS,
			expected: processed{
				time: UBUNTU_3_0_KILLCURSORS_TIME,
				includeData: map[string]interface{}{
					"duration_ms": 0.0,
					"operation":   "killcursors",
					"component":   "QUERY",
				},
				excludeKeys: []string{},
			},
		},
		{
			line: MONGO_3_4_QUERY,
			expected: processed{
				time: MONGO_3_4_QUERY_TIME,
				includeData: map[string]interface{}{
					"duration_ms":      1061.0,
					"component":        "COMMAND",
					"command_type":     "findAndModify",
					"keysInserted":     3.0,
					"normalized_query": `{ "_id": 1, "state": 1 }`,
				},
				excludeKeys: []string{},
			},
		},
		{
			line: MONGO_3_4_GETMORE,
			expected: processed{
				time: MONGO_3_4_GETMORE_TIME,
				includeData: map[string]interface{}{
					"duration_ms":  120.0,
					"component":    "COMMAND",
					"command_type": "getMore",
					"collection":   "TestColl",
					"docsExamined": 59899.0,
				},
				excludeKeys: []string{},
			},
		},
		{
			line: MONGO_3_4_INIT,
			expected: processed{
				time: MONGO_3_4_INIT_TIME,
				includeData: map[string]interface{}{
					"component": "CONTROL",
				},
				excludeKeys: []string{},
			},
		},
		{
			line: MONGO_3_4_INDEX,
			expected: processed{
				time: MONGO_3_4_INDEX_TIME,
				includeData: map[string]interface{}{
					"component": "INDEX",
					"context":   "conn5",
				},
				excludeKeys: []string{},
			},
		},
		{
			line: MONGO_3_4_SHARDING,
			expected: processed{
				time: MONGO_3_4_SHARDING_TIME,
				includeData: map[string]interface{}{
					"component":           "SHARDING",
					"context":             "conn1",
					"namespace":           "TestDB.TestColl",
					"database":            "TestDB",
					"collection":          "TestColl",
					"sharding_collection": "changelog",
					"changelog_primary":   "shard0000:ip-10-69-189-27:23760",
					"changelog_what":      "shardCollection.start",
					"changelog_changeid":  "ip-10-69-189-27-2016-10-20T22:27:59.516+0000-580944efeaec999c2d8e0d58",
				},
				excludeKeys: []string{},
			},
		},
		{
			line: UPDATE_SIMPLE_COMMAND,
			expected: processed{
				time: UPDATE_COMMAND_TIME.AddDate(httime.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"normalized_query": `{ "updates": [ { "$query": { "mood": 1 }, "$update": { "$set": { "mood": 1 } } } ] }`,
				},
				excludeKeys: []string{},
			},
		},
		{
			line: UPDATE_COMMAND,
			expected: processed{
				time: UPDATE_COMMAND_TIME.AddDate(httime.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"normalized_query": `{ "updates": [ { "$query": { "hulkForm": 1 }, "$update": { "$set": { "hulkForm": 1 }, "$setOnInsert": { "hulkForm": 1 } } } ] }`,
				},
				excludeKeys: []string{},
			},
		},
		{
			line: DELETE_SIMPLE_COMMAND,
			expected: processed{
				time: DELETE_COMMAND_TIME.AddDate(httime.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"normalized_query": `{ "deletes": [ { "$query": { "mood": 1 } } ] }`,
				},
				excludeKeys: []string{},
			},
		},
		{
			line: DELETE_COMMAND,
			expected: processed{
				time: DELETE_COMMAND_TIME.AddDate(httime.Now().Year(), 0, 0),
				includeData: map[string]interface{}{
					"normalized_query": `{ "deletes": [ { "$limit": 1, "$query": { "hulkForm": 1, "issue": { "$ne": 1 } } } ] }`,
				},
				excludeKeys: []string{},
			},
		},
	}
	m := &Parser{
		conf: Options{
			NumParsers: 1,
		},
	}
	lines := make(chan string, len(tlm))
	send := make(chan event.Event, len(tlm))
	// prep the incoming channel with test lines for the processor
	go func() {
		for _, pair := range tlm {
			lines <- pair.line
		}
		close(lines)
	}()
	// spin up the processor to process our test lines
	m.ProcessLines(lines, send, nil)
	// We expect that it got sent through in order, serialized thanks to NumParsers = 1
	for _, pair := range tlm {
		ev := <-send
		assert.Equal(t, pair.expected.time.UnixNano(), ev.Timestamp.UnixNano())

		var missing []string
		for k := range pair.expected.includeData {
			if _, ok := ev.Data[k]; !ok {
				missing = append(missing, k)
			} else {
				assert.Equal(t, pair.expected.includeData[k], ev.Data[k])
			}
		}
		assert.Nil(t, missing)
		for _, k := range pair.expected.excludeKeys {
			_, ok := ev.Data[k]
			assert.False(t, ok)
		}
	}
	close(send)
}
