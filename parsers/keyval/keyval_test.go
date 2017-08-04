package keyval

import (
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/honeycombio/honeytail/event"
)

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
		conf: Options{
			NumParsers: 5,
		},
		lineParser: &NoopLineParser{
			outgoingMap: map[string]interface{}{"key": "val"},
		},
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
		conf: Options{
			NumParsers: 5,
		},
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
