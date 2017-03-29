// Package keyval parses logs whose format is many key=val pairs
package keyval

import (
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kr/logfmt"

	"github.com/honeycombio/honeytail/event"
	"github.com/honeycombio/honeytail/parsers"
)

var possibleTimeFieldNames = []string{
	"time", "Time",
	"timestamp", "Timestamp", "TimeStamp",
	"date", "Date",
	"datetime", "Datetime", "DateTime",
}

type Options struct {
	TimeFieldName string `long:"timefield" description:"Name of the field that contains a timestamp"`
	Format        string `long:"format" description:"Format of the timestamp found in timefield (supports strftime and Golang time formats)"`
}

type Parser struct {
	conf       Options
	lineParser LineParser
	nower      Nower

	warnedAboutTime bool
}

type Nower interface {
	Now() time.Time
}

type RealNower struct{}

func (r *RealNower) Now() time.Time {
	return time.Now().UTC()
}

func (p *Parser) Init(options interface{}) error {
	p.conf = *options.(*Options)

	p.nower = &RealNower{}
	p.lineParser = &KeyValLineParser{}
	return nil
}

// TODO dedupe the LineParser interface with the json parser or remove entirely
type LineParser interface {
	ParseLine(line string) (map[string]interface{}, error)
}

type KeyValLineParser struct {
}

func (j *KeyValLineParser) ParseLine(line string) (map[string]interface{}, error) {
	parsed := make(map[string]interface{})
	f := func(key, val []byte) error {
		keyStr := string(key)
		valStr := string(val)
		if b, err := strconv.ParseBool(valStr); err == nil {
			parsed[keyStr] = b
			return nil
		}
		if i, err := strconv.Atoi(valStr); err == nil {
			parsed[keyStr] = i
			return nil
		}
		if f, err := strconv.ParseFloat(valStr, 64); err == nil {
			parsed[keyStr] = f
			return nil
		}
		parsed[keyStr] = valStr
		return nil
	}
	err := logfmt.Unmarshal([]byte(line), logfmt.HandlerFunc(f))
	return parsed, err
}

func (p *Parser) ProcessLines(lines <-chan string, send chan<- event.Event, prefixRegex *parsers.ExtRegexp) {
	for line := range lines {
		logrus.WithFields(logrus.Fields{
			"line": line,
		}).Debug("Attempting to process keyval log line")

		// take care of any headers on the line
		var prefixFields map[string]string
		if prefixRegex != nil {
			var prefix string
			prefix, prefixFields = prefixRegex.FindStringSubmatchMap(line)
			line = strings.TrimPrefix(line, prefix)
		}

		parsedLine, err := p.lineParser.ParseLine(line)

		if err != nil {
			// skip lines that won't parse
			logrus.WithFields(logrus.Fields{
				"line":  line,
				"error": err,
			}).Debug("skipping line; failed to parse.")
			continue
		}
		// merge the prefix fields and the parsed line contents
		for k, v := range prefixFields {
			parsedLine[k] = v
		}

		// look for the timestamp in any of the prefix fields or regular content
		timestamp := p.getTimestamp(parsedLine)

		// send an event to Transmission
		e := event.Event{
			Timestamp: timestamp,
			Data:      parsedLine,
		}
		send <- e
	}
	logrus.Debug("lines channel is closed, ending keyval processor")
}

// getTimestamp looks through the event map for something that looks
// like a timestamp. It will guess at the key name or use
// the one from Config if it is not ""
// if unable to parse it will return the current time
// it is highly recommended that you populate the Config with time format
// sample from logrus: "time":"2014-03-10 19:57:38.562264131 -0400 EDT"
// TODO remove fancy time parsing from the keyval parser, since timestamps
// are likely to be more well structured and come from the prefix
func (p *Parser) getTimestamp(m map[string]interface{}) time.Time {
	var ts time.Time
	if p.conf.TimeFieldName != "" {
		// remove the timestamp from the body when we stuff it in the header
		defer delete(m, p.conf.TimeFieldName)
		if t, found := m[p.conf.TimeFieldName]; found {
			timeStr, ok := t.(string)
			if !ok {
				timeInt, ok := t.(int)
				if !ok {
					p.warnAboutTime(p.conf.TimeFieldName, t, "found time field but unknown type")
					timeStr = p.nower.Now().String()
				} else {
					timeStr = strconv.Itoa(timeInt)
				}
			}
			ts = p.tryTimeFormats(timeStr)
			if ts.IsZero() {
				p.warnAboutTime(p.conf.TimeFieldName, t, "found time field but failed to parse")
				ts = p.nower.Now()
			}
		} else {
			p.warnAboutTime(p.conf.TimeFieldName, nil, "couldn't find specified time field")
			ts = p.nower.Now()
		}
		// we were told to look for a specific field;
		// let's return what we found instead of continuing to look.
		return ts
	}
	// go through all the possible fields that might have a timestamp
	// for the first one we find, if it's a string field, try and parse it
	// if we succeed, stop looking. Otherwise keep trying
	for _, timeField := range possibleTimeFieldNames {
		if t, found := m[timeField]; found {
			timeStr, found := t.(string)
			if found {
				defer delete(m, timeField)
				ts = p.tryTimeFormats(timeStr)
				if !ts.IsZero() {
					break
				}
				p.warnAboutTime(timeField, t, "inferred timestamp field but failed parse as valid time")
			}
		}
	}
	if ts.IsZero() {
		ts = p.nower.Now()
	}
	return ts
}

func (p *Parser) tryTimeFormats(t string) time.Time {
	// golang can't parse times with decimal fractional seconds marked by a comma
	// hack it by just replacing all commas with periods and hope it works out.
	// https://github.com/golang/go/issues/6189
	t = strings.Replace(t, ",", ".", -1)
	if p.conf.Format == UnixTimestampFmt {
		if unix, err := strconv.ParseInt(t, 0, 64); err == nil {
			return time.Unix(unix, 0)
		}
	}
	if p.conf.Format != "" {
		format := strings.Replace(p.conf.Format, ",", ".", -1)
		if strings.Contains(format, StrftimeChar) {
			if ts, err := time.Parse(convertTimeFormat(format), t); err == nil {
				return ts
			}
		}

		// Still try Go style, just in case
		if ts, err := time.Parse(format, t); err == nil {
			return ts
		}
	}

	var ts time.Time
	if tOther, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", t); err == nil {
		ts = tOther
	} else if tOther, err := time.Parse(time.RFC3339Nano, t); err == nil {
		ts = tOther
	} else if tOther, err := time.Parse(time.RubyDate, t); err == nil {
		ts = tOther
	} else if tOther, err := time.Parse(time.UnixDate, t); err == nil {
		ts = tOther
	}
	return ts
}

func (p *Parser) warnAboutTime(fieldName string, foundTimeVal interface{}, msg string) {
	if p.warnedAboutTime {
		return
	}
	logrus.WithField("time_field", fieldName).WithField("time_value", foundTimeVal).Warn(msg + "\n  Please refer to https://honeycomb.io/docs/json#timestamp-parsing")
	p.warnedAboutTime = true
}
