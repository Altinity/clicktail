package mysql

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/honeycombio/mysqltools/query/normalizer"
)

type slowQueryData struct {
	rawE      []string
	sq        SlowQuery
	timestamp time.Time
	// processSlowQuery will error
	psqWillError bool
}

const (
	hostLine  = "# User@Host: root[root] @ localhost [127.0.0.1]  Id:   136"
	timerLine = "# Query_time: 0.000171  Lock_time: 0.000000 Rows_sent: 1  Rows_examined: 0"
	useLine   = "use honeycomb;"
)

var t1, _ = time.Parse("02/Jan/2006:15:04:05.000000", "01/Apr/2016:00:31:09.817887")
var tUnparseable, _ = time.Parse("02/Jan/2006:15:04:05.000000", "02/Aug/2010:13:24:56")

// `Time` field has ms resolution but no time zone; `SET timestamp=` is UNIX timestamp but no ms resolution
// e: mysql… i guess we could/should just combine the unix timestamp second and the parsed timestamp ms
// justify parsing both
// could be terrible

func intptr(i int) *int           { return &i }
func floatptr(f float64) *float64 { return &f }

var sqds = []slowQueryData{
	{
		rawE: []string{
			"# Time: 2016-04-01T00:31:09.817887Z",
			"# Query_time: 0.008393  Lock_time: 0.000154 Rows_sent: 1  Rows_examined: 357",
		},
		sq: SlowQuery{
			QueryTime:    floatptr(0.008393),
			LockTime:     floatptr(0.000154),
			RowsSent:     intptr(1),
			RowsExamined: intptr(357),
		},
		timestamp: t1,
	},
	{
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"# User@Host: someuser @ hostfoo [192.168.2.1]  Id:   666",
		},
		sq: SlowQuery{
			User:     "someuser",
			Client:   "hostfoo",
			ClientIP: "192.168.2.1",
		},
		timestamp: tUnparseable,
	},
	{
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"# User@Host: root @ localhost []  Id:   233",
		},
		sq: SlowQuery{
			User:   "root",
			Client: "localhost",
		},
		timestamp: tUnparseable,
	},
	{
		rawE: []string{
			"# Time: not-a-recognizable time stamp",
			"# administrator command: Ping;",
		},
		sq: SlowQuery{
			skipQuery: true,
		},
		timestamp:    tUnparseable,
		psqWillError: true,
	},
	{
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"show status like 'Uptime';",
		},
		sq: SlowQuery{
			Query:           "show status like 'Uptime'",
			NormalizedQuery: "show status like ?",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{
		// fails to parse "Time" but we capture unix time and we fall back to the scan normalizer
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT * FROM (SELECT  T1.orderNumber,  STATUS,  SUM(quantityOrdered * priceEach) AS  total FROM orders WHERE total > 1000 AS T1 INNER JOIN orderdetails AS T2 ON T1.orderNumber = T2.orderNumber GROUP BY  orderNumber) T WHERE total > 100;",
		},
		sq: SlowQuery{
			Query:           "SELECT * FROM (SELECT  T1.orderNumber,  STATUS,  SUM(quantityOrdered * priceEach) AS  total FROM orders WHERE total > 1000 AS T1 INNER JOIN orderdetails AS T2 ON T1.orderNumber = T2.orderNumber GROUP BY  orderNumber) T WHERE total > 100",
			NormalizedQuery: "select * from (select t1.ordernumber, status, sum(quantityordered * priceeach) as total from orders where total > ? as t1 inner join orderdetails as t2 on t1.ordernumber = t2.ordernumber group by ordernumber) t where total > ?",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"SELECT * FROM orders WHERE total > 1000;",
		},
		sq: SlowQuery{
			Query:           "SELECT * FROM orders WHERE total > 1000",
			NormalizedQuery: "select * from orders where total > ?",
			Tables:          "orders",
			Statement:       "select",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{
		rawE: []string{
			"# Time: not-a-parsable-time-stampZ",
			"SET timestamp=1459470669;",
			"use someDB;",
		},
		sq: SlowQuery{
			DB:              "someDB",
			Query:           "use someDB",
			NormalizedQuery: "use someDB",
		},
		timestamp: t1.Truncate(time.Second),
	},
	{
		rawE: []string{},
		sq:   SlowQuery{},
	},
}

func TestHandleEvent(t *testing.T) {
	p := &Parser{
		nower:      &FakeNower{},
		normalizer: &normalizer.Parser{},
	}
	for i, sqd := range sqds {
		res, timestamp := p.handleEvent(sqd.rawE)
		if !reflect.DeepEqual(res, sqd.sq) {
			t.Errorf("case num %d: expected\n %+v, got\n %+v", i, sqd.sq, res)
		}
		if timestamp.UnixNano() != sqd.timestamp.UnixNano() {
			t.Errorf("case num %d: time parsed incorrectly:\n\tExpected: %+v, Actual: %+v",
				i, sqd.timestamp, timestamp)
		}
	}
}

func TestProcessSlowQuery(t *testing.T) {
	p := &Parser{
		nower:      &FakeNower{},
		normalizer: &normalizer.Parser{},
	}
	for i, sqd := range sqds {
		res, err := p.processSlowQuery(sqd.sq, sqd.timestamp)
		if err == nil && sqd.psqWillError {
			t.Errorf("case num %d: expected processSlowQuery to error (%+v) but it didn't. sq: %+v, res: %+v", i, err, sqd, res)
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
		res, timestamp := p.handleEvent(tt.lines)
		if timestamp.Unix() != tt.expected.Unix() {
			t.Errorf("Didn't capture unix ts from lines:\n%+v\n\tExpected: %d, Actual: %d",
				strings.Join(tt.lines, "\n"), tt.expected.Unix(), timestamp.Unix())
		}
		if timestamp.Nanosecond() != tt.expected.Nanosecond() {
			t.Errorf("Didn't capture time with MS resolution from lines:\n%+v\n\tExpected: %d, Actual: %d",
				strings.Join(tt.lines, "\n"), tt.expected.Nanosecond(), timestamp.Nanosecond())
		}

		ev, err := p.processSlowQuery(res, timestamp)
		if err != nil {
			t.Error("unexpected error processing SlowQuery:", err)
		}
		if ev.Timestamp.UnixNano() != tt.expected.UnixNano() {
			t.Errorf("Processed SlowQuery should contain correct unix ts\n%+v\n\tExpected: %d, Actual: %d (%+v)",
				strings.Join(tt.lines, "\n"), tt.expected.UnixNano(), ev.Timestamp.UnixNano(), ev.Timestamp)
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse("02/Jan/2006:15:04:05.000000 -0700", "02/Aug/2010:13:24:56 -0000")
	return fakeTime
}
