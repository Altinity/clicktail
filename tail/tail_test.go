package tail

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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
	lines := tailSingleFile(tailer, filename, statefilename)
	checkLinesChan(t, lines, jsonLines)
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
			jsonLines[i][j] = fmt.Sprintf("{\"a\":%d", i)
		}

		filename := filenameRoot + fmt.Sprint(i)
		conf.Paths[i] = filename
		ts.writeFile(t, filename, strings.Join(jsonLines[i], "\n"))
	}

	chanArr, err := GetEntries(conf)
	if err != nil {
		t.Fatal(err)
	}
	for i, ch := range chanArr {
		checkLinesChan(t, ch, jsonLines[i])
	}
}

// boilerplate to spin up a httptest server, create tmpdir, etc.
// to create an environment in which to run these tests
type testSetup struct {
	tmpdir string
}

func (ts *testSetup) start(t *testing.T) {
	logrus.SetOutput(ioutil.Discard)
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	ts.tmpdir = tmpdir
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
