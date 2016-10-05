package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/libhoney-go"
	flag "github.com/jessevdk/go-flags"

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
}

// GlobalOptions has all the top level CLI flags that honeytail supports
type GlobalOptions struct {
	APIHost string `hidden:"true" long:"api_host" description:"Host for the Honeycomb API" default:"https://api.honeycomb.io/"`

	ConfigFile string `short:"c" long:"config" description:"config file for honeytail in INI format." no-ini:"true"`

	SampleRate     uint `short:"r" long:"samplerate" description:"Only send 1 / N log lines" default:"1"`
	NumSenders     uint `short:"P" long:"poolsize" description:"Number of concurrent connections to open to Honeycomb" default:"10"`
	Debug          bool `long:"debug" description:"Print debugging output"`
	StatusInterval uint `long:"status_interval" description:"how frequently, in seconds, to print out summary info" default:"60"`
	BackOff        bool `long:"backoff" description:"When rate limited by the API, back off and retry sending failed events. Otherwise failed events are dropped."`

	ScrubFields    []string `long:"scrub_field" description:"for the field listed, apply a one-way hash to the field content. May be specified multiple times"`
	DropFields     []string `long:"drop_field" description:"do not send the field to Honeycomb. May be specified multiple times"`
	AddFields      []string `long:"add_field" description:"add the field to every event. Field should be key=val. May be specified multiple times"`
	RequestShape   []string `long:"request_shape" description:"identify a field that contains an HTTP request of the form 'METHOD /path HTTP/1.x' or just the request path. Break apart that field into subfields that contain components. May be specified multiple times. Defaults to 'request' when using the nginx parser"`
	RequestPattern []string `long:"request_pattern" description:"a pattern for the request path on which to base the derived request_shape. May be specified multiple times. Patterns are considered in order; first match wins."`
	ShapePrefix    string   `long:"shape_prefix" description:"prefix to use on fields generated from request_shape to prevent field collision"`

	Reqs  RequiredOptions `group:"Required Options"`
	Modes OtherModes      `group:"Other Modes"`

	Tail tail.TailOptions `group:"Tail Options" namespace:"tail"`

	Nginx nginx.Options   `group:"Nginx Parser Options" namespace:"nginx"`
	JSON  htjson.Options  `group:"JSON Parser Options" namespace:"json"`
	MySQL mysql.Options   `group:"MySQL Parser Options" namespace:"mysql"`
	Mongo mongodb.Options `group:"MongoDB Parser Options" namespace:"mongo"`
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
	WriteDefaultConfig bool `long:"write_default_config" description:"write a default config file to STDOUT" no-ini:"true"`
	WriteCurrentConfig bool `long:"write_current_config" description:"write out the current config to STDOUT" no-ini:"true"`

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

	setVersion()
	handleOtherModes(flagParser, options)
	addParserDefaultOptions(&options)
	sanityCheckOptions(&options)

	verifyWritekey(options)
	run(options)
}

// setVersion sets the internal version ID and updates libhoney's user-agent
func setVersion() {
	if BuildID == "" {
		version = "dev"
	} else {
		version = BuildID
	}
	libhoney.UserAgentAddition = fmt.Sprintf("honeytail/%s", version)
}

// handleOtherModes takse care of all flags that say we should just do something
// and exit rather than actually parsing logs
func handleOtherModes(fp *flag.Parser, options GlobalOptions) {
	if options.Modes.Version {
		fmt.Println("Honeytail version", version)
		os.Exit(0)
	}
	if options.Modes.Help {
		fp.WriteHelp(os.Stdout)
		fmt.Println("")
		os.Exit(0)
	}
	if options.Modes.WriteManPage {
		fp.WriteManPage(os.Stdout)
		os.Exit(0)
	}
	if options.Modes.WriteDefaultConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeDefaults|flag.IniCommentDefaults|flag.IniIncludeComments)
		os.Exit(0)
	}
	if options.Modes.WriteCurrentConfig {
		ip := flag.NewIniParser(fp)
		ip.Write(os.Stdout, flag.IniIncludeComments)
		os.Exit(0)
	}

	if options.Modes.ListParsers {
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
}

func sanityCheckOptions(options *GlobalOptions) {
	switch {
	case options.Reqs.ParserName == "":
		fmt.Println("parser required.")
		usage()
		os.Exit(1)
	case options.Reqs.WriteKey == "" || options.Reqs.WriteKey == "NULL":
		fmt.Println("write key required.")
		usage()
		os.Exit(1)
	case len(options.Reqs.LogFiles) == 0:
		fmt.Println("log file name or '-' required.")
		usage()
		os.Exit(1)
	case options.Reqs.Dataset == "":
		fmt.Println("dataset name required.")
		usage()
		os.Exit(1)
	case options.Tail.ReadFrom == "end" && options.Tail.Stop:
		fmt.Println("Reading from the end and stopping when we get there. Zero lines to process. Ok, all done! ;)")
		usage()
		os.Exit(1)
	case len(options.Reqs.LogFiles) > 1 && options.Tail.StateFile != "":
		fmt.Println("Statefile can not be set when tailing from multiple files.")
		usage()
		os.Exit(1)
	case options.Tail.StateFile != "":
		files, err := filepath.Glob(options.Reqs.LogFiles[0])
		if err != nil {
			fmt.Printf("Trying to glob log file %s failed: %+v\n",
				options.Reqs.LogFiles[0], err)
			usage()
			os.Exit(1)
		}
		if len(files) > 1 {
			fmt.Println("Statefile can not be set when tailing from multiple files.")
			usage()
			os.Exit(1)
		}
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
func verifyWritekey(options GlobalOptions) {
	url := fmt.Sprintf("%s/1/team_slug", options.APIHost)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", libhoney.UserAgentAddition)
	req.Header.Add("X-Honeycomb-Team", options.Reqs.WriteKey)
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
