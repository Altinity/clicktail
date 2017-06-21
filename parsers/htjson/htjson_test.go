package htjson

import (
	"reflect"
	"testing"
	"time"
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
		input: `{"mystr": "myval", "myint": 3, "myfloat": 4.234}`,
		expected: map[string]interface{}{
			"mystr":   "myval",
			"myint":   float64(3),
			"myfloat": 4.234,
		},
	},
	{ // time
		input: `{"time": "2014-03-10 19:57:38.123456789 -0800 PST", "myint": 3, "myfloat": 4.234}`,
		expected: map[string]interface{}{
			"time":    "2014-03-10 19:57:38.123456789 -0800 PST",
			"myint":   float64(3),
			"myfloat": 4.234,
		},
	},
	{ // non-flat json object
		input: `{"array": [3, 4, 6], "obj": {"subkey":"subval"}, "myfloat": 4.234}`,
		expected: map[string]interface{}{
			"array":   []interface{}{float64(3), float64(4), float64(6)},
			"obj":     map[string]interface{}{"subkey": "subval"},
			"myfloat": 4.234,
		},
	},
}

func TestParseLine(t *testing.T) {
	jlp := JSONLineParser{}
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

type testTimestamp struct {
	format    string    // the format this test's time is in
	fieldName string    // the field in the map containing the time
	input     string    // the value corresponding to the fieldName
	auto      bool      // whether the input should be parsable even without specifying format/fieldName
	expected  time.Time // the expected time object to get back
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
		format:    "%Y-%m-%d %z",
		input:     "2014-04-10 -0700",
		fieldName: "time",
		expected:  time.Unix(1397113200, 0),
	},
	{
		format:    "%Y/%m/%d %H:%M:%S.%f %z",
		fieldName: "timestamp",
		input:     "2014/04/10 20:57:38.777456 -0700",
		expected:  time.Unix(1397188658, 777456000).UTC(),
	},
	{
		format:    "%Y/%m/%d %H:%M:%S.%f%z",
		fieldName: "timestamp",
		input:     "2014/04/10 20:57:38.789-0700",
		expected:  time.Unix(1397188658, 789000000),
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
