package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/libhoney-go"
	flag "github.com/jessevdk/go-flags"

	"github.com/honeycombio/honeytail/parsers/arangodb"
	"github.com/honeycombio/honeytail/parsers/htjson"
	"github.com/honeycombio/honeytail/parsers/mongodb"
	"github.com/honeycombio/honeytail/parsers/mysql"
	"github.com/honeycombio/honeytail/parsers/nginx"
	"github.com/honeycombio/honeytail/tail"
)

// BuildID is set by Travis CI
var BuildID string

// internal version identifier
var version string

var validParsers = []string{
	"nginx",
	"mongo",
	"json",
	"mysql",
	"arangodb",
}

// GlobalOptions has all the top level CLI flags that honeytail supports
type GlobalOptions struct {
	APIHost    string `hidden:"true" long:"api_host" description:"Host for the Honeycomb API" default:"https://api.honeycomb.io/"`
	TailSample bool   `hidden:"true" description:"When true, sample while tailing. When false, sample post-parser events"`

	ConfigFile string `short:"c" long:"config" description:"Config file for honeytail in INI format." no-ini:"true"`

	SampleRate     uint `short:"r" long:"samplerate" description:"Only send 1 / N log lines" default:"1"`
	NumSenders     uint `short:"P" long:"poolsize" description:"Number of concurrent connections to open to Honeycomb" default:"10"`
	Debug          bool `long:"debug" description:"Print debugging output"`
	StatusInterval uint `long:"status_interval" description:"How frequently, in seconds, to print out summary info" default:"60"`
	Backfill       bool `long:"backfill" description:"Configure honeytail to ingest old data in order to backfill Honeycomb. Sets the correct values for --backoff, --tail.read_from, and --tail.stop"`

	ScrubFields       []string `long:"scrub_field" description:"For the field listed, apply a one-way hash to the field content. May be specified multiple times"`
	DropFields        []string `long:"drop_field" description:"Do not send the field to Honeycomb. May be specified multiple times"`
	AddFields         []string `long:"add_field" description:"Add the field to every event. Field should be key=val. May be specified multiple times"`
	RequestShape      []string `long:"request_shape" description:"Identify a field that contains an HTTP request of the form 'METHOD /path HTTP/1.x' or just the request path. Break apart that field into subfields that contain components. May be specified multiple times. Defaults to 'request' when using the nginx parser"`
	ShapePrefix       string   `long:"shape_prefix" description:"Prefix to use on fields generated from request_shape to prevent field collision"`
	RequestPattern    []string `long:"request_pattern" description:"A pattern for the request path on which to base the derived request_shape. May be specified multiple times. Patterns are considered in order; first match wins."`
	RequestParseQuery string   `long:"request_parse_query" description:"How to parse the request query parameters. 'whitelist' means only extract listed query keys. 'all' means to extract all query parameters as individual columns" default:"whitelist"`
	RequestQueryKeys  []string `long:"request_query_keys" description:"Request query parameter key names to extract, when request_parse_query is 'whitelist'. May be specified multiple times."`
	BackOff           bool     `long:"backoff" description:"When rate limited by the API, back off and retry sending failed events. Otherwise failed events are dropped. When --backfill is set, it will override this option=true"`
	PrefixRegex       string   `long:"log_prefix" description:"pass a regex to this flag to strip the matching prefix from the line before handing to the parser. Useful when log aggregation prepends a line header. Use named groups to extract fields into the event."`

	Reqs  RequiredOptions `group:"Required Options"`
	Modes OtherModes      `group:"Other Modes"`

	Tail tail.TailOptions `group:"Tail Options" namespace:"tail"`

	Nginx    nginx.Options    `group:"Nginx Parser Options" namespace:"nginx"`
	JSON     htjson.Options   `group:"JSON Parser Options" namespace:"json"`
	MySQL    mysql.Options    `group:"MySQL Parser Options" namespace:"mysql"`
	Mongo    mongodb.Options  `group:"MongoDB Parser Options" namespace:"mongo"`
	ArangoDB arangodb.Options `group:"ArangoDB Parser Options" namespace:"arangodb"`
}

type RequiredOptions struct {
	ParserName string   `short:"p" long:"parser" description:"Parser module to use. Use --list to list available options."`
	WriteKey   string   `short:"k" long:"writekey" description:"Team write key"`
	LogFiles   []string `short:"f" long:"file" description:"Log file(s) to parse. Use '-' for STDIN, use this flag multiple times to tail multiple files, or use a glob (/path/to/foo-*.log)"`
	Dataset    string   `short:"d" long:"dataset" description:"Name of the dataset"`
}

type OtherModes struct {
	Help               bool `short:"h" long:"help" description:"Show this help message"`
	ListParsers        bool `short:"l" long:"list" description:"List available parsers"`
	Version            bool `short:"V" long:"version" description:"Show version"`
	WriteDefaultConfig bool `long:"write_default_config" description:"Write a default config file to STDOUT" no-ini:"true"`
	WriteCurrentConfig bool `long:"write_current_config" description:"Write out the current config to STDOUT" no-ini:"true"`

	WriteManPage bool `hidden:"true" long:"write-man-page" description:"Write out a man page"`
}

func main() {
	var options GlobalOptions
	flagParser := flag.NewParser(&options, flag.PrintErrors)
	flagParser.Usage = "-p <parser> -k <writekey> -f </path/to/logfile> -d <mydata> [optional arguments]"

	if extraArgs, err := flagParser.Parse(); err != nil || len(extraArgs) != 0 {
		fmt.Println("Error: failed to parse the command line.")
		if err != nil {
			fmt.Printf("\t%s\n", err)
		} else {
			fmt.Printf("\tUnexpected extra arguments: %s\n", strings.Join(extraArgs, " "))
		}
		usage()
		os.Exit(1)
	}
	// read the config file if present
	if options.ConfigFile != "" {
		ini := flag.NewIniParser(flagParser)
		ini.ParseAsDefaults = true
		if err := ini.ParseFile(options.ConfigFile); err != nil {
			fmt.Printf("Error: failed to parse the config file %s\n", options.ConfigFile)
			fmt.Printf("\t%s\n", err)
			usage()
			os.Exit(1)
		}
	}

	rand.Seed(time.Now().UnixNano())

	if options.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Support flag alias: --backfill should cover --backoff --tail.read_from=beginning --tail.stop
	if options.Backfill {
		options.BackOff = true
		options.Tail.ReadFrom = "beginning"
		options.Tail.Stop = true
	}

	setVersionUserAgent(options.Backfill, options.Reqs.ParserName)
	handleOtherModes(flagParser, options.Modes)
	addParserDefaultOptions(&options)
	sanityCheckOptions(&options)

	verifyWritekey(options.APIHost, options.Reqs.WriteKey)
	run(options)
}

// setVersion sets the internal version ID and updates libhoney's user-agent
func setVersionUserAgent(backfill bool, parserName string) {
	if BuildID == "" {
		version = "dev"
	} else {
		version = BuildID
	}
	if backfill {
		parserName += " backfill"
	}
	libhoney.UserAgentAddition = fmt.Sprintf("honeytail/%s (%s)", version, parserName)
}

// handleOtherModes takse care of all flags that say we should just do something
// and exit rather than actually parsing logs
func handleOtherModes(fp *flag.Parser, modes OtherModes) {
	if modes.Version {
		fmt.Println("Honeytail version", version)
		os.Exit(0)
	}
	if modes.Help {
		fp.WriteHelp(os.Stdout)
		fmt.Println("")
		os.Exit(0)
	}
	if modes.WriteManPage {
		fp.WriteManPage(os.Stdout)
		os.Exit(0)
	}
	if modes.WriteDefaultConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeDefaults|flag.IniCommentDefaults|flag.IniIncludeComments)
		os.Exit(0)
	}
	if modes.WriteCurrentConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeComments)
		os.Exit(0)
	}

	if modes.ListParsers {
		fmt.Println("Available parsers:", strings.Join(validParsers, ", "))
		os.Exit(0)
	}
}

func addParserDefaultOptions(options *GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "nginx":
		// automatically normalize the request when using the nginx parser
		options.RequestShape = append(options.RequestShape, "request")
	}
	if options.Reqs.ParserName != "mysql" {
		// mysql is the only parser that requires post-parsed sampling.
		// Sample all other parser when tailing to conserve CPU
		options.TailSample = true
	} else {
		options.TailSample = false
	}
}

func sanityCheckOptions(options *GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "":
		fmt.Println("Parser required.")
		usage()
		os.Exit(1)
	case options.Reqs.WriteKey == "" || options.Reqs.WriteKey == "NULL":
		fmt.Println("Write key required.")
		usage()
		os.Exit(1)
	case len(options.Reqs.LogFiles) == 0:
		fmt.Println("Log file name or '-' required.")
		usage()
		os.Exit(1)
	case options.Reqs.Dataset == "":
		fmt.Println("Dataset name required.")
		usage()
		os.Exit(1)
	case options.SampleRate == 0:
		fmt.Println("Sample rate must be an integer >= 1")
		usage()
		os.Exit(1)
	case options.Tail.ReadFrom == "end" && options.Tail.Stop:
		fmt.Println("Reading from the end and stopping when we get there. Zero lines to process. Ok, all done! ;)")
		usage()
		os.Exit(1)
	case options.RequestParseQuery != "whitelist" && options.RequestParseQuery != "all":
		fmt.Println("request_parse_query flag must be either 'whitelist' or 'all'.")
		usage()
		os.Exit(1)
	}

	// check the prefix regex for validity
	if options.PrefixRegex != "" {
		// make sure the regex is anchored against the start of the string
		if options.PrefixRegex[0] != '^' {
			options.PrefixRegex = "^" + options.PrefixRegex
		}
		// make sure it's valid
		_, err := regexp.Compile(options.PrefixRegex)
		if err != nil {
			fmt.Printf("Prefix regex %s doesn't compile: error %s\n", options.PrefixRegex, err)
			usage()
			os.Exit(1)
		}
	}

	// Make sure input files exist
	shouldExit := false
	for _, f := range options.Reqs.LogFiles {
		if f == "-" {
			continue
		}
		if files, err := filepath.Glob(f); err != nil || files == nil {
			fmt.Printf("Log file specified by --file=%s not found!\n", f)
			shouldExit = true
		}
	}
	if shouldExit {
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`
Usage: honeytail -p <parser> -k <writekey> -f </path/to/logfile> -d <mydata> [optional arguments]

For even more detail on required and optional parameters, run
honeytail --help
`)
}

// verifyWritekey calls out to api to validate the writekey, so we can exit
// immediately instead of happily sending events that are all rejected.
func verifyWritekey(apiHost string, writeKey string) {
	url := fmt.Sprintf("%s/1/team_slug", apiHost)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", libhoney.UserAgentAddition)
	req.Header.Add("X-Honeycomb-Team", writeKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to validate your writekey:")
		fmt.Println("\t", err)
		fmt.Println("Sorry! Please try again.")
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Failed to validate your writekey:")
		fmt.Println("\t", string(body))
		fmt.Println("Sorry! Please try again.")
		os.Exit(1)
	}
}
