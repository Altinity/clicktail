package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/honeytail/parsers/htjson"
	"github.com/honeycombio/honeytail/parsers/mongodb"
	"github.com/honeycombio/honeytail/parsers/mysql"
	"github.com/honeycombio/honeytail/parsers/nginx"
	"github.com/honeycombio/honeytail/tail"
	"github.com/honeycombio/libhoney-go"
)

// actually go and be leashy
func run(options GlobalOptions) {
	logrus.Info("Starting leash")

	// spin up our transmission to send events to Honeycomb
	libhConfig := libhoney.Config{
		WriteKey:             options.Reqs.WriteKey,
		Dataset:              options.Reqs.Dataset,
		SampleRate:           options.SampleRate,
		APIHost:              options.APIHost,
		MaxConcurrentBatches: options.NumSenders,
		// block on send should be true so if we can't send fast enough, we slow
		// down reading the log rather than drop lines.
		BlockOnSend: true,
	}
	if err := libhoney.Init(libhConfig); err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(
			"Error occured while spinning up Transimission")
	}

	// get our lines channel from which to read log lines
	lines, err := tail.GetEntries(tail.Config{
		Paths:   options.Reqs.LogFiles,
		Type:    tail.RotateStyleSyslog,
		Options: options.Tail})
	if err != nil {
		logrus.WithFields(logrus.Fields{"err": err}).Fatal(
			"Error occurred while trying to tail logfile")
	}

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
	toBeSent := make(chan event.Event)
	doneSending := make(chan bool)

	// apply any filters to the events before they get sent
	modifiedToBeSent := modifyEventContents(toBeSent, options)

	// start up the sender
	go sendToLibhoney(modifiedToBeSent, doneSending)

	// start a goroutine that reads from responses and logs.
	responses := libhoney.Responses()
	go handleResponses(responses, options)

	// ProcessLines won't return until lines is closed
	parser.ProcessLines(lines, toBeSent)

	// trigger the sending goroutine to finish up
	close(toBeSent)
	// wait for all the events in toBeSent to be handed to libhoney
	<-doneSending

	// tell libhoney to finish up sending events
	libhoney.Close()

	// Nothing bad happened, yay
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
	case "json":
		parser = &htjson.Parser{}
		opts = &options.JSON
	case "mongo", "mongodb":
		parser = &mongodb.Parser{}
		opts = &options.Mongo
	case "mysql":
		parser = &mysql.Parser{}
		opts = &options.MySQL
	}
	parser, _ = parser.(parsers.Parser)
	return parser, opts
}

// modifyEventContents takes a channel from which it will read events. It
// returns a channel on which it will send the munged events.
// It is responsible for hashing or dropping or adding fields to the events
func modifyEventContents(toBeSent chan event.Event, options GlobalOptions) chan event.Event {
	for _, field := range options.DropFields {
		toBeSent = dropEventField(field, toBeSent)
	}
	for _, field := range options.ScrubFields {
		toBeSent = scrubEventField(field, toBeSent)
	}
	for _, field := range options.AddFields {
		toBeSent = addEventField(field, toBeSent)
	}
	return toBeSent
}

// dropEventField drops any fields that are to be dropped, drop them before
// passing the event on down the line to the next consumer
func dropEventField(field string, toBeSent chan event.Event) chan event.Event {
	newSent := make(chan event.Event)
	go func() {
		for ev := range toBeSent {
			delete(ev.Data, field)
			newSent <- ev
		}
		close(newSent)
	}()
	return newSent
}

// scrubEventField replaces the value for  any fields that are to be scrubbed
// with a sha256 hash of the value, then passes the event on down the line to
// the next consumer
func scrubEventField(field string, toBeSent chan event.Event) chan event.Event {
	newSent := make(chan event.Event)
	go func() {
		for ev := range toBeSent {
			if val, ok := ev.Data[field]; ok {
				// generate a sha256 hash
				newVal := sha256.Sum256([]byte(fmt.Sprintf("%v", val)))
				// and use the base16 string version of it
				ev.Data[field] = fmt.Sprintf("%x", newVal)
			}
			newSent <- ev
		}
		close(newSent)
	}()
	return newSent
}

// addEventField adds any fields that are to be added to the event before
// passing the event on down the line to the next consumer
func addEventField(field string, toBeSent chan event.Event) chan event.Event {
	newSent := make(chan event.Event)
	// separate the k=v field we got from the command line
	splitField := strings.SplitN(field, "=", 2)
	if len(splitField) != 2 {
		logrus.WithFields(logrus.Fields{
			"add_field": field,
		}).Fatal("unable to separate provided field into a key=val pair")
	}
	key := splitField[0]
	val := splitField[1]
	go func() {
		for ev := range toBeSent {
			ev.Data[key] = val
			newSent <- ev
		}
		close(newSent)
	}()
	return newSent
}

// sendToLibhoney reads from the toBeSent channel and shoves the events into
// libhoney events, sending them on their way.
func sendToLibhoney(toBeSent chan event.Event, doneSending chan bool) {
	for ev := range toBeSent {
		libhEv := libhoney.NewEvent()
		libhEv.Metadata = rand.Intn(1000000)
		libhEv.Timestamp = ev.Timestamp
		if err := libhEv.Add(ev.Data); err != nil {
			logrus.WithFields(logrus.Fields{
				"event": ev,
				"error": err,
			}).Error("Unexpected error adding data to libhoney event")
		}
		if err := libhEv.Send(); err != nil {
			logrus.WithFields(logrus.Fields{
				"event": ev,
				"error": err,
			}).Error("Unexpected error event to libhoney send")
		}
	}
	doneSending <- true
}

// handleResponses reads from the response queue, logging a summary and debug
func handleResponses(responses chan libhoney.Response, options GlobalOptions) {
	stats := newResponseStats()
	go logStats(stats, options.StatusInterval)

	for rsp := range responses {
		stats.update(rsp)
		logrus.WithFields(logrus.Fields{
			"event_id":    rsp.Metadata,
			"status_code": rsp.StatusCode,
			"body":        strings.TrimSpace(string(rsp.Body)),
			"duration":    rsp.Duration,
			"error":       rsp.Err,
		}).Debug("event sent")
	}
}

// logStats dumps and resets the stats once every minute
func logStats(stats *responseStats, interval uint) {
	logrus.Debugf("Initializing stats reporting. Will print stats once/%d seconds", interval)
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for range ticker.C {
		stats.logAndReset()
	}
}
