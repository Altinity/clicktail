package keyval

import (
	"reflect"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
)

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(time.RFC3339, "2010-06-21T15:04:05Z")
	return fakeTime
}

type testLineMap struct {
	input    string
	expected map[string]interface{}
}

var tlms = []testLineMap{
	{ // strings, floats, and ints
		input: `mystr="myval" myint=3 myfloat=4.234 mybool=true`,
		expected: map[string]interface{}{
			"mystr":   "myval",
			"myint":   3,
			"myfloat": 4.234,
			"mybool":  true,
		},
	},
	{ // missing keyval pairs
		input: `foo bar 123 baz`,
		expected: map[string]interface{}{
			"foo": "",
			"bar": "",
			"123": "",
			"baz": "",
		},
	},
	{ // time
		input: `time="2014-03-10 19:57:38.123456789 -0800 PST" myint=3 myfloat=4.234`,
		expected: map[string]interface{}{
			"time":    "2014-03-10 19:57:38.123456789 -0800 PST",
			"myint":   3,
			"myfloat": 4.234,
		},
	},
}

func TestParseLine(t *testing.T) {
	jlp := KeyValLineParser{}
	for _, tlm := range tlms {
		resp, err := jlp.ParseLine(tlm.input)
		if err != nil {
			t.Error("jlp.ParseLine unexpectedly returned error ", err)
		}
		if !reflect.DeepEqual(resp, tlm.expected) {
			t.Errorf("response %+v didn't match expected %+v", resp, tlm.expected)
		}
	}
}

func TestBrokenFilterRegex(t *testing.T) {
	// test filter that doesn't compile
	broken := &Parser{}
	err := broken.Init(&Options{
		FilterRegex: "regex [ won't compile",
	})
	if err == nil {
		t.Error("Parser Init with broken regex should err, instead got nil")
	}
}

func TestFilterRegex(t *testing.T) {
	p := &Parser{
		lineParser: &NoopLineParser{
			outgoingMap: map[string]interface{}{"key": "val"},
		},
		nower: &FakeNower{},
	}
	tsts := []struct {
		filterString   string
		invertFilter   bool
		lines          []string
		expectedEvents int
	}{
		{
			"aaaa",
			false,
			[]string{
				"line one",
				"line two aoeu",
				"line three",
			},
			0, // no lines have 'aaaa'
		},
		{
			"aaaa",
			true,
			[]string{
				"line one",
				"line two aoeu",
				"line three",
			},
			3, // all lines don't have 'aaaa'
		},
		{
			"aoeu",
			false,
			[]string{
				"line one",
				"line two aoeu",
				"line three",
			},
			1, // only line two has 'aoeu'
		},
		{
			"aoeu",
			true,
			[]string{
				"line one",
				"line two aoeu",
				"line three",
			},
			2, // lines one and three don't have 'aoeu'
		},
	}
	for _, tst := range tsts {
		p.filterRegex = regexp.MustCompile(tst.filterString)
		p.conf.InvertFilter = tst.invertFilter
		lines := make(chan string)
		send := make(chan event.Event)
		// send input into lines in a goroutine then close the lines channel
		go func() {
			for _, line := range tst.lines {
				lines <- line
			}
			close(lines)
		}()
		// read from the send channel and see if we got back what we expected
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			var counter int
			for range send {
				counter++
			}
			if counter != tst.expectedEvents {
				t.Errorf("expected %d messages out the send channel, got %d\n", tst.expectedEvents, counter)
			}
			wg.Done()
		}()
		p.ProcessLines(lines, send, nil)
		close(send)
		wg.Wait()
	}
}

func TestDontReturnEmptyEvents(t *testing.T) {
	p := &Parser{
		lineParser: &NoopLineParser{},
		nower:      &FakeNower{},
	}
	lines := make(chan string)
	send := make(chan event.Event)
	// send input into lines in a goroutine then close the lines channel
	go func() {
		for _, line := range []string{"one", "two", "three"} {
			lines <- line
		}
		close(lines)
	}()
	// read from the send channel and see if we got back what we expected
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		var counter int
		for range send {
			counter++
		}
		if counter != 0 {
			t.Errorf("expected no messages out the send channel, got %d\n", counter)
		}
		wg.Done()
	}()
	p.ProcessLines(lines, send, nil)
	close(send)
	wg.Wait()
}

// TestDontReturnUselessEvents a useless event is one with all keys and no
// values
func TestDontReturnUselessEvents(t *testing.T) {
	p := &Parser{
		lineParser: &NoopLineParser{
			outgoingMap: map[string]interface{}{
				"key": "",
				"k2":  "",
			},
		},
		nower: &FakeNower{},
	}
	lines := make(chan string)
	send := make(chan event.Event)
	// send input into lines in a goroutine then close the lines channel
	go func() {
		for _, line := range []string{"one", "two", "three"} {
			lines <- line
		}
		close(lines)
	}()
	// read from the send channel and see if we got back what we expected
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		var counter int
		for range send {
			counter++
		}
		if counter != 0 {
			t.Errorf("expected no messages out the send channel, got %d\n", counter)
		}
		wg.Done()
	}()
	p.ProcessLines(lines, send, nil)
	close(send)
	wg.Wait()
}

func TestAllEmpty(t *testing.T) {
	tsts := []struct {
		incoming map[string]interface{}
		empty    bool
	}{
		{
			map[string]interface{}{
				"k1": "v1",
			},
			false,
		},
		{
			map[string]interface{}{
				"k1": 3,
			},
			false,
		},
		{
			map[string]interface{}{
				"k1": []string{"foo", "bar"},
			},
			false,
		},
		{
			map[string]interface{}{},
			true,
		},
		{
			map[string]interface{}{
				"k1": "",
			},
			true,
		},
		{
			map[string]interface{}{
				"k1": "",
				"k2": "",
				"k3": "",
			},
			true,
		},
	}
	for _, tst := range tsts {
		res := allEmpty(tst.incoming)
		if res != tst.empty {
			t.Errorf("expected %v's empty val would be %v, got %v",
				tst.incoming, tst.empty, res)
		}
	}
}

type testTimestamp struct {
	format    string      // the format this test's time is in
	fieldName string      // the field in the map containing the time
	input     interface{} // the value corresponding to the fieldName
	auto      bool        // whether the input should be parsable even without specifying format/fieldName
	expected  time.Time   // the expected time object to get back
}

var tts = []testTimestamp{
	{
		format:    "2006-01-02 15:04:05.999999999 -0700 MST",
		fieldName: "time",
		input:     "2014-04-10 19:57:38.123456789 -0800 PST",
		auto:      true,
		expected:  time.Unix(1397188658, 123456789),
	},
	{
		format:    time.RFC3339Nano,
		fieldName: "timestamp",
		input:     "2014-04-10T19:57:38.123456789-08:00",
		auto:      true,
		expected:  time.Unix(1397188658, 123456789),
	},
	{
		format:    time.RFC3339,
		fieldName: "Date",
		input:     "2014-04-10T19:57:38-08:00",
		auto:      true,
		expected:  time.Unix(1397188658, 0),
	},
	{
		format:    time.RFC3339,
		fieldName: "Date",
		input:     "2014-04-10T19:57:38Z",
		auto:      true,
		expected:  time.Unix(1397159858, 0),
	},
	{
		format:    time.RubyDate,
		fieldName: "datetime",
		input:     "Thu Apr 10 19:57:38.123456789 -0800 2014",
		auto:      true,
		expected:  time.Unix(1397188658, 123456789),
	},
	{
		format:    "%Y-%m-%d %H:%M",
		fieldName: "time",
		input:     "2014-07-30 07:02",
		expected:  time.Unix(1406703720, 0),
	},
	{
		format:    "%Y-%m-%d %k:%M", // check trailing space behavior
		fieldName: "time",
		input:     "2014-07-30  7:02",
		expected:  time.Unix(1406703720, 0),
	},
	{
		format:    "%Y-%m-%d %H:%M:%S",
		fieldName: "time",
		input:     "2014-07-30 07:02:15",
		expected:  time.Unix(1406703735, 0),
	},
	{
		format:    UnixTimestampFmt,
		fieldName: "time",
		input:     "1440116565",
		expected:  time.Unix(1440116565, 0),
	},
	{
		format:    UnixTimestampFmt,
		fieldName: "time",
		input:     1440116565,
		expected:  time.Unix(1440116565, 0),
	},
	{
		format:    "%Y-%m-%d %z",
		input:     "2014-04-10 -0700",
		fieldName: "time",
		expected:  time.Unix(1397113200, 0),
	},
}

func TestGetTimestampValid(t *testing.T) {
	p := &Parser{nower: &FakeNower{}}
	for i, tTimeSet := range tts {
		if tTimeSet.auto {
			p.conf = Options{}
			resp := p.getTimestamp(map[string]interface{}{tTimeSet.fieldName: tTimeSet.input})
			if !resp.Equal(tTimeSet.expected) {
				t.Errorf("time %d: should've been parsed automatically, without required config", i)
			}
		}

		p.conf = Options{TimeFieldName: tTimeSet.fieldName, Format: tTimeSet.format}
		resp := p.getTimestamp(map[string]interface{}{tTimeSet.fieldName: tTimeSet.input})
		if !resp.Equal(tTimeSet.expected) {
			t.Errorf("time %d: resp time %s didn't match expected time %s", i, resp, tTimeSet.expected)
		}
	}
}

func TestGetTimestampInvalid(t *testing.T) {
	p := &Parser{nower: &FakeNower{}}
	// time field missing
	resp := p.getTimestamp(map[string]interface{}{"noTimeField": "not used"})
	if !resp.Equal(p.nower.Now()) {
		t.Errorf("resp time %s didn't match expected time %s", resp, p.nower.Now())
	}
	// time field unparsable
	resp = p.getTimestamp(map[string]interface{}{"time": "not a valid date"})
	if !resp.Equal(p.nower.Now()) {
		t.Errorf("resp time %s didn't match expected time %s", resp, p.nower.Now())
	}
}

func TestGetTimestampCustomFormat(t *testing.T) {
	weirdFormat := "Mon // 02 ---- Jan ... 06 15:04:05 -0700"

	testStr := "Mon // 09 ---- Aug ... 10 15:34:56 -0800"
	expected := time.Date(2010, 8, 9, 15, 34, 56, 0, time.FixedZone("PST", -28800))

	// with just Format defined
	p := &Parser{
		nower: &FakeNower{},
		conf:  Options{Format: weirdFormat},
	}
	resp := p.getTimestamp(map[string]interface{}{"timestamp": testStr})
	if !resp.Equal(expected) {
		t.Errorf("resp time %s didn't match expected time %s", resp, expected)
	}

	// with just TimeFieldName defined
	p = &Parser{
		nower: &FakeNower{},
		conf: Options{
			TimeFieldName: "funkyTime",
		},
	}
	// use one of the expected/fallback formats
	resp = p.getTimestamp(map[string]interface{}{"funkyTime": expected.Format(time.RubyDate)})
	if !resp.Equal(expected) {
		t.Errorf("resp time %s didn't match expected time %s", resp, expected)
	}

	// Now with both defined
	p = &Parser{
		nower: &FakeNower{},
		conf: Options{
			TimeFieldName: "funkyTime",
			Format:        weirdFormat,
		},
	}
	resp = p.getTimestamp(map[string]interface{}{"funkyTime": testStr})
	if !resp.Equal(expected) {
		t.Errorf("resp time %s didn't match expected time %s", resp, expected)
	}
	// don't parse the "time" field if we're told to look for time in "funkyTime"
	resp = p.getTimestamp(map[string]interface{}{"time": "2014-04-10 19:57:38.123456789 -0800 PST"})
	if !resp.Equal(p.nower.Now()) {
		t.Errorf("resp time %s didn't match expected time %s", resp, p.nower.Now())
	}
}

func TestCommaInTimestamp(t *testing.T) {
	p := &Parser{
		nower: &FakeNower{},
		conf:  Options{},
	}
	commaTimes := []testTimestamp{
		{ // test commas as the fractional portion separator
			format:    "2006-01-02 15:04:05,999999999 -0700 MST",
			fieldName: "time",
			input:     "2014-03-10 12:57:38,123456789 -0700 PDT",
			expected:  time.Unix(1394481458, 123456789),
		},
		{
			format:    "2006-01-02 15:04:05.999999999 -0700 MST",
			fieldName: "time",
			input:     "2014-03-10 12:57:38,123456789 -0700 PDT",
			expected:  time.Unix(1394481458, 123456789),
		},
	}
	for i, tTimeSet := range commaTimes {
		p.conf.Format = tTimeSet.format
		expectedTime := tTimeSet.expected
		resp := p.getTimestamp(map[string]interface{}{tTimeSet.fieldName: tTimeSet.input})
		if !resp.Equal(expectedTime) {
			t.Errorf("time %d: resp time %s didn't match expected time %s", i, resp, expectedTime)
		}
	}

}
