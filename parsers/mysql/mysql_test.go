package mysql

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/mysqltools/query/normalizer"
)

type slowQueryData struct {
	rawE      []string
	sq        map[string]interface{}
	timestamp time.Time
}

const (
	hostLine  = "# User@Host: root[root] @ localhost [127.0.0.1]  Id:   136"
	timerLine = "# Query_time: 0.000171  Lock_time: 0.000000 Rows_sent: 1  Rows_examined: 0"
	useLine   = "use honeycomb;"
)

var t1, _ = time.Parse("02/Jan/2006:15:04:05.000000", "01/Apr/2016:00:31:09.817887")
var t2, _ = time.Parse("02/Jan/2006:15:04:05.000000", "01/Apr/2016:00:31:10.817887")
var tUnparseable, _ = time.Parse("02/Jan/2006:15:04:05.000000", "02/Aug/2010:13:24:56")

// `Time` field has ms resolution but no time zone; `SET timestamp=` is UNIX timestamp but no ms resolution
// e: mysql… i guess we could/should just combine the unix timestamp second and the parsed timestamp ms
// justify parsing both
// could be terrible

var sqds = []slowQueryData{
	{ /* 0 */
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"# Query_time: 0.008393  Lock_time: 0.000154 Rows_sent: 1  Rows_examined: 357",
		},
		sq: map[string]interface{}{
			queryTimeKey:    0.008393,
			lockTimeKey:     0.000154,
			rowsSentKey:     1,
			rowsExaminedKey: 357,
		},
		timestamp: t1,
	},
	{ /* 1 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"# User@Host: someuser @ hostfoo [192.168.2.1]  Id:   666",
		},
		sq: map[string]interface{}{
			userKey:     "someuser",
			clientKey:   "hostfoo",
			clientIPKey: "192.168.2.1",
		},
		timestamp: tUnparseable,
	},
	{ /* 2 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"# User@Host: root @ localhost []  Id:   233",
		},
		sq: map[string]interface{}{
			userKey:     "root",
			clientKey:   "localhost",
			clientIPKey: "",
		},
		timestamp: tUnparseable,
	},
	{ /* 3 */
		// RDS style user host line
		rawE: []string{
			"# User@Host: root[root] @  [10.0.1.76]  Id: 325920",
		},
		sq: map[string]interface{}{
			userKey:     "root",
			clientKey:   "",
			clientIPKey: "10.0.1.76",
		},
		timestamp: tUnparseable,
	},
	{ /* 4 */
		// RDS style user host line with hostname
		rawE: []string{
			"# User@Host: root[root] @ foobar [10.0.1.76]  Id: 325920",
		},
		sq: map[string]interface{}{
			userKey:     "root",
			clientKey:   "foobar",
			clientIPKey: "10.0.1.76",
		},
		timestamp: tUnparseable,
	},
	{ /* 5 */
		rawE: []string{
			"# Time: not-a-recognizable time stamp",
			"# administrator command: Ping;",
		},
		sq:        nil,
		timestamp: tUnparseable,
	},
	{ /* 6 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"show status like 'Uptime';",
		},
		sq: map[string]interface{}{
			queryKey:           "show status like 'Uptime'",
			normalizedQueryKey: "show status like ?",
			statementKey:       "",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 7 */
		// fails to parse "Time" but we capture unix time and we fall back to the scan normalizer
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT * FROM (SELECT  T1.orderNumber,  STATUS,  SUM(quantityOrdered * priceEach) AS  total FROM orders WHERE total > 1000 AS T1 INNER JOIN orderdetails AS T2 ON T1.orderNumber = T2.orderNumber GROUP BY  orderNumber) T WHERE total > 100;",
		},
		sq: map[string]interface{}{
			queryKey:           "SELECT * FROM (SELECT  T1.orderNumber,  STATUS,  SUM(quantityOrdered * priceEach) AS  total FROM orders WHERE total > 1000 AS T1 INNER JOIN orderdetails AS T2 ON T1.orderNumber = T2.orderNumber GROUP BY  orderNumber) T WHERE total > 100",
			normalizedQueryKey: "select * from (select t1.ordernumber, status, sum(quantityordered * priceeach) as total from orders where total > ? as t1 inner join orderdetails as t2 on t1.ordernumber = t2.ordernumber group by ordernumber) t where total > ?",
			statementKey:       "",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 8 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT * FROM orders WHERE total > 1000;",
		},
		sq: map[string]interface{}{
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 9 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT *",
			"FROM orders WHERE",
			"total > 1000;",
		},
		sq: map[string]interface{}{
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 10 */
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"use someDB;",
		},
		sq: map[string]interface{}{
			databaseKey:        "someDB",
			queryKey:           "use someDB",
			normalizedQueryKey: "use someDB",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 11 */
		// a use as well as query
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"use someDB;",
			"SELECT *",
			"FROM orders WHERE",
			"total > 1000;",
		},
		sq: map[string]interface{}{
			databaseKey:        "someDB",
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1.Truncate(time.Second),
	},
	// some tests for corrupted logs
	{ /* 12 */
		// invalid query + use + query, ignore the invalid query
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT *",
			"use someDB;",
			"SELECT *",
			"FROM orders WHERE",
			"total > 1000;",
		},
		sq: map[string]interface{}{
			databaseKey:        "someDB",
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{ /* 13 */
		// invalid query + set time + query, ignore the invalid query
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT * FROM orders WHERE",
			"SET timestamp=1459470670;",
			"SELECT * FROM orders WHERE total > 1000;",
		},
		sq: map[string]interface{}{
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t2.Truncate(time.Second),
	},
	{ /* 14 */
		// query + query_time comment + query, ignore the first query
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"SELECT * FROM orders WHERE total < 1000;",
			"# Query_time: 0.008393  Lock_time: 0.000154 Rows_sent: 1  Rows_examined: 357",
			"SELECT * FROM orders WHERE total > 1000;",
		},
		sq: map[string]interface{}{
			queryTimeKey:       0.008393,
			lockTimeKey:        0.000154,
			rowsSentKey:        1,
			rowsExaminedKey:    357,
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1,
	},
	{ /* 15 */
		// invalid query + user@host comment + query, ignore the invalid query
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"SELECT * FROM orders WHERE",
			"# User@Host: someuser @ hostfoo [192.168.2.1]  Id:   666",
			"SELECT * FROM orders WHERE total > 1000;",
		},
		sq: map[string]interface{}{
			userKey:            "someuser",
			clientKey:          "hostfoo",
			clientIPKey:        "192.168.2.1",
			queryKey:           "SELECT * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
		},
		timestamp: t1,
	},
	{ /* 16 */
		// query with a comment
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"# User@Host: someuser @ hostfoo [192.168.2.1]  Id:   666",
			"SELECT /* from mysql.go:245 */ /* another comment */ * FROM orders WHERE total > 1000;",
		},
		sq: map[string]interface{}{
			userKey:            "someuser",
			clientKey:          "hostfoo",
			clientIPKey:        "192.168.2.1",
			queryKey:           "SELECT /* from mysql.go:245 */ /* another comment */ * FROM orders WHERE total > 1000",
			normalizedQueryKey: "select * from orders where total > ?",
			tablesKey:          "orders",
			statementKey:       "select",
			commentsKey:        "/* from mysql.go:245 */ /* another comment */",
		},
		timestamp: t1,
	},
	{ /* 17 */
		// query without its last line
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"SELECT * FROM orders",
		},
		sq:        map[string]interface{}{},
		timestamp: t1,
	},
	{ /* 18 */
		rawE: []string{},
		sq:   map[string]interface{}{},
	},
	{ /* 19 */
		rawE: []string{
			"# Time: 160907  3:10:22",
			"# User@Host: rw[rw] @  [10.96.81.110]  Id: 1394495950",
			"# Schema: our_index  Last_errno: 0  Killed: 0",
			"# Query_time: 1.294391  Lock_time: 0.000119  Rows_sent: 4049  Rows_examined: 4049  Rows_affected: 1",
			"# Bytes_sent: 153824  Tmp_tables: 1  Tmp_disk_tables: 2  Tmp_table_sizes: 3",
			"# InnoDB_trx_id: A569C193C7",
			"# QC_Hit: No  Full_scan: No  Full_join: No  Tmp_table: No  Tmp_table_on_disk: No",
			"# Filesort: No  Filesort_on_disk: No  Merge_passes: 0",
			"#   InnoDB_IO_r_ops: 0  InnoDB_IO_r_bytes: 0  InnoDB_IO_r_wait: 0.000000",
			"#   InnoDB_rec_lock_wait: 0.000000  InnoDB_queue_wait: 0.000000",
			"#   InnoDB_pages_distinct: 6756",
			"SET timestamp=1473217822;",
			"/* [vreegU1vU6FPXHBnW6OU_DalWUR8] [our_index_A] [/dynamic_sitemaps.php] */ SELECT * FROM `cats_index` as `Cat_Cat` WHERE `Cat_Cat`.`cat_id` BETWEEN 9670064 AND 9680063 ORDER BY `Cat_Cat`.`cat_id`;",
		},
		sq: map[string]interface{}{
			userKey:            "rw",
			clientKey:          "",
			clientIPKey:        "10.96.81.110",
			queryTimeKey:       1.294391,
			lockTimeKey:        0.000119,
			rowsSentKey:        4049,
			rowsExaminedKey:    4049,
			rowsAffectedKey:    1,
			queryKey:           "/* [vreegU1vU6FPXHBnW6OU_DalWUR8] [our_index_A] [/dynamic_sitemaps.php] */ SELECT * FROM `cats_index` as `Cat_Cat` WHERE `Cat_Cat`.`cat_id` BETWEEN 9670064 AND 9680063 ORDER BY `Cat_Cat`.`cat_id`",
			normalizedQueryKey: "select * from `cats_index` as `cat_cat` where `cat_cat`.`cat_id` between ? and ? order by `cat_cat`.`cat_id`",
			statementKey:       "select",
			tablesKey:          "cat_cat cats_index",
			bytesSentKey:       153824,
			tmpTablesKey:       1,
			tmpDiskTablesKey:   2,
			tmpTableSizesKey:   3,
			transactionIDKey:   "A569C193C7",
			queryCacheHitKey:   false,
			fullScanKey:        false,
			fullJoinKey:        false,
			tmpTableKey:        false,
			tmpTableOnDiskKey:  false,
			fileSortKey:        false,
			fileSortOnDiskKey:  false,
			mergePassesKey:     0,
			ioROpsKey:          0,
			ioRBytesKey:        0,
			ioRWaitKey:         0.0,
			recLockWaitKey:     0.0,
			queueWaitKey:       0.0,
			pagesDistinctKey:   6756,
		},
		timestamp: time.Unix(1473217822, 0),
	},
	{ /* 20 */
		rawE: []string{
			"# Time: 160907  3:10:22",
			"# User@Host: rw[rw] @  [10.96.81.110]  Id: 1394495950",
			"# Schema: our_index  Last_errno: 0  Killed: 0",
			"# Query_time: 1.294391  Lock_time: 0.000119  Rows_sent: 4049  Rows_examined: 4049  Rows_affected: 1",
			"# Bytes_sent: 153824  Tmp_tables: 1  Tmp_disk_tables: 2  Tmp_table_sizes: 3",
			"# InnoDB_trx_id: A569C193C7",
			"# QC_Hit: Yes  Full_scan: Yes  Full_join: Yes  Tmp_table: Yes  Tmp_table_on_disk: Yes",
			"# Filesort: Yes  Filesort_on_disk: Yes  Merge_passes: 42",
			"#   InnoDB_IO_r_ops: 1  InnoDB_IO_r_bytes: 2  InnoDB_IO_r_wait: 3.000000",
			"#   InnoDB_rec_lock_wait: 4.000000  InnoDB_queue_wait: 5.000000",
			"#   InnoDB_pages_distinct: 6756",
			"SET timestamp=1473217822;",
			"/* [vreegU1vU6FPXHBnW6OU_DalWUR8] [our_index_A] [/dynamic_sitemaps.php] */ SELECT * FROM `cats_index` as `Cat_Cat` WHERE `Cat_Cat`.`cat_id` BETWEEN 9670064 AND 9680063 ORDER BY `Cat_Cat`.`cat_id`;",
		},
		sq: map[string]interface{}{
			userKey:            "rw",
			clientKey:          "",
			clientIPKey:        "10.96.81.110",
			queryTimeKey:       1.294391,
			lockTimeKey:        0.000119,
			rowsSentKey:        4049,
			rowsExaminedKey:    4049,
			rowsAffectedKey:    1,
			queryKey:           "/* [vreegU1vU6FPXHBnW6OU_DalWUR8] [our_index_A] [/dynamic_sitemaps.php] */ SELECT * FROM `cats_index` as `Cat_Cat` WHERE `Cat_Cat`.`cat_id` BETWEEN 9670064 AND 9680063 ORDER BY `Cat_Cat`.`cat_id`",
			normalizedQueryKey: "select * from `cats_index` as `cat_cat` where `cat_cat`.`cat_id` between ? and ? order by `cat_cat`.`cat_id`",
			statementKey:       "select",
			tablesKey:          "cat_cat cats_index",
			bytesSentKey:       153824,
			tmpTablesKey:       1,
			tmpDiskTablesKey:   2,
			tmpTableSizesKey:   3,
			transactionIDKey:   "A569C193C7",
			queryCacheHitKey:   true,
			fullScanKey:        true,
			fullJoinKey:        true,
			tmpTableKey:        true,
			tmpTableOnDiskKey:  true,
			fileSortKey:        true,
			fileSortOnDiskKey:  true,
			mergePassesKey:     42,
			ioROpsKey:          1,
			ioRBytesKey:        2,
			ioRWaitKey:         3.0,
			recLockWaitKey:     4.0,
			queueWaitKey:       5.0,
			pagesDistinctKey:   6756,
		},
		timestamp: time.Unix(1473217822, 0),
	},
	{ /* 21 */
		rawE: []string{
			"# User@Host: weaverw[weaverw] @ [10.14.214.13]",
			"# Thread_id: 78959 Schema: weave3 Last_errno: 0 Killed: 0",
			"# Query_time: 10.749944 Lock_time: 0.017599 Rows_sent: 0 Rows_examined: 0 Rows_affected: 10 Rows_read: 12",
			"# Bytes_sent: 51 Tmp_tables: 0 Tmp_disk_tables: 0 Tmp_table_sizes: 0",
			"# InnoDB_trx_id: 98CF",
			"use weave3;",
			"SET timestamp=1364506803;",
			"SELECT COUNT(*) FROM foo;",
		},
		sq: map[string]interface{}{
			userKey:            "weaverw",
			clientKey:          nil,
			clientIPKey:        "10.14.214.13",
			queryTimeKey:       10.749944,
			lockTimeKey:        0.017599,
			rowsSentKey:        0,
			rowsExaminedKey:    0,
			rowsAffectedKey:    10,
			queryKey:           "SELECT COUNT(*) FROM foo",
			normalizedQueryKey: "select count(*) from foo",
			statementKey:       "select",
			tablesKey:          "foo",
			bytesSentKey:       51,
			tmpTablesKey:       0,
			tmpDiskTablesKey:   0,
			tmpTableSizesKey:   0,
			transactionIDKey:   "98CF",
		},
		timestamp: time.Unix(1364506803, 0),
	},
}

func TestHandleEvent(t *testing.T) {
	p := &Parser{
		nower:      &FakeNower{},
		normalizer: &normalizer.Parser{},
	}
	for i, sqd := range sqds {
		res, timestamp := p.handleEvent(sqd.rawE)
		if len(res) != len(sqd.sq) {
			t.Errorf("case num %d: expected to parse %d fields, got %d", i, len(sqd.sq), len(res))
		}
		for k, v := range sqd.sq {
			if !reflect.DeepEqual(res[k], v) {
				t.Errorf("case num %d, key %s:\n\texpected:\t%+v\n\tgot:\t\t%+v", i, k, v, res[k])
			}
		}
		if timestamp.UnixNano() != sqd.timestamp.UnixNano() {
			t.Errorf("case num %d: time parsed incorrectly:\n\tExpected: %+v, Actual: %+v",
				i, sqd.timestamp, timestamp)
		}
	}
}

func TestTimeProcessing(t *testing.T) {
	p := &Parser{
		nower: &FakeNower{},
	}
	tsts := []struct {
		lines    []string
		expected time.Time
	}{
		{[]string{
			"# Time: 2016-09-15T10:16:17.898325Z", hostLine, timerLine, useLine,
			"SET timestamp=1473934577;",
		}, time.Date(2016, time.September, 15, 10, 16, 17, 898325000, time.UTC)},
		{[]string{
			"# Time: not-a-parsable-time-stampZ", hostLine, timerLine, useLine,
			"SET timestamp=1459470669;",
		}, time.Date(2016, time.April, 1, 0, 31, 9, 0, time.UTC)},
		{[]string{
			"# Time: 2016-09-16T19:37:39.006083Z", hostLine, timerLine, useLine,
		}, time.Date(2016, time.September, 16, 19, 37, 39, 6083000, time.UTC)},
		{[]string{hostLine, timerLine, useLine}, p.nower.Now()},
	}

	for _, tt := range tsts {
		_, timestamp := p.handleEvent(tt.lines)
		if timestamp.Unix() != tt.expected.Unix() {
			t.Errorf("Didn't capture unix ts from lines:\n%+v\n\tExpected: %d, Actual: %d",
				strings.Join(tt.lines, "\n"), tt.expected.Unix(), timestamp.Unix())
		}
		if timestamp.Nanosecond() != tt.expected.Nanosecond() {
			t.Errorf("Didn't capture time with MS resolution from lines:\n%+v\n\tExpected: %d, Actual: %d",
				strings.Join(tt.lines, "\n"), tt.expected.Nanosecond(), timestamp.Nanosecond())
		}
	}
}

// test that ProcessLines correctly splits the mysql slow query log stream into
// individual events. It should read in alternating sets of commented then
// uncommented lines and split them at the first comment after an uncommented
// line.
func TestProcessLines(t *testing.T) {
	ts1, _ := time.Parse(time.RFC3339Nano, "2016-04-01T00:31:09.817887Z")

	tsts := []struct {
		in       []string
		expected []event.Event
	}{
		{
			[]string{
				"# administrator command: Prepare;",
				"# Time: 2016-04-01T00:31:09.817887Z",
				"# User@Host: someuser @ hostfoo [192.168.2.1]  Id:   666",
				"# Query_time: 0.000073  Lock_time: 0.000000 Rows_sent: 0  Rows_examined: 0",
				"SELECT * FROM orders WHERE total > 1000;",
				"# administrator command: Prepare;",
				"# Time: 2016-04-01T00:31:09.817887Z",
				"# User@Host: otheruser @ hostbar [192.168.2.1]  Id:   666",
				"# Query_time: 0.00457  Lock_time: 0.1 Rows_sent: 5  Rows_examined: 35",
				"SELECT * FROM",
				"customers;",
			},
			[]event.Event{
				{
					Timestamp: ts1,
					Data: map[string]interface{}{
						"client":           "hostfoo",
						"client_ip":        "192.168.2.1",
						"user":             "someuser",
						"query_time":       0.000073,
						"lock_time":        0.0,
						"rows_sent":        0,
						"rows_examined":    0,
						"query":            "SELECT * FROM orders WHERE total > 1000",
						"normalized_query": "select * from orders where total > ?",
						"tables":           "orders",
						"statement":        "select",
					},
				},
				{
					Timestamp: ts1,
					Data: map[string]interface{}{
						"client":           "hostbar",
						"client_ip":        "192.168.2.1",
						"user":             "otheruser",
						"query_time":       0.00457,
						"lock_time":        0.1,
						"rows_sent":        5,
						"rows_examined":    35,
						"query":            "SELECT * FROM customers",
						"normalized_query": "select * from customers",
						"tables":           "customers",
						"statement":        "select",
					},
				},
			},
		},
		{ // missing a # Time: line on the second event
			[]string{
				"# Time: 151008  0:31:04",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.030974  Lock_time: 0.000019 Rows_sent: 0  Rows_examined: 30259",
				"SET timestamp=1444264264;",
				"SELECT `metadata`.* FROM `metadata` WHERE (`metadata`.app_id = 993089);",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921",
				"SET timestamp=1444264264;",
				"SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1;",
			},
			[]event.Event{
				{
					Timestamp: time.Unix(1444264264, 0),
					Data: map[string]interface{}{
						"client":           "",
						"client_ip":        "10.252.9.33",
						"user":             "rails",
						"query_time":       0.030974,
						"lock_time":        0.000019,
						"rows_sent":        0,
						"rows_examined":    30259,
						"query":            "SELECT `metadata`.* FROM `metadata` WHERE (`metadata`.app_id = 993089)",
						"normalized_query": "select `metadata`.* from `metadata` where (`metadata`.app_id = ?)",
						"tables":           "metadata",
						"statement":        "select",
					},
				},
				{
					Timestamp: time.Unix(1444264264, 0), // should pick up the SET timestamp=... cmd
					Data: map[string]interface{}{
						"client":           "",
						"client_ip":        "10.252.9.33",
						"user":             "rails",
						"query_time":       0.002280,
						"lock_time":        0.000023,
						"rows_sent":        0,
						"rows_examined":    921,
						"query":            "SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1",
						"normalized_query": "select `certs`.* from `certs` where (`certs`.app_id = ?) limit ?",
						"tables":           "certs",
						"statement":        "select",
					},
				},
			},
		},
		{ // statement blocks with no query should be skipped
			[]string{
				"# Time: 151008  0:31:04",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.030974  Lock_time: 0.000019 Rows_sent: 0  Rows_examined: 30259",
				"SET timestamp=1444264264;",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921",
				"SET timestamp=1444264264;",
				"SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1;",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921",
				"SET timestamp=1444264264;",
			},
			[]event.Event{
				{
					Timestamp: time.Unix(1444264264, 0), // should pick up the SET timestamp=... cmd
					Data: map[string]interface{}{
						"client":           "",
						"client_ip":        "10.252.9.33",
						"user":             "rails",
						"query_time":       0.002280,
						"lock_time":        0.000023,
						"rows_sent":        0,
						"rows_examined":    921,
						"query":            "SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1",
						"normalized_query": "select `certs`.* from `certs` where (`certs`.app_id = ?) limit ?",
						"tables":           "certs",
						"statement":        "select",
					},
				},
			},
		},
		{ // fewer queries than expected - only one query is here but two are
			// expected. put empty event there to match
			[]string{
				"# Time: 151008  0:31:04",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.030974  Lock_time: 0.000019 Rows_sent: 0  Rows_examined: 30259",
				"SET timestamp=1444264264;",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921",
				"SET timestamp=1444264264;",
				"SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1;",
				"# User@Host: rails[rails] @  [10.252.9.33]",
				"# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921",
				"SET timestamp=1444264264;",
			},
			[]event.Event{
				{
					Timestamp: time.Unix(1444264264, 0), // should pick up the SET timestamp=... cmd
					Data: map[string]interface{}{
						"client":           "",
						"client_ip":        "10.252.9.33",
						"user":             "rails",
						"query_time":       0.002280,
						"lock_time":        0.000023,
						"rows_sent":        0,
						"rows_examined":    921,
						"query":            "SELECT `certs`.* FROM `certs` WHERE (`certs`.app_id = 993089) LIMIT 1",
						"normalized_query": "select `certs`.* from `certs` where (`certs`.app_id = ?) limit ?",
						"tables":           "certs",
						"statement":        "select",
					},
				},
				{}, // to match already closed channel
			},
		},
	}

	for _, tt := range tsts {
		p := &Parser{
			nower:      &FakeNower{},
			normalizer: &normalizer.Parser{},
		}
		lines := make(chan string, 10)
		send := make(chan event.Event, 5)
		go func() {
			p.ProcessLines(lines, send)
			close(send)
		}()
		for _, line := range tt.in {
			lines <- line
		}
		close(lines)

		for _, exp := range tt.expected {
			ev := <-send
			if !ev.Timestamp.Equal(exp.Timestamp) {
				t.Errorf("time parsing mismatch. got %+v, expected %+v", ev.Timestamp, exp.Timestamp)
			}
			if !reflect.DeepEqual(ev.Data, exp.Data) {
				t.Errorf("data parsing mismatch. got %+v, expected %+v", ev.Data, exp.Data)
			}
		}
		if len(send) > 0 {
			t.Errorf("unexpected: %d additional events were extracted", len(send))
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse("02/Jan/2006:15:04:05.000000 -0700", "02/Aug/2010:13:24:56 -0000")
	return fakeTime
}
