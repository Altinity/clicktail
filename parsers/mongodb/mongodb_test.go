package mongodb

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
)

const commonLogFormatTimeLayout = "02/Jan/2006:15:04:05 -0700"

type testLineMaps struct {
	line string
	resp map[string]interface{}
	ev   event.Event
}

type FakeLineParser struct {
	tlm []testLineMaps
}

func (f *FakeLineParser) ParseLogLine(line string) (map[string]interface{}, error) {
	for _, lm := range f.tlm {
		if line == lm.line {
			return lm.resp, nil
		}
	}
	return nil, errors.New("line not found in test slice")
}

func TestProcessLines(t *testing.T) {
	rand.Seed(5) // set a fixed seed so the tests are predictable with randomTime
	t1, _ := time.Parse(commonLogFormatTimeLayout, "28/Dec/2009:01:38:56 +0000")
	tlm := []testLineMaps{
		{
			line: "test case 1",
			resp: map[string]interface{}{
				"query": "foobar",
				"key2":  "val2",
			},
			ev: event.Event{
				Timestamp: t1,
				Data: map[string]interface{}{
					"query": "foobar",
					"key2":  "val2",
				},
			},
		},
	}
	m := &Parser{
		conf:  Options{},
		nower: &FakeNower{},
		lineParser: &FakeLineParser{
			tlm: tlm,
		},
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
	go m.ProcessLines(lines, send)
	for _, pair := range tlm {
		resp := <-send
		if !reflect.DeepEqual(resp, pair.ev) {
			t.Fatalf("line resp didn't match up for %s. Expected %+v, actual: %+v",
				// pair.line, string(pair.ev.Blob), string(resp.Blob))
				pair.line, pair.ev, resp)
		}
	}
}

type FakeNower struct{}

func (f *FakeNower) Now() time.Time {
	fakeTime, _ := time.Parse(commonLogFormatTimeLayout, "02/Jan/2010:12:34:56 -0000")
	return fakeTime
}
