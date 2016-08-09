package mysql

import (
	"reflect"
	"testing"
	"time"
)

type slowQueryData struct {
	rawE rawEvent
	sq   SlowQuery
	// processSlowQuery will error
	psqWillError bool
}

var t1, _ = time.Parse("02/Jan/2006:15:04:05.000000", "01/Apr/2016:00:31:09.817887")
var t2, _ = time.Parse("02/Jan/2006:15:04:05.000000", "02/Aug/2010:13:24:56")
var sqds = []slowQueryData{
	{
		rawE: rawEvent{
			lines: []string{
				"# Time: 2016-04-01T00:31:09.817887Z",
				"# Query_time: 0.008393  Lock_time: 0.000154 Rows_sent: 1  Rows_examined: 357",
			},
		},
		sq: SlowQuery{
			Timestamp:    t1,
			QueryTime:    0.008393,
			LockTime:     0.000154,
			RowsSent:     1,
			RowsExamined: 357,
		},
		psqWillError: false,
	},
	{
		rawE: rawEvent{
			lines: []string{
				"# Time: not-a-parsable-time-stampZ",
				"# User@Host: root[root] @ localhost []  Id:   233",
			},
		},
		sq: SlowQuery{
			Timestamp: t2,
			User:      "root[root]",
			Host:      "localhost",
		},
	},
	{
		rawE: rawEvent{
			lines: []string{
				"# Time: not-a-recognizable time stamp",
				"# administrator command: Ping;",
			},
		},
		sq: SlowQuery{
			Timestamp: t2,
			skipQuery: true,
		},
		psqWillError: true,
	},
	{
		rawE: rawEvent{
			lines: []string{
				"# Time: not-a-parsable-time-stampZ",
				"SET timestamp=1459470669;",
				"show status like 'Uptime';",
			},
		},
		sq: SlowQuery{
			Timestamp: t2,
			UnixTime:  1459470669,
			Query:     "show status like 'Uptime';",
		},
	},
	{
		rawE: rawEvent{
			lines: []string{},
		},
		sq: SlowQuery{},
	},
}

func TestHandleEvent(t *testing.T) {
	p := &Parser{
		nower: &FakeNower{},
	}
	for i, sqd := range sqds {
		res := p.handleEvent(sqd.rawE)
		if !reflect.DeepEqual(res, sqd.sq) {
			t.Errorf("case num %d: expected %+v, got %+v", i, sqd.sq, res)
		}
	}
}

func TestProcessSlowQuery(t *testing.T) {
	p := &Parser{
		nower: &FakeNower{},
	}
	for i, sqd := range sqds {
		res, err := p.processSlowQuery(sqd.sq)
		if err == nil && sqd.psqWillError {
			t.Fatalf("case num %d: expected processSlowQuery to error (%+v) but it didn't. sq: %+v, res: %+v", i, err, sqd, res)
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse("02/Jan/2006:15:04:05.000000 -0700", "02/Aug/2010:13:24:56 -0000")
	return fakeTime
}
