// Package regex consumes logs given user-defined regex format for lines

// RE2 regex syntax reference: https://github.com/google/re2/wiki/Syntax
// Example format for a named capture group: `(?P<name>re)`

package regex

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/httime"
	"github.com/honeycombio/honeytail/parsers"
)

type Options struct {
	// Note: `LineRegex` and `line_regex` are named as singular so that
	// it's less confusing to users to input them.
	// Might be worth making this consistent across the entire repo
	LineRegex       []string `long:"line_regex" description:"Regular expression with named capture groups representing the fields you want parsed (RE2 syntax). You can enter multiple regexes to match (--regex.line_regex=\"(?P<foo>re)\" --regex.line_regex=\"(?P<bar>...)\"). Parses using the first regex to match a line, so list them in most-to-least-specific order."`
	TimeFieldName   string   `long:"timefield" description:"Name of the field that contains a timestamp"`
	TimeFieldFormat string   `long:"time_format" description:"Timestamp format to use (strftime and Golang time.Parse supported)"`
	NumParsers      int      `hidden:"true" description:"number of regex parsers to spin up"`
}

type Parser struct {
	conf       Options
	lineParser parsers.LineParser
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)
	if len(p.conf.LineRegex) == 0 {
		return errors.New("Must provide at least one regex for parsing log lines; use `--regex.line_regex` flag.")
	}
	lineParser, err := NewRegexLineParser(p.conf.LineRegex)
	if err != nil {
		return err
	}
	p.lineParser = lineParser
	return nil
}

// Compile multiple log line regexes
func ParseLineRegexes(regexStrs []string) ([]*regexp.Regexp, error) {
	regexes := make([]*regexp.Regexp, 0)
	for _, regexStr := range regexStrs {
		regex, err := ParseLineRegex(regexStr)
		if err != nil {
			return regexes, err
		}
		regexes = append(regexes, regex)
	}
	return regexes, nil
}

// Compile a regex & validate expectations for log line parsing
func ParseLineRegex(regexStr string) (*regexp.Regexp, error) {
	// Regex can't be blank
	if regexStr == "" {
		logrus.Debug("LineRegex is blank; required field")
		return nil, errors.New("Must provide a regex for parsing log lines; use `--regex.line_regex` flag.")
	}

	// Compile regex
	lineRegex, err := regexp.Compile(regexStr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"lineRegex": regexStr,
		}).Error("Could not compile line regex")
		return nil, err
	}

	// Require at least one named group
	var numNamedGroups int
	for _, groupName := range lineRegex.SubexpNames() {
		if groupName != "" {
			numNamedGroups++
		}
	}
	if numNamedGroups == 0 {
		logrus.WithFields(logrus.Fields{
			"LineRegex": regexStr,
		}).Error("No named capture groups")
		return nil, errors.New(fmt.Sprintf("No named capture groups found in regex: '%s'. Must provide at least one named group with line regex. Example: `(?P<name>re)`", regexStr))
	}

	return lineRegex, nil
}

/* LineParser for regexes */

type RegexLineParser struct {
	lineRegexes []*regexp.Regexp
}

// RegexLineParser factory
func NewRegexLineParser(regexStrs []string) (*RegexLineParser, error) {
	lineRegexes, err := ParseLineRegexes(regexStrs)
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"lineRegexes": lineRegexes,
	}).Debug("Compiled line regexes")
	return &RegexLineParser{lineRegexes}, nil
}

func (p *RegexLineParser) ParseLine(line string) (map[string]interface{}, error) {
	for _, lineRegex := range p.lineRegexes {
		parsed := make(map[string]interface{})
		match := lineRegex.FindAllStringSubmatch(line, -1)
		if match == nil || len(match) == 0 {
			logrus.WithFields(logrus.Fields{
				"line":      line,
				"lineRegex": lineRegex,
			}).Debug("No matches for regex log line")
			continue // No matches found, skip to next regex
		}

		// Map capture groups
		var firstMatch []string = match[0] // We only care about the first full lineRegex match
		for i, name := range lineRegex.SubexpNames() {
			if i != 0 && i < len(firstMatch) {
				parsed[name] = firstMatch[i]
			}
		}
		logrus.WithFields(logrus.Fields{
			"parsed":    parsed,
			"line":      line,
			"lineRegex": lineRegex,
		}).Debug("Regex parsing log line")

		return parsed, nil
	}
	return make(map[string]interface{}), nil
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	// parse lines one by one
	wg := sync.WaitGroup{}
	numParsers := 1
	if p.conf.NumParsers > 0 {
		numParsers = p.conf.NumParsers
	}
	for i := 0; i < numParsers; i++ {
		wg.Add(1)
		go func() {
			for line := range lines {
				logrus.WithFields(logrus.Fields{
					"line": line,
				}).Debug("Attempting to process regex log line")

				// take care of any headers on the line
				var prefixFields map[string]string
				if prefixRegex != nil {
					var prefix string
					prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
					line = strings.TrimPrefix(line, prefix)
				}

				parsedLine, err := p.lineParser.ParseLine(line)
				if err != nil {
					continue
				}

				// merge the prefix fields and the parsed line contents
				for k, v := range prefixFields {
					parsedLine[k] = v
				}

				if len(parsedLine) == 0 {
					logrus.WithFields(logrus.Fields{
						"line": line,
					}).Debug("Skipping line; no capture groups found")
					continue
				}

				// look for the timestamp in any of the prefix fields or regular content
				timestamp := httime.GetTimestamp(parsedLine, p.conf.TimeFieldName, p.conf.TimeFieldFormat)

				// send an event to Transmission
				e := event.Event{
					Timestamp: timestamp,
					Data:      parsedLine,
				}
				send <- e
			}
			wg.Done()
		}()
	}
	wg.Wait()
	logrus.Debug("lines channel is closed, ending regex processor")
}
