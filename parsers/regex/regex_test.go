package regex

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

const (
	commonLogFormatTimeLayout = "02/Jan/2006:15:04:05 -0700"
	iso8601TimeLayout         = "2006-01-02T15:04:05-07:00"
)

// Test Init(...) success/fail

type testInitMap struct {
	options      *Options
	expectedPass bool
}

var testInitCases = []testInitMap{
	{
		expectedPass: true,
		options: &Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
			LineRegex:       `(?P<foo>[A-Za-z]+)`,
		},
	},
	{
		expectedPass: false,
		options: &Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
			LineRegex:       ``, // Empty regex should fail
		},
	},
	{
		expectedPass: false,
		options: &Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
			LineRegex:       `(?P<foo>[A-Za-`, // Broken regex should fail
		},
	},
	{
		expectedPass: false,
		options: &Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
			LineRegex:       `[a-z]+`, // Require at least one named group
		},
	},
	{
		expectedPass: false,
		options: &Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
			LineRegex:       `(?P[a-z]+)`, // Require at least one named group
		},
	},
}

func TestInit(t *testing.T) {
	for _, testCase := range testInitCases {
		p := &Parser{}
		err := p.Init(testCase.options)
		if (err == nil) != testCase.expectedPass {
			if err == nil {
				t.Error("Parser Init(...) passed; expected it to fail.")
			} else {
				t.Error("Parser Init(...) failed; expected it to pass. Error:", err)
			}
		} else {
			t.Logf("Init pass status is %t as expected", (err == nil))
		}
	}
}

// Test cases for RegexLineParser.ParseLine

type testLineMap struct {
	lineRegex string
	input     string
	expected  map[string]interface{}
}

var tlms = []testLineMap{
	{
		// Simple word parsing
		lineRegex: `(?P<word1>\w+) (?P<word2>\w+) (?P<word3>\w+)`,
		input:     `apple banana orange`,
		expected: map[string]interface{}{
			"word1": "apple",
			"word2": "banana",
			"word3": "orange",
		},
	},
	{
		// Simple time parsing
		lineRegex: `(?P<Year>\d{4})-(?P<Month>\d{2})-(?P<Day>\d{2})`,
		input:     `2017-01-30 1980-01-02`, // Ignore the second date
		expected: map[string]interface{}{
			"Year":  "2017",
			"Month": "01",
			"Day":   "30",
		},
	},
	{
		// Fields containing whitespace
		lineRegex: `\[(?P<BracketedField>[0-9A-Za-z\s]+)\] (?P<UnbracketedField>[0-9A-Za-z]+)`,
		input:     `[some value] unbracketed`,
		expected: map[string]interface{}{
			"BracketedField":   "some value",
			"UnbracketedField": "unbracketed",
		},
	},
	{
		// Nested regex grouping
		lineRegex: `(?P<outer>[^ ]* (?P<inner1>[^ ]*) (?P<inner2>[^ ]*))`,
		input:     `foo bar baz`,
		expected: map[string]interface{}{
			"outer":  "foo bar baz",
			"inner1": "bar",
			"inner2": "baz",
		},
	},
	{
		// Sample nginx error log line
		lineRegex: `(?P<time>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}) \[(?P<status>.*)\].* request: "(?P<request>[^"]*)"`,
		input:     `2017/11/07 22:59:46 [error] 5812#0: *777536449 connect() failed (111: Connection refused) while connecting to upstream, client: 127.0.0.1, server: localhost, request: "GET /isbns HTTP/1.1", upstream: "http://127.0.0.1:8080/isbns", host: "localhost"`,
		expected: map[string]interface{}{
			"time":    "2017/11/07 22:59:46",
			"status":  "error",
			"request": "GET /isbns HTTP/1.1",
		},
	},
}

func TestParseLine(t *testing.T) {
	p := &Parser{
		conf: Options{
			NumParsers: 5,
		},
		lineParser: &RegexLineParser{regexp.MustCompile(``)},
	}
	for _, tlm := range tlms {
		p.lineParser = &RegexLineParser{regexp.MustCompile(tlm.lineRegex)}
		resp, err := p.lineParser.ParseLine(tlm.input)
		t.Logf("%+v", resp)
		if err != nil {
			t.Error("p.ParseLine unexpectedly returned error ", err)
		}
		if !reflect.DeepEqual(resp, tlm.expected) {
			t.Errorf("response %+v didn't match expected %+v", resp, tlm.expected)
		}
	}
}

type testLineMaps struct {
	line        string
	trimmedLine string
	resp        map[string]interface{}
	typedResp   map[string]interface{}
	ev          event.Event
}

// Test event emitted from ProcessLines
func TestProcessLines(t *testing.T) {
	t1, _ := time.ParseInLocation(commonLogFormatTimeLayout, "08/Oct/2015:00:26:26 -0000", time.UTC)
	preReg := &parsers.ExtRegexp{regexp.MustCompile("^.*:..:.. (?P<pre_hostname>[a-zA-Z-.]+): ")}
	tlm := []testLineMaps{
		{
			line: "https - 10.252.4.24 - - [08/Oct/2015:00:26:26 +0000] 200 174 0.099",
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"http_x_forwarded_proto": "https",
					"remote_addr":            "10.252.4.24",
				},
			},
		},
	}
	lineRegex, err := regexp.Compile(`(?P<http_x_forwarded_proto>\w+) - (?P<remote_addr>\d{1,4}\.\d{1,4}\.\d{1,4}\.\d{1,4}) - - \[(?P<local_time>\d{2}\/[A-Za-z]+\/\d{4}:\d{2}:\d{2}:\d{2}.*)\]`)
	if err != nil {
		t.Errorf("couldnt parse regex")
	}
	p := &Parser{
		conf: Options{
			NumParsers:      5,
			TimeFieldName:   "local_time",
			TimeFieldFormat: "%d/%b/%Y:%H:%M:%S %z",
		},
		lineParser: &RegexLineParser{lineRegex},
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
