package nginx

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

type testLineMaps struct {
	line        string
	trimmedLine string
	resp        map[string]string
	typedResp   map[string]interface{}
	ev          event.Event
}

type FakeLineParser struct {
	tlm []testLineMaps
}

func (f *FakeLineParser) ParseLine(line string) (map[string]string, error) {
	for _, pair := range f.tlm {
		if pair.trimmedLine == line {
			return pair.resp, nil
		}
	}
	return nil, errors.New("failed to parse line")
}

func TestProcessLines(t *testing.T) {
	t1, _ := time.Parse(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 +0000")
	preReg := &parsers.ExtRegexp{regexp.MustCompile("^.*:..:.. (?P<pre_hostname>[a-zA-Z-.]+): ")}
	tlm := []testLineMaps{
		{
			line:        "Nov 05 10:23:45 myhost: https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			trimmedLine: "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			resp: map[string]string{
				"http_x_forwarded_proto": "https",
				"remote_addr":            "10.252.4.24",
				"remote_user":            "-",
				"time_local":             "08/Oct/2015:00:26:26 +0000",
				"status":                 "200",
				"body_bytes_sent":        "174",
				"request_time":           "0.099",
			},
			typedResp: map[string]interface{}{
				"pre_hostname":           "myhost",
				"http_x_forwarded_proto": "https",
				"remote_addr":            "10.252.4.24",
				"remote_user":            "-",
				"time_local":             "08/Oct/2015:00:26:26 +0000",
				"status":                 200,
				"body_bytes_sent":        174,
				"request_time":           0.099,
			},
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
		conf: Options{},
		lineParser: &FakeLineParser{
			tlm: tlm,
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
			t.Fatalf("line resp didn't match up for %s. Expected: %v, actual: %v",
				pair.line, pair.ev.Data, resp.Data)
		}
	}
}
func TestProcessLinesNoPreReg(t *testing.T) {
	t1, _ := time.Parse(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 +0000")
	tlm := []testLineMaps{
		{
			line:        "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			trimmedLine: "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			resp: map[string]string{
				"http_x_forwarded_proto": "https",
				"remote_addr":            "10.252.4.24",
				"remote_user":            "-",
				"time_local":             "08/Oct/2015:00:26:26 +0000",
				"status":                 "200",
				"body_bytes_sent":        "174",
				"request_time":           "0.099",
			},
			typedResp: map[string]interface{}{
				"http_x_forwarded_proto": "https",
				"remote_addr":            "10.252.4.24",
				"remote_user":            "-",
				"time_local":             "08/Oct/2015:00:26:26 +0000",
				"status":                 200,
				"body_bytes_sent":        174,
				"request_time":           0.099,
			},
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
		conf: Options{},
		lineParser: &FakeLineParser{
			tlm: tlm,
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
	res, _ := typeifyParsedLine(tc.untyped)
	if !reflect.DeepEqual(res, tc.typed) {
		t.Fatalf("Comparison failed. Expected: %v, Actual: %v", tc.typed, res)
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(commonLogFormatTimeLayout, "02/Jan/2010:12:34:56 -0000")
	return fakeTime
}

func TestGetTimestamp(t *testing.T) {
	t1, _ := time.Parse(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 +0000")
	t2, _ := time.Parse(commonLogFormatTimeLayout, "02/Jan/2010:12:34:56 -0000")
	testCases := []struct {
		input     map[string]interface{}
		postMunge map[string]interface{}
		retval    time.Time
	}{
		{ //well formatted time_local
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": "08/Oct/2015:00:26:26 +0000",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t1,
		},
		{ //well formatted time_iso
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": "2015-10-08T00:26:26-00:00",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t1,
		},
		{ //broken formatted time_local
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": "08aoeu00:26:26 +0000",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{ //broken formatted time_iso
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": "2015-aoeu00:00",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{ //non-string formatted time_local
			input: map[string]interface{}{
				"foo":        "bar",
				"time_local": 1234,
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{ //non-string formatted time_iso
			input: map[string]interface{}{
				"foo":          "bar",
				"time_iso8601": 1234,
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
		{ //missing time field
			input: map[string]interface{}{
				"foo": "bar",
			},
			postMunge: map[string]interface{}{
				"foo": "bar",
			},
			retval: t2,
		},
	}
	for _, tc := range testCases {
		res := getTimestamp(&FakeNower{}, tc.input)
		if !reflect.DeepEqual(tc.input, tc.postMunge) {
			t.Errorf("didn't remove time field: %v", tc.input)
		}
		if !reflect.DeepEqual(res, tc.retval) {
			t.Errorf("got wrong time. expected %v got %v", tc.retval, res)
		}
	}
}
