package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/tail"
)

var tailOptions = tail.TailOptions{
	ReadFrom: "start",
	Stop:     true,
}

// defaultOptions is a fully populated GlobalOptions with good defaults to start from
var defaultOptions = GlobalOptions{
	// each test will have to populate APIHost with the location of its test server
	APIHost:          "",
	SampleRate:       1,
	NumSenders:       1,
	BatchFrequencyMs: 1000, // Longer batch sends to accommodate for slower CI machines
	Reqs: RequiredOptions{
		// using the json parser for everything because we're not testing parsers here.
		ParserName: "json",
		WriteKey:   "abcabc123123",
		// each test will specify its own logfile
		// LogFiles:   []string{tmpdir + ""},
		Dataset: "pika",
	},
	Tail:           tailOptions,
	StatusInterval: 1,
}

// test testing framework
func TestHTTPtest(t *testing.T) {
	ts := &testSetup{}
	ts.start(t, &GlobalOptions{})
	defer ts.close()
	ts.rsp.responseBody = "whatup pikachu"
	res, err := http.Get(ts.server.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, string(greeting), "whatup pikachu")
	assert.Equal(t, ts.rsp.reqCounter, 1)
}

func TestBasicSend(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/first.log"
	fh, err := os.Create(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	fmt.Fprintf(fh, `{"format":"json"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Equal(t, ts.rsp.evtCounter, 1)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json"}`)
	teamID := ts.rsp.req.Header.Get("X-Honeycomb-Team")
	assert.Equal(t, teamID, "abcabc123123")
	requestURL := ts.rsp.req.URL.Path
	assert.Equal(t, requestURL, "/1/batch/pika")
}

func TestMultipleFiles(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFile1 := ts.tmpdir + "/first.log"
	fh1, err := os.Create(logFile1)
	if err != nil {
		t.Fatal(err)
	}
	logFile2 := ts.tmpdir + "/second.log"
	fh2, err := os.Create(logFile2)
	if err != nil {
		t.Fatal(err)
	}
	defer fh1.Close()
	fmt.Fprintf(fh1, `{"key1":"val1"}`)
	defer fh2.Close()
	fmt.Fprintf(fh2, `{"key2":"val2"}`)
	opts.Reqs.LogFiles = []string{logFile1, logFile2}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Equal(t, ts.rsp.evtCounter, 2)
	assert.Contains(t, ts.rsp.reqBody, `{"key1":"val1"}`)
	assert.Contains(t, ts.rsp.reqBody, `{"key2":"val2"}`)
	teamID := ts.rsp.req.Header.Get("X-Honeycomb-Team")
	assert.Equal(t, teamID, "abcabc123123")
	requestURL := ts.rsp.req.URL.Path
	assert.Equal(t, requestURL, "/1/batch/pika")
}

func TestMultiLineMultiFile(t *testing.T) {
	opts := GlobalOptions{
		NumSenders: 1,
		Reqs: RequiredOptions{
			ParserName: "mysql",
			WriteKey:   "----",
			Dataset:    "---",
		},
		Tail: tailOptions,
	}
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFile1 := ts.tmpdir + "/first.log"
	fh1, err := os.Create(logFile1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(fh1, `# Time: 2016-04-01T00:29:09.817887Z",
# administrator command: Close stmt;
# User@Host: root[root] @  [10.0.72.76]  Id: 432399
# Query_time: 0.000114  Lock_time: 0.000000 Rows_sent: 0  Rows_examined: 0
SET timestamp=1459470669;
SELECT *
FROM orders WHERE
total > 1000;
# Time: 2016-04-01T00:31:09.817887Z
SET timestamp=1459470669;
show status like 'Uptime';`)
	logFile2 := ts.tmpdir + "/second.log"
	fh2, err := os.Create(logFile2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintf(fh2, `# User@Host: rails[rails] @  [10.252.9.33]
# Query_time: 0.002280  Lock_time: 0.000023 Rows_sent: 0  Rows_examined: 921
SET timestamp=1444264264;
SELECT certs.* FROM certs WHERE (certs.app_id = 993089) LIMIT 1;
# administrator command: Prepare;
# User@Host: root[root] @  [10.0.99.122]  Id: 432407
# Query_time: 0.000122  Lock_time: 0.000033 Rows_sent: 1  Rows_examined: 1
SET timestamp=1476702000;
SELECT
                  id, team_id, name, description, slug, limit_kb, created_at, updated_at
                FROM datasets WHERE team_id=17 AND slug='api-prod';`)
	opts.Reqs.LogFiles = []string{logFile1, logFile2}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Equal(t, ts.rsp.evtCounter, 4)
	assert.Contains(t, ts.rsp.reqBody, `"query":"SELECT * FROM orders`)
	assert.Contains(t, ts.rsp.reqBody, `"tables":"orders"`)
	assert.Contains(t, ts.rsp.reqBody, `"query":"show status like 'Uptime'`)
	assert.Contains(t, ts.rsp.reqBody, `"query":"SELECT certs.* FROM`)
	assert.Contains(t, ts.rsp.reqBody, `"tables":"certs"`)
	assert.Contains(t, ts.rsp.reqBody, `"query":"SELECT id, team_id, name`)
	assert.Contains(t, ts.rsp.reqBody, `"tables":"datasets"`)
}

func TestSetVersion(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/setv.log"
	fh, _ := os.Create(logFileName)
	defer fh.Close()
	fmt.Fprintf(fh, `{"format":"json"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	run(opts)
	userAgent := ts.rsp.req.Header.Get("User-Agent")
	assert.Contains(t, userAgent, "libhoney-go")
	setVersionUserAgent(false, "fancyParser")
	run(opts)
	userAgent = ts.rsp.req.Header.Get("User-Agent")
	assert.Contains(t, userAgent, "libhoney-go")
	assert.Contains(t, userAgent, "fancyParser")
	BuildID = "test"
	setVersionUserAgent(false, "fancyParser")
	run(opts)
	userAgent = ts.rsp.req.Header.Get("User-Agent")
	assert.Contains(t, userAgent, " honeytail/test")
	setVersionUserAgent(true, "fancyParser")
	run(opts)
	userAgent = ts.rsp.req.Header.Get("User-Agent")
	assert.Contains(t, userAgent, " honeytail/test")
	assert.Contains(t, userAgent, "fancyParser backfill")
}

func TestDropField(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/drop.log"
	fh, _ := os.Create(logFileName)
	defer fh.Close()
	fmt.Fprintf(fh, `{"dropme":"chew","format":"json","reallygone":"notyet"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Contains(t, ts.rsp.reqBody, `{"dropme":"chew","format":"json","reallygone":"notyet"}`)
	opts.DropFields = []string{"dropme"}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 2)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json","reallygone":"notyet"}`)
	opts.DropFields = []string{"dropme", "reallygone"}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 3)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json"}`)
}

func TestScrubField(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/scrub.log"
	fh, _ := os.Create(logFileName)
	defer fh.Close()
	fmt.Fprintf(fh, `{"format":"json","name":"hidden"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	opts.ScrubFields = []string{"name"}
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json","name":"e564b4081d7a9ea4b00dada53bdae70c99b87b6fce869f0c3dd4d2bfa1e53e1c"}`)
}

func TestAddField(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/add.log"
	logfh, _ := os.Create(logFileName)
	defer logfh.Close()
	fmt.Fprintf(logfh, `{"format":"json"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	opts.AddFields = []string{`newfield=newval`}
	run(opts)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json","newfield":"newval"}`)
	opts.AddFields = []string{"newfield=newval", "second=new"}
	run(opts)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json","newfield":"newval","second":"new"}`)
}

func TestLinePrefix(t *testing.T) {
	opts := defaultOptions
	// linePrefix of "Nov 13 10:19:31 app23 process.port[pid]: "
	// let's capture timestamp and hostname, skip process.port and pid
	opts.PrefixRegex = `(?P<server_timestamp>... .. ..:..:..) (?P<hostname>[a-zA-Z0-9]+) [^:]*: `
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	logFileName := ts.tmpdir + "/linePrefix.log"
	logfh, _ := os.Create(logFileName)
	defer logfh.Close()
	fmt.Fprintf(logfh, `Nov 13 10:19:31 app23 process.port[pid]: {"format":"json"}`)
	opts.Reqs.LogFiles = []string{logFileName}
	run(opts)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json","hostname":"app23","server_timestamp":"Nov 13 10:19:31"}`)
}

func TestRequestShapeRaw(t *testing.T) {
	reqField := "request"
	opts := defaultOptions
	opts.RequestShape = []string{"request"}
	opts.RequestPattern = []string{"/about", "/about/:lang", "/about/:lang/books"}
	urlsWhitelistQuery := map[string]map[string]interface{}{
		"GET /about/en/books HTTP/1.1": {
			"request_method":           "GET",
			"request_protocol_version": "HTTP/1.1",
			"request_uri":              "/about/en/books",
			"request_path":             "/about/en/books",
			"request_query":            nil, // field missing instead of empty
			"request_path_lang":        "en",
			"request_shape":            "/about/:lang/books",
			"request_pathshape":        "/about/:lang/books",
			"request_queryshape":       nil, // field missing instead of empty
		},
		"GET /about?foo=bar HTTP/1.0": {
			"request_method":           "GET",
			"request_protocol_version": "HTTP/1.0",
			"request_uri":              "/about?foo=bar",
			"request_path":             "/about",
			"request_query":            "foo=bar",
			"request_query_foo":        "bar",
			"request_shape":            "/about?foo=?",
			"request_pathshape":        "/about",
			"request_queryshape":       "foo=?",
		},
		"/about/en/books": {
			"request_uri":        "/about/en/books",
			"request_path":       "/about/en/books",
			"request_query":      nil, // field missing instead of empty
			"request_path_lang":  "en",
			"request_shape":      "/about/:lang/books",
			"request_pathshape":  "/about/:lang/books",
			"request_queryshape": nil, // field missing instead of empty
		},
		"/about/en?foo=bar&bar=bar2": {
			"request_uri":        "/about/en?foo=bar&bar=bar2",
			"request_path":       "/about/en",
			"request_query":      "foo=bar&bar=bar2",
			"request_query_foo":  "bar",
			"request_path_lang":  "en",
			"request_shape":      "/about/:lang?bar=?&foo=?",
			"request_pathshape":  "/about/:lang",
			"request_queryshape": "bar=?&foo=?",
		},
		"/about/en?foo=bar&baz&foo=bend&foo=alpha&bend=beta": {
			"request_uri":        "/about/en?foo=bar&baz&foo=bend&foo=alpha&bend=beta",
			"request_path":       "/about/en",
			"request_query":      "foo=bar&baz&foo=bend&foo=alpha&bend=beta",
			"request_query_foo":  "alpha, bar, bend",
			"request_query_bend": "beta",
			"request_path_lang":  "en",
			"request_shape":      "/about/:lang?baz=?&bend=?&foo=?&foo=?&foo=?",
			"request_pathshape":  "/about/:lang",
			"request_queryshape": "baz=?&bend=?&foo=?&foo=?&foo=?",
		},
	}
	urlsAllQuery := map[string]map[string]interface{}{
		"/about/en?foo=bar&bar=bar2": {
			"request_uri":        "/about/en?foo=bar&bar=bar2",
			"request_path":       "/about/en",
			"request_query":      "foo=bar&bar=bar2",
			"request_query_foo":  "bar",
			"request_query_bar":  "bar2",
			"request_path_lang":  "en",
			"request_shape":      "/about/:lang?bar=?&foo=?",
			"request_pathshape":  "/about/:lang",
			"request_queryshape": "bar=?&foo=?",
		},
	}
	// test whitelisting keys foo, baz, and bend but not bar
	opts.RequestQueryKeys = []string{"foo", "baz", "bend"}
	tbs := make(chan event.Event)
	output := modifyEventContents(tbs, opts)
	for input, expectedResult := range urlsWhitelistQuery {
		ev := event.Event{
			Data: map[string]interface{}{
				reqField: input,
			},
		}
		// feed it the sample event
		tbs <- ev
		// get the munged event out
		res := <-output
		for evKey, expectedVal := range expectedResult {
			assert.Equal(t, res.Data[evKey], expectedVal)
		}
	}
	close(tbs)

	// change the query parsing rules and get a new output channel - bar should be
	// included
	opts.RequestParseQuery = "all"
	tbs = make(chan event.Event)
	output = modifyEventContents(tbs, opts)
	for input, expectedResult := range urlsAllQuery {
		ev := event.Event{
			Data: map[string]interface{}{
				reqField: input,
			},
		}
		// feed it the sample event
		tbs <- ev
		// get the munged event out
		res := <-output
		for evKey, expectedVal := range expectedResult {
			assert.Equal(t, res.Data[evKey], expectedVal)
		}
	}
	close(tbs)
}

func TestSampleRate(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	rand.Seed(1)
	sampleLogFile := ts.tmpdir + "/sample.log"
	logfh, _ := os.Create(sampleLogFile)
	defer logfh.Close()
	for i := 0; i < 50; i++ {
		fmt.Fprintf(logfh, `{"format":"json%d"}`+"\n", i)
	}
	opts.Reqs.LogFiles = []string{sampleLogFile}
	opts.TailSample = false

	run(opts)
	// with no sampling, 50 lines -> 50 events
	assert.Equal(t, ts.rsp.evtCounter, 50)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json49"}`)
	ts.rsp.reset()

	opts.SampleRate = 3
	opts.TailSample = true
	run(opts)
	// setting a sample rate of 3 gets 17 requests.
	// tail does the sampling
	assert.Equal(t, ts.rsp.evtCounter, 17)
	assert.Contains(t, ts.rsp.reqBody, `{"format":"json49"},"samplerate":3,`)
}

func TestReadFromOffset(t *testing.T) {
	opts := defaultOptions
	ts := &testSetup{}
	ts.start(t, &opts)
	defer ts.close()
	offsetLogFile := ts.tmpdir + "/offset.log"
	offsetStateFile := ts.tmpdir + "/offset.leash.state"
	logfh, _ := os.Create(offsetLogFile)
	defer logfh.Close()
	logStat := unix.Stat_t{}
	unix.Stat(offsetLogFile, &logStat)
	for i := 0; i < 10; i++ {
		fmt.Fprintf(logfh, `{"format":"json%d"}`+"\n", i)
	}
	opts.Reqs.LogFiles = []string{offsetLogFile}
	opts.Tail.ReadFrom = "last"
	opts.Tail.StateFile = offsetStateFile
	osf, _ := os.Create(offsetStateFile)
	defer osf.Close()
	fmt.Fprintf(osf, `{"INode":%d,"Offset":38}`, logStat.Ino)
	run(opts)
	assert.Equal(t, ts.rsp.reqCounter, 1)
	assert.Equal(t, ts.rsp.evtCounter, 8)
}

// boilerplate to spin up a httptest server, create tmpdir, etc.
// to create an environment in which to run these tests
type testSetup struct {
	server *httptest.Server
	rsp    *responder
	tmpdir string
}

func (t *testSetup) start(tst *testing.T, opts *GlobalOptions) {
	logrus.SetOutput(ioutil.Discard)
	t.rsp = &responder{}
	t.server = httptest.NewServer(http.HandlerFunc(t.rsp.serveResponse))
	tmpdir, err := ioutil.TempDir(os.TempDir(), "test")
	if err != nil {
		tst.Fatal(err)
	}
	t.tmpdir = tmpdir
	opts.APIHost = t.server.URL
	t.rsp.responseCode = 200
}
func (t *testSetup) close() {
	t.server.Close()
	os.RemoveAll(t.tmpdir)
}

type responder struct {
	req          *http.Request // the most recent request answered by the server
	reqBody      string        // the body sent along with the request
	reqCounter   int           // the number of requests answered since last reset
	evtCounter   int           // the number of events (<= reqCounter, will be < if events are batched)
	responseCode int           // the http status code with which to respond
	responseBody string        // the body to send as the response
}

func (r *responder) serveResponse(w http.ResponseWriter, req *http.Request) {
	r.req = req
	r.reqCounter += 1

	var reader io.ReadCloser
	switch req.Header.Get("Content-Encoding") {
	case "gzip":
		buf := bytes.Buffer{}
		if _, err := io.Copy(&buf, req.Body); err != nil {
			panic(err)
		}
		gzReader, err := gzip.NewReader(&buf)
		if err != nil {
			panic(err)
		}
		req.Body.Close()
		reader = gzReader
	default:
		reader = req.Body
	}
	defer reader.Close()

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	payload := []map[string]interface{}{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			r.evtCounter++ // likely not a batch request
		} else {
			r.evtCounter += len(payload)
		}
	}
	r.reqBody = string(body)
	w.WriteHeader(r.responseCode)
	fmt.Fprintf(w, r.responseBody)
}
func (r *responder) reset() {
	r.reqCounter = 0
	r.evtCounter = 0
	r.responseCode = 200
}
