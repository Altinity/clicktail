package main

import (
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/libhoney-go"
)

// responseStats is a container for collecting statistics about events sent
// via libhoney. It counts interesting aspects of the events it gets and
// presents them for printing whenever it's called.
//
// the intent is to periodically print and flush the counters, eg once/minute

type responseStats struct {
	lock *sync.Mutex

	count       int
	statusCodes map[int]int
	bodies      map[string]int
	errors      map[string]int
	maxDuration time.Duration
	sumDuration time.Duration
	minDuration time.Duration
	event       *event.Event

	totalCount       int
	totalStatusCodes map[int]int
}

// newResponseStats initializes the struct's complex data types
func newResponseStats() *responseStats {
	r := &responseStats{}
	r.totalStatusCodes = make(map[int]int)
	r.lock = &sync.Mutex{}
	r.reset()
	return r
}

// update adds a response into the stats container
func (r *responseStats) update(rsp libhoney.Response) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.count += 1
	r.statusCodes[rsp.StatusCode] += 1
	r.bodies[strings.TrimSpace(string(rsp.Body))] += 1
	if rsp.Err != nil {
		r.errors[rsp.Err.Error()] += 1
	}
	if r.minDuration == 0 {
		r.minDuration = rsp.Duration
	}
	if rsp.Duration < r.minDuration {
		r.minDuration = rsp.Duration
	} else if rsp.Duration > r.maxDuration {
		r.maxDuration = rsp.Duration
	}
	r.sumDuration += rsp.Duration
	ev := rsp.Metadata.(event.Event)
	r.event = &ev
}

// log the current stats and reset them all to zero.
// thread safe.
func (r *responseStats) logAndReset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.log()
	r.reset()
}

// log the current statistics to logrus.
// NOT thread safe.
func (r *responseStats) log() {
	var avg time.Duration
	if r.count != 0 {
		avg = r.sumDuration / time.Duration(r.count)
	} else {
		avg = 0
	}
	logrus.WithFields(logrus.Fields{
		"count":            r.count,
		"lifetime_count":   r.totalCount + r.count,
		"slowest":          r.maxDuration,
		"fastest":          r.minDuration,
		"avg_duration":     avg,
		"count_per_status": r.statusCodes,
		"response_bodies":  r.bodies,
		"errors":           r.errors,
	}).Info("Summary of sent events")
	if r.event != nil {
		fields := make(map[string]interface{})
		fields["event"] = r.event.Data
		fields["event_timestamp"] = r.event.Timestamp
		logrus.WithFields(fields).Info("Last parsed event")
	}
}

// log the total count on its own
func (r *responseStats) logFinal() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.totalCount += r.count
	for code, count := range r.statusCodes {
		r.totalStatusCodes[code] += count
	}
	logrus.WithFields(logrus.Fields{
		"total attempted sends":               r.totalCount,
		"number sent by response status code": r.totalStatusCodes,
	}).Info("Total number of events sent")
}

// reset the counters to zero.
// NOT thread safe
func (r *responseStats) reset() {
	r.totalCount += r.count
	for code, count := range r.statusCodes {
		r.totalStatusCodes[code] += count
	}
	r.count = 0
	r.statusCodes = make(map[int]int)
	r.bodies = make(map[string]int)
	r.errors = make(map[string]int)
	r.maxDuration = 0
	r.sumDuration = 0
	r.minDuration = 0
}
