package htjson

import (
	"reflect"
	"testing"
)

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
