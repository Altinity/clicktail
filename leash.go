package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/urlshaper"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/arangodb"
	"github.com/honeycombio/honeytail/parsers/htjson"
	"github.com/honeycombio/honeytail/parsers/keyval"
	"github.com/honeycombio/honeytail/parsers/mongodb"
	"github.com/honeycombio/honeytail/parsers/mysql"
	"github.com/honeycombio/honeytail/parsers/nginx"
	"github.com/honeycombio/honeytail/tail"
)

// actually go and be leashy
func run(options GlobalOptions) {
	logrus.Info("Starting honeytail")

	sigs := make(chan os.Signal, 1)
	abort := make(chan struct{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// spin up our transmission to send events to Honeycomb
	libhConfig := libhoney.Config{
		WriteKey:             options.Reqs.WriteKey,
		Dataset:              options.Reqs.Dataset,
		SampleRate:           options.SampleRate,
		APIHost:              options.APIHost,
		MaxConcurrentBatches: options.NumSenders,
		SendFrequency:        time.Duration(options.BatchFrequencyMs) * time.Millisecond,
		// block on send should be true so if we can't send fast enough, we slow
		// down reading the log rather than drop lines.
		BlockOnSend: true,
		// block on response is true so that if we hit rate limiting we make sure
		// to re-enqueue all dropped events
		BlockOnResponse: true,

		// limit pending work capacity so that we get backpressure from libhoney
		// and block instead of sleeping inside sendToLibHoney.
		PendingWorkCapacity: 20 * options.NumSenders,
	}
	if err := libhoney.Init(libhConfig); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(
			"Error occured while spinning up Transimission")
	}

	// compile the prefix regex once for use on all channels
	var prefixRegex *parsers.ExtRegexp
	if options.PrefixRegex == "" {
		prefixRegex = nil
	} else {
		prefixRegex = &parsers.ExtRegexp{regexp.MustCompile(options.PrefixRegex)}
	}

	// get our lines channel from which to read log lines
	var linesChans []chan string
	var err error
	tc := tail.Config{
		Paths:   options.Reqs.LogFiles,
		Type:    tail.RotateStyleSyslog,
		Options: options.Tail,
	}
	if options.TailSample {
		linesChans, err = tail.GetSampledEntries(tc, options.SampleRate, abort)
	} else {
		linesChans, err = tail.GetEntries(tc, abort)
	}
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(
			"Error occurred while trying to tail logfile")
	}

	// set up our signal handler, now that we know how many files we're tailing,
	// we can send the right number of abort signals.
	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stderr, "Aborting! Caught signal \"%s\"\n", sig)
		fmt.Fprintf(os.Stderr, "Cleaning up...\n")
		close(abort)
		// and if they insist, catch a second CTRL-C or timeout on 10sec
		select {
		case <-sigs:
			fmt.Fprintf(os.Stderr, "Caught second signal... Aborting.\n")
			os.Exit(1)
		case <-time.After(10 * time.Second):
			fmt.Fprintf(os.Stderr, "Taking too long... Aborting.\n")
			os.Exit(1)
		}
	}()

	// for each channel we got back from tail.GetEntries, spin up a parser.
	parsersWG := sync.WaitGroup{}
	for _, lines := range linesChans {
		// get our parser
		parser, opts := getParserAndOptions(options)
		if parser == nil {
			logrus.WithFields(logrus.Fields{"parser": options.Reqs.ParserName}).Fatal(
				"Parser not found. Use --list to show valid parsers")
		}

		// and initialize it
		if err := parser.Init(opts); err != nil {
			logrus.WithFields(logrus.Fields{"parser": options.Reqs.ParserName, "err": err}).Fatal(
				"err initializing parser module")
		}

		// create a channel for sending events into libhoney
		toBeSent := make(chan event.Event, options.NumSenders)
		doneSending := make(chan bool)

		// two channels to handle backing off when rate limited and resending failed
		// send attempts that are recoverable
		toBeResent := make(chan event.Event, 2*options.NumSenders)
		// time in milliseconds to delay the send
		delaySending := make(chan int, 2*options.NumSenders)

		// apply any filters to the events before they get sent
		modifiedToBeSent := modifyEventContents(toBeSent, options)

		realToBeSent := make(chan event.Event, 10*options.NumSenders)
		go func() {
			wg := sync.WaitGroup{}
			for i := uint(0); i < options.NumSenders; i++ {
				wg.Add(1)
				go func() {
					for ev := range modifiedToBeSent {
						realToBeSent <- ev
					}
					wg.Done()
				}()
			}
			wg.Wait()
			close(realToBeSent)
		}()

		// start up the sender. all sources are either sampled when tailing or in-
		// parser, so always tell libhoney events are pre-sampled
		go sendToLibhoney(realToBeSent, toBeResent, delaySending, doneSending)

		// start a goroutine that reads from responses and logs.
		responses := libhoney.Responses()
		go handleResponses(responses, toBeResent, delaySending, options)

		parsersWG.Add(1)
		go func(plines chan string) {
			// ProcessLines won't return until lines is closed
			parser.ProcessLines(plines, toBeSent, prefixRegex)
			// trigger the sending goroutine to finish up
			close(toBeSent)
			// wait for all the events in toBeSent to be handed to libhoney
			<-doneSending
			parsersWG.Done()
		}(lines)
	}
	parsersWG.Wait()
	// tell libhoney to finish up sending events
	libhoney.Close()

	// Nothing bad happened, yay
	logrus.Info("Honeytail is all done, goodbye!")
}

// getParserOptions takes a parser name and the global options struct
// it returns the options group for the specified parser
func getParserAndOptions(options GlobalOptions) (parsers.Parser, interface{}) {
	var parser parsers.Parser
	var opts interface{}
	switch options.Reqs.ParserName {
	case "nginx":
		parser = &nginx.Parser{}
		opts = &options.Nginx
		opts.(*nginx.Options).NumParsers = int(options.NumSenders)
	case "json":
		parser = &htjson.Parser{}
		opts = &options.JSON
		opts.(*htjson.Options).NumParsers = int(options.NumSenders)
	case "keyval":
		parser = &keyval.Parser{}
		opts = &options.KeyVal
		opts.(*keyval.Options).NumParsers = int(options.NumSenders)
	case "mongo", "mongodb":
		parser = &mongodb.Parser{}
		opts = &options.Mongo
		opts.(*mongodb.Options).NumParsers = int(options.NumSenders)
	case "mysql":
		parser = &mysql.Parser{
			SampleRate: int(options.SampleRate),
		}
		opts = &options.MySQL
		opts.(*mysql.Options).NumParsers = int(options.NumSenders)
	case "arangodb":
		parser = &arangodb.Parser{}
		opts = &options.ArangoDB
	}
	parser, _ = parser.(parsers.Parser)
	return parser, opts
}

// modifyEventContents takes a channel from which it will read events. It
// returns a channel on which it will send the munged events.
// It is responsible for hashing or dropping or adding fields to the events
func modifyEventContents(toBeSent chan event.Event, options GlobalOptions) chan event.Event {
	// short circuit this if no field scrubbing is enabled
	if len(options.DropFields) == 0 && len(options.ScrubFields) == 0 &&
		len(options.AddFields) == 0 && len(options.RequestShape) == 0 {
		return toBeSent
	}
	// parse the addField bit once instead of for every event
	parsedAddFields := map[string]string{}
	for _, addField := range options.AddFields {
		splitField := strings.SplitN(addField, "=", 2)
		if len(splitField) != 2 {
			logrus.WithFields(logrus.Fields{
				"add_field": addField,
			}).Fatal("unable to separate provided field into a key=val pair")
		}
		parsedAddFields[splitField[0]] = splitField[1]
	}
	// do all the advance work for request shaping
	shaper := &requestShaper{}
	if len(options.RequestShape) != 0 {
		shaper.pr = &urlshaper.Parser{}
		if options.ShapePrefix != "" {
			shaper.prefix = options.ShapePrefix + "_"
		}
		for _, rpat := range options.RequestPattern {
			pat := urlshaper.Pattern{Pat: rpat}
			if err := pat.Compile(); err != nil {
				logrus.WithField("request_pattern", rpat).WithError(err).Fatal(
					"Failed to compile provided pattern.")
			}
			shaper.pr.Patterns = append(shaper.pr.Patterns, &pat)
		}
	}
	// ok, we need to munge events. Sing up enough goroutines to handle this
	newSent := make(chan event.Event, options.NumSenders)
	go func() {
		wg := sync.WaitGroup{}
		for i := uint(0); i < options.NumSenders; i++ {
			wg.Add(1)
			go func() {
				for ev := range toBeSent {
					for _, field := range options.DropFields {
						delete(ev.Data, field)
					}
					for _, field := range options.ScrubFields {
						if val, ok := ev.Data[field]; ok {
							// generate a sha256 hash and use the base16 for the content
							newVal := sha256.Sum256([]byte(fmt.Sprintf("%v", val)))
							ev.Data[field] = fmt.Sprintf("%x", newVal)
						}
					}
					for k, v := range parsedAddFields {
						ev.Data[k] = v
					}
					for _, field := range options.RequestShape {
						shaper.requestShape(field, &ev, options)
					}
					newSent <- ev

				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(newSent)
	}()
	return newSent
}

// requestShaper holds the bits about request shaping that want to be
// precompiled instead of compute on every event
type requestShaper struct {
	prefix string
	pr     *urlshaper.Parser
}

// requestShape expects the field passed in to have the form
// VERB /path/of/request HTTP/1.x
// If it does, it will break it apart into components, normalize the URL,
// and add a handful of additional fields based on what it finds.
func (r *requestShaper) requestShape(field string, ev *event.Event,
	options GlobalOptions) {
	if val, ok := ev.Data[field]; ok {
		// start by splitting out method, uri, and version
		parts := strings.Split(val.(string), " ")
		var path string
		if len(parts) == 3 {
			// treat it as METHOD /path HTTP/1.X
			ev.Data[r.prefix+field+"_method"] = parts[0]
			ev.Data[r.prefix+field+"_protocol_version"] = parts[2]
			path = parts[1]
		} else {
			// treat it as just the /path
			path = parts[0]
		}
		// next up, get all the goodies out of the path
		res, err := r.pr.Parse(path)
		if err != nil {
			// couldn't parse it, just pass along the event
			return
		}
		ev.Data[r.prefix+field+"_uri"] = res.URI
		ev.Data[r.prefix+field+"_path"] = res.Path
		if res.Query != "" {
			ev.Data[r.prefix+field+"_query"] = res.Query
		}
		for k, v := range res.QueryFields {
			// only include the keys we want
			if options.RequestParseQuery == "all" ||
				whitelistKey(options.RequestQueryKeys, k) {
				if len(v) > 1 {
					sort.Strings(v)
				}
				ev.Data[r.prefix+field+"_query_"+k] = strings.Join(v, ", ")
			}
		}
		for k, v := range res.PathFields {
			ev.Data[r.prefix+field+"_path_"+k] = v[0]
		}
		ev.Data[r.prefix+field+"_shape"] = res.Shape
		ev.Data[r.prefix+field+"_pathshape"] = res.PathShape
		if res.QueryShape != "" {
			ev.Data[r.prefix+field+"_queryshape"] = res.QueryShape
		}
	}
}

// return true if the key is in the whitelist
func whitelistKey(whiteKeys []string, key string) bool {
	for _, whiteKey := range whiteKeys {
		if key == whiteKey {
			return true
		}
	}
	return false
}

// sendToLibhoney reads from the toBeSent channel and shoves the events into
// libhoney events, sending them on their way.
func sendToLibhoney(toBeSent chan event.Event, toBeResent chan event.Event,
	delaySending chan int, doneSending chan bool) {
	for {
		// check and see if we need to back off the API because of rate limiting
		select {
		case delay := <-delaySending:
			time.Sleep(time.Duration(delay) * time.Millisecond)
		default:
		}
		// if we have events to retransmit, send those first
		select {
		case ev := <-toBeResent:
			// retransmitted events have already been sampled; always use
			// SendPresampled() for these
			sendEvent(ev)
			continue
		default:
		}
		// otherwise pick something up off the regular queue and send it
		select {
		case ev, ok := <-toBeSent:
			if !ok {
				// channel is closed
				// NOTE: any unrtransmitted retransmittable events will be dropped
				doneSending <- true
				return
			}
			sendEvent(ev)
			continue
		default:
		}
		// no events at all? chill for a sec until we get the next one
		time.Sleep(100 * time.Millisecond)
	}
}

// sendEvent does the actual handoff to libhoney
func sendEvent(ev event.Event) {
	libhEv := libhoney.NewEvent()
	libhEv.Metadata = ev
	libhEv.Timestamp = ev.Timestamp
	if err := libhEv.Add(ev.Data); err != nil {
		logrus.WithFields(logrus.Fields{
			"event": ev,
			"error": err,
		}).Error("Unexpected error adding data to libhoney event")
	}
	if err := libhEv.SendPresampled(); err != nil {
		logrus.WithFields(logrus.Fields{
			"event": ev,
			"error": err,
		}).Error("Unexpected error event to libhoney send")
	}
}

// handleResponses reads from the response queue, logging a summary and debug
// re-enqueues any events that failed to send in a retryable way
func handleResponses(responses chan libhoney.Response,
	toBeResent chan event.Event, delaySending chan int,
	options GlobalOptions) {
	stats := newResponseStats()
	go logStats(stats, options.StatusInterval)

	for rsp := range responses {
		stats.update(rsp)
		logfields := logrus.Fields{
			"status_code": rsp.StatusCode,
			"body":        strings.TrimSpace(string(rsp.Body)),
			"duration":    rsp.Duration,
			"error":       rsp.Err,
			"timestamp":   rsp.Metadata.(event.Event).Timestamp,
		}
		// if this is an error we should retry sending, re-enqueue the event
		if options.BackOff && (rsp.StatusCode == 429 || rsp.StatusCode == 500) {
			logfields["retry_send"] = true
			delaySending <- 1000 / int(options.NumSenders) // back off for a little bit
			toBeResent <- rsp.Metadata.(event.Event)       // then retry sending the event
		} else {
			logfields["retry_send"] = false
		}
		logrus.WithFields(logfields).Debug("event send record received")
	}
}

// logStats dumps and resets the stats once every minute
func logStats(stats *responseStats, interval uint) {
	logrus.Debugf("Initializing stats reporting. Will print stats once/%d seconds", interval)
	if interval == 0 {
		// interval of 0 means don't print summary status
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		stats.logAndReset()
	}
}
