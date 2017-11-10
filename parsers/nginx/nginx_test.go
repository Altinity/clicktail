package nginx

import (
	"log"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/honeycombio/gonx"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/httime/httimetest"
	"github.com/honeycombio/honeytail/parsers"
)

func init() {
	fakeNow, err := time.ParseInLocation(commonLogFormatTimeLayout, "02/Jan/2010:12:34:56 -0000", time.UTC)
	if err != nil {
		log.Fatal(err)
	}
	httime.DefaultNower = &httimetest.FakeNower{fakeNow}
}

type testLineMaps struct {
	line        string
	trimmedLine string
	ev          event.Event
}

func TestProcessLines(t *testing.T) {
	t1, _ := time.ParseInLocation(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 -0000", time.UTC)
	preReg := &parsers.ExtRegexp{regexp.MustCompile("^.*:..:.. (?P<pre_hostname>[a-zA-Z-.]+): ")}
	tlm := []testLineMaps{
		{
			line:        "Nov 05 10:23:45 myhost: https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			trimmedLine: "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"pre_hostname":           "myhost",
					"body_bytes_sent":        int64(174),
					"http_x_forwarded_proto": "https",
					"remote_addr":            "10.252.4.24",
					"request_time":           0.099,
					"status":                 int64(200),
				},
			},
		},
	}
	p := &Parser{
		conf: Options{
			NumParsers: 5,
		},
		lineParser: &GonxLineParser{
			parser: gonx.NewParser("$http_x_forwarded_proto - $remote_addr - $remote_user [$time_local] $status $body_bytes_sent $request_time"),
		},
	}
	lines := make(chan string)
	send := make(chan event.Event)
	go func() {
		for _, pair := range tlm {
			lines <- pair.line
		}
		close(lines)
	}()
	go p.ProcessLines(lines, send, preReg)
	for _, pair := range tlm {
		resp := <-send
		if !reflect.DeepEqual(resp, pair.ev) {
			t.Fatalf("line resp didn't match up for %s. Expected: %+v, actual: %+v",
				pair.line, pair.ev, resp)
		}
	}
}

func TestProcessLinesNoPreReg(t *testing.T) {
	t1, _ := time.ParseInLocation(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 +0000", time.UTC)
	tlm := []testLineMaps{
		{
			line:        "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			trimmedLine: "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"body_bytes_sent":        int64(174),
					"http_x_forwarded_proto": "https",
					"remote_addr":            "10.252.4.24",
					"request_time":           0.099,
					"status":                 int64(200),
				},
			},
		},
	}
	p := &Parser{
		conf: Options{
			NumParsers: 5,
		},
		lineParser: &GonxLineParser{
			parser: gonx.NewParser("$http_x_forwarded_proto - $remote_addr - $remote_user [$time_local] $status $body_bytes_sent $request_time"),
		},
	}
	lines := make(chan string)
	send := make(chan event.Event)
	go func() {
		for _, pair := range tlm {
			lines <- pair.line
		}
		close(lines)
	}()
	go p.ProcessLines(lines, send, nil)
	for _, pair := range tlm {
		resp := <-send
		if !reflect.DeepEqual(resp, pair.ev) {
			t.Fatalf("line resp didn't match up for %s. Expected: %v, actual: %v",
				pair.line, pair.ev.Data, resp.Data)
		}
	}
}

type typeifyTestCase struct {
	untyped map[string]string
	typed   map[string]interface{}
}

func TestTypeifyParsedLine(t *testing.T) {
	tc := typeifyTestCase{
		untyped: map[string]string{
			"str":    "str",            // should stay string
			"space":  "str with space", // should stay string
			"ver":    "5.1.0",          // should stay string
			"dash":   "-",              // should vanish
			"float":  "4.134",          // should become float
			"int":    "987",            // should become int
			"negint": "-5",             // should become int
		},
		typed: map[string]interface{}{
			"str":    "str",
			"space":  "str with space",
			"ver":    "5.1.0",
			"float":  float64(4.134),
			"int":    int64(987),
			"negint": int64(-5),
		},
	}
	res := typeifyParsedLine(tc.untyped)
	if !reflect.DeepEqual(res, tc.typed) {
		t.Fatalf("Comparison failed. Expected: %v, Actual: %v", tc.typed, res)
	}
}

func TestGetTimestamp(t *testing.T) {
	t1, _ := time.ParseInLocation(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 +0000", time.UTC)
	t2, _ := time.ParseInLocation(commonLogFormatTimeLayout, "02/Jan/2010:12:34:56 -0000", time.UTC)
	userDefinedTimeFormat := "2006-01-02T15:04:05.9999Z"
	exampleCustomFormatTimestamp := "2017-07-31T20:40:57.980264Z"
	t3, _ := time.ParseInLocation(userDefinedTimeFormat, exampleCustomFormatTimestamp, time.UTC)
	testCases := []struct {
		desc      string
		conf      Options
		input     map[string]interface{}
		postMunge map[string]interface{}
		retval    time.Time
	}{
		{
			desc: "well formatted time_local",
			conf: Options{},
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": "08/Oct/2015:00:26:26 +0000",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t1,
		},
		{
			desc: "well formatted time_iso",
			conf: Options{},
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": "2015-10-08T00:26:26-00:00",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t1,
		},
		{
			desc: "broken formatted time_local",
			conf: Options{},
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": "08aoeu00:26:26 +0000",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{
			desc: "broken formatted time_iso",
			conf: Options{},
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": "2015-aoeu00:00",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{
			desc: "non-string formatted time_local",
			conf: Options{},
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": 1234,
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{
			desc: "non-string formatted time_iso",
			conf: Options{},
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": 1234,
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{
			desc: "missing time field",
			conf: Options{},
			input: map[string]interface{}{
				"foo": "bar",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{
			desc: "timestamp in user-defined format",
			conf: Options{
				TimeFieldFormat: userDefinedTimeFormat,
				TimeFieldName:   "timestamp",
			},
			input: map[string]interface{}{
				"foo":       "bar",
				"timestamp": exampleCustomFormatTimestamp,
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t3,
		},
	}

	for _, tc := range testCases {
		parser := &Parser{
			conf: tc.conf,
		}
		res := parser.getTimestamp(tc.input)
		if !reflect.DeepEqual(tc.input, tc.postMunge) {
			t.Errorf("didn't remove time field in %q: %v", tc.desc, tc.input)
		}
		if !reflect.DeepEqual(res, tc.retval) {
			t.Errorf("got wrong time in %q. expected %v got %v", tc.desc, tc.retval, res)
		}
	}
}
