package tail

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
)

var tailOpts = TailOptions{
	ReadFrom: "start",
	Stop:     true,
}

func TestTailSingleFile(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()

	filename := ts.tmpdir + "/first.log"
	statefilename := filename + ".mystate"
	jsonLines := []string{"{\"a\":1}", "{\"b\":2}", "{\"c\":3}"}
	ts.writeFile(t, filename, strings.Join(jsonLines, "\n"))

	conf := Config{
		Options: tailOpts,
	}
	tailer, err := getTailer(conf, filename, statefilename)
	if err != nil {
		t.Fatal(err)
	}
	lines := tailSingleFile(ts.ctx, tailer, filename, statefilename)
	checkLinesChan(t, lines, jsonLines)
}

func TestTailSTDIN(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()
	conf := Config{
		Options: tailOpts,
		Paths:   make([]string, 1),
	}
	conf.Paths[0] = "-"
	lineChans, err := GetEntries(ts.ctx, conf)
	if err != nil {
		t.Fatal(err)
	}
	if len(lineChans) != 1 {
		t.Errorf("lines chans should have had one channel; instead was length %d", len(lineChans))
	}
}

func TestGetSampledEntries(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()
	rand.Seed(3)

	conf := Config{
		Paths:   make([]string, 3),
		Options: tailOpts,
	}

	jsonLines := make([][]string, 3)
	filenameRoot := ts.tmpdir + "/json.log"
	for i := 0; i < 3; i++ {
		jsonLines[i] = make([]string, 6)
		for j := 0; j < 6; j++ {
			jsonLines[i][j] = fmt.Sprintf("{\"a\":%d", i)
		}

		filename := filenameRoot + fmt.Sprint(i)
		conf.Paths[i] = filename
		ts.writeFile(t, filename, strings.Join(jsonLines[i], "\n"))
	}

	chanArr, err := GetSampledEntries(ts.ctx, conf, 2)
	if err != nil {
		t.Fatal(err)
	}
	// can't check each line because the parallel goroutines screw with the random
	// dropping lines, so you can't know which channel will drop which messages.
	// But the overall count of messages is predictable.
	var lineCounter int

	for _, ch := range chanArr {
		for _ = range ch {
			lineCounter++
		}
	}
	expectedLines := 10
	if lineCounter != expectedLines {
		t.Errorf("expected to get %d lines, got %d instead", expectedLines, lineCounter)
	}
}

func TestGetEntries(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()

	conf := Config{
		Paths:   make([]string, 3),
		Options: tailOpts,
	}

	jsonLines := make([][]string, 3)
	filenameRoot := ts.tmpdir + "/json.log"
	for i := 0; i < 3; i++ {
		jsonLines[i] = make([]string, 3)
		for j := 0; j < 3; j++ {
			jsonLines[i][j] = fmt.Sprintf("{\"a\":%d}", i)
		}

		filename := filenameRoot + fmt.Sprint(i)
		conf.Paths[i] = filename
		ts.writeFile(t, filename, strings.Join(jsonLines[i], "\n"))
	}

	chanArr, err := GetEntries(ts.ctx, conf)
	if err != nil {
		t.Fatal(err)
	}
	for i, ch := range chanArr {
		checkLinesChan(t, ch, jsonLines[i])
	}

	// test that if all statefile-like filenames and missing files are removed
	// from the list, it errors
	fn1 := ts.tmpdir + "/sparklestate"
	ts.writeFile(t, fn1, "body")
	fn2 := ts.tmpdir + "/foo.leash.state"
	ts.writeFile(t, fn2, "body")
	conf = Config{
		Paths: []string{fn1, fn2, "/file/does/not/exist"},
		Options: TailOptions{
			StateFile: fn1,
		},
	}
	nilChan, err := GetEntries(ts.ctx, conf)
	if nilChan != nil {
		t.Error("errored getEntries was supposed to respond with a nil channel list")
	}
	if err == nil {
		t.Error("expected error from GetEntries; got nil instead.")
	}
}

func TestAbortChannel(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()

	var tailWait = TailOptions{
		ReadFrom: "start",
		Stop:     false,
	}

	conf := Config{
		Paths:   make([]string, 3),
		Options: tailWait,
	}

	jsonLines := make([][]string, 3)
	filenameRoot := ts.tmpdir + "/json.log"
	for i := 0; i < 3; i++ {
		jsonLines[i] = make([]string, 3)
		for j := 0; j < 3; j++ {
			jsonLines[i][j] = fmt.Sprintf("{\"a\":%d}", i)
		}

		filename := filenameRoot + fmt.Sprint(i)
		conf.Paths[i] = filename
		ts.writeFile(t, filename, strings.Join(jsonLines[i], "\n"))
	}

	chanArr, err := GetEntries(ts.ctx, conf)
	if err != nil {
		t.Fatal(err)
	}

	// ok, let's see what happens when we want to quit
	ts.cancel()
	for _, ch := range chanArr {
		checkLinesChanClosed(t, ch)
	}
}

func TestRemoveStateFiles(t *testing.T) {
	files := []string{
		"foo.bar",
		"/bar.baz",
		"bar.leash.state",
		"myspecialstatefile",
		"baz.foo",
	}
	expectedFilesNoStatefile := []string{
		"foo.bar",
		"/bar.baz",
		"myspecialstatefile",
		"baz.foo",
	}
	expectedFilesConfStatefile := []string{
		"foo.bar",
		"/bar.baz",
		"baz.foo",
	}
	conf := Config{
		Options: TailOptions{},
	}
	newFiles := removeStateFiles(files, conf)
	if !reflect.DeepEqual(newFiles, expectedFilesNoStatefile) {
		t.Errorf("expected %v, instead got %v", expectedFilesNoStatefile, newFiles)
	}
	conf = Config{
		Options: TailOptions{
			StateFile: "myspecialstatefile",
		},
	}
	newFiles = removeStateFiles(files, conf)
	if !reflect.DeepEqual(newFiles, expectedFilesConfStatefile) {
		t.Errorf("expected %v, instead got %v", expectedFilesConfStatefile, newFiles)
	}
}

func TestGetStateFile(t *testing.T) {
	ts := &testSetup{}
	ts.start(t)
	defer ts.stop()

	conf := Config{
		Paths:   make([]string, 3),
		Options: tailOpts,
	}

	filename := "foobar.log"
	statefilename := "foobar.leash.state"

	existingStateFile := filepath.Join(ts.tmpdir, "existing.state")
	ts.writeFile(t, existingStateFile, "")
	newStateFile := filepath.Join(ts.tmpdir, "new.state")

	tsts := []struct {
		stateFileConfig string
		numFiles        int
		expected        string
	}{
		{existingStateFile, 1, existingStateFile},
		{existingStateFile, 2, filepath.Join(os.TempDir(), statefilename)},
		{newStateFile, 1, newStateFile},
		{newStateFile, 2, filepath.Join(os.TempDir(), statefilename)},
		{ts.tmpdir, 1, filepath.Join(ts.tmpdir, statefilename)},
		{ts.tmpdir, 2, filepath.Join(ts.tmpdir, statefilename)},
		{"", 1, filepath.Join(os.TempDir(), statefilename)},
		{"", 2, filepath.Join(os.TempDir(), statefilename)},
	}

	for _, tt := range tsts {
		conf.Options.StateFile = tt.stateFileConfig
		actual := getStateFile(conf, filename, tt.numFiles)
		if actual != tt.expected {
			t.Errorf("getStateFile with config statefile: %s\n\tgot: %s, expected: %s",
				tt.stateFileConfig, actual, tt.expected)
		}
	}
}

// boilerplate to spin up a httptest server, create tmpdir, etc.
// to create an environment in which to run these tests
type testSetup struct {
	tmpdir string
	ctx    context.Context
	cancel context.CancelFunc
}

func (ts *testSetup) start(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	ts.tmpdir = tmpdir
	ts.ctx, ts.cancel = context.WithCancel(context.Background())
}

func (ts *testSetup) writeFile(t *testing.T, path string, body string) {
	fh, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	fmt.Fprint(fh, body)
}

func (ts *testSetup) stop() {
	os.RemoveAll(ts.tmpdir)
}

func checkLinesChan(t *testing.T, actual chan string, expected []string) {
	idx := 0
	for line := range actual {
		if idx < len(expected) && expected[idx] != line {
			t.Errorf("got line '%s', expected line '%s'", line, expected[idx])
		}
		idx++
	}
	if idx != len(expected) {
		t.Errorf("read %d lines from lines channel; expected %d", idx, len(expected))
	}
}

func checkLinesChanClosed(t *testing.T, actual chan string) {
	// this will block if actual never gets closed
	for {
		select {
		case _, ok := <-actual:
			if !ok {
				return
			}
		case <-time.After(1 * time.Second):
			t.Error("channel read timed out; channel not closed")
			return
		}
	}
}
