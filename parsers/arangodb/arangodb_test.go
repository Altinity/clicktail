package arangodb

import (
	"reflect"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
)

const (
	DEBUGLINE = `2016-10-31T16:03:02Z [6402] DEBUG {requests} "http-request-begin","0x7f87ba86b290","127.0.0.1","GET","HTTP/1.1",/_api/version`

	INFOLINE = `2016-10-31T16:03:02Z [6402] INFO {requests} "http-request-end","0x7f87ba86b290","127.0.0.1","GET","HTTP/1.1",200,0,64,"/_api/version",0.000139`

	TIMESTRING = `2016-10-31T16:03:02Z`
)

var (
	T1, _ = time.Parse(iso8601UTCTimeFormat, TIMESTRING)
)

type processed struct {
	time        time.Time
	includeData map[string]interface{}
	excludeKeys []string
}

func TestProcessLines(t *testing.T) {
	tlm := []struct {
		line     string
		expected processed
	}{
		{
			line: DEBUGLINE,
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"pid":      "6402",
					"logLevel": "DEBUG",
					"id":       "0x7f87ba86b290",
					"sourceIP": "127.0.0.1",
					"method":   "GET",
					"protocol": "HTTP/1.1",
					"fullURL":  "/_api/version",
				},
			},
		},

		{
			line: INFOLINE,
			expected: processed{
				time: T1,
				includeData: map[string]interface{}{
					"pid":          "6402",
					"logLevel":     "INFO",
					"id":           "0x7f87ba86b290",
					"sourceIP":     "127.0.0.1",
					"method":       "GET",
					"protocol":     "HTTP/1.1",
					"responseCode": int64(200),
					"reqBodyLen":   int64(0),
					"resBodyLen":   int64(64),
					"fullURL":      "/_api/version",
					"totalTime":    0.000139,
				},
			},
		},
	}
	m := &Parser{
		conf:       Options{},
		lineParser: &ArangoLineParser{},
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
	go m.ProcessLines(lines, send, nil)
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
