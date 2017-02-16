package logparser

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	endRune rune = 1114112
)

type partialLogLineError struct {
	InnerError error
}

func (p partialLogLineError) Error() string {
	return fmt.Sprintf("Partial log line: %v", p.InnerError)
}

func IsPartialLogLine(err error) bool {
	_, ok := err.(partialLogLineError)
	return ok
}

func ParseLogLine(input string) (map[string]interface{}, error) {
	p := LogLineParser{Buffer: input}
	p.Init()
	if err := p.Parse(); err != nil {
		return nil, err
	}
	return p.Fields, nil
}

func ParseQuery(query string) (map[string]interface{}, error) {
	p := LogLineParser{Buffer: query}
	p.Init()
	rv, err := p.parseFieldValue("query")
	if err != nil {
		return nil, err
	}
	if m, ok := rv.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, errors.New("query string does not parse to a doc")
}

type LogLineParser struct {
	Buffer string
	Fields map[string]interface{}

	runes    []rune
	position int
}

func (p *LogLineParser) Init() {
	p.runes = append([]rune(p.Buffer), endRune)
	p.Fields = make(map[string]interface{})
}

func (p *LogLineParser) Parse() error {
	var err error
	if err = p.parseTimestamp(); err != nil {
		return err
	}
	if p.eatWS().lookahead(0) == '[' {
		// we assume version < 3.0
		if err = p.parseContext(); err != nil {
			return err
		}
		err = p.parseMessage()
	} else {
		// we assume version > 3.0
		if err = p.parseSeverity(); err != nil {
			return err
		}
		if err = p.parseComponent(); err != nil {
			return err
		}
		if err = p.parseContext(); err != nil {
			return err
		}
		err = p.parseMessage()
	}

	if err != nil {
		return partialLogLineError{InnerError: err}
	}

	return nil
}

func (p *LogLineParser) parseTimestamp() error {
	var readTimestamp string
	var err error

	c := p.eatWS().lookahead(0)
	if unicode.IsDigit(c) {
		// we assume it's either iso8601-utc or iso8601-local
		if readTimestamp, err = p.readUntil(unicode.Space); err != nil {
			return err
		}
	} else {
		// we assume it's ctime or ctime-no-ms
		var dayOfWeek, month, day, time string

		if dayOfWeek, err = validDayOfWeek(p.readUntil(unicode.Space)); err != nil {
			return err
		}

		if month, err = validMonth(p.eatWS().readUntil(unicode.Space)); err != nil {
			return err
		}

		if day, err = p.eatWS().readUntil(unicode.Space); err != nil {
			return err
		}

		if time, err = p.eatWS().readUntil(unicode.Space); err != nil {
			return err
		}
		readTimestamp = dayOfWeek + " " + month + " " + day + " " + time
	}

	p.Fields["timestamp"] = readTimestamp
	return nil
}

func (p *LogLineParser) parseSeverity() error {
	var err error
	if p.Fields["severity"], err = severityToString(p.eatWS().advance()); err != nil {
		return err
	}
	if err = p.expect(unicode.Space); err != nil {
		return err
	}
	return nil
}

func (p *LogLineParser) parseComponent() error {
	var component string
	var err error

	if p.eatWS().lookahead(0) == '-' {
		component = "-"
		p.advance() // skip the '-'
	} else {
		if component, err = p.readWhile([]interface{}{unicode.Upper}); err != nil {
			return err
		}
	}
	if !p.validComponentName(component) {
		return errors.New("unrecognized component name")
	}

	p.Fields["component"] = component
	return nil
}

func (p *LogLineParser) parseContext() error {
	var err error
	if err = p.eatWS().expect('['); err != nil {
		return err
	}

	var context string
	if context, err = p.readUntilRune(']'); err != nil {
		return err
	}
	p.advance() // skip the ']'

	p.Fields["context"] = context
	return nil
}

func (p *LogLineParser) parseSharding() error {
	message, err := p.readUntilRune(':')
	if err != nil {
		return err
	}

	p.advance() // skip the ':'
	p.eatWS()

	if !strings.HasPrefix(message, "about to log metadata event into") {
		return errors.New("unrecognized sharding log line")
	}

	p.Fields["sharding_message"] = message
	lastSpace := strings.LastIndex(message, " ")
	p.Fields["sharding_collection"] = message[lastSpace+1:]

	var changelog interface{}
	if changelog, err = p.parseJSONMap(); err != nil {
		return err
	}
	p.Fields["sharding_changelog"] = changelog
	return nil
}

func (p *LogLineParser) parseMessage() error {
	p.eatWS()

	savedPosition := p.position

	if p.Fields["component"] == "SHARDING" {
		savedPosition := p.position
		err := p.parseSharding()
		if err == nil {
			return nil
		}
		p.position = savedPosition
	}

	// check if this message is an operation
	operation, err := p.readUntil(unicode.Space)
	if err == nil && p.validOperationName(operation) {
		// yay, an operation.
		p.Fields["operation"] = operation

		var namespace string
		if namespace, err = p.eatWS().readUntil(unicode.Space); err != nil {
			return err
		}
		p.Fields["namespace"] = namespace

		if err = p.parseOperationBody(); err != nil {
			return err
		}
	} else {
		p.position = savedPosition

		if p.Fields["message"], err = p.readUntilEOL(); err != nil {
			return err
		}
	}

	return nil
}

func (p *LogLineParser) parseOperationBody() error {
	for p.runes[p.position] != endRune {
		var err error
		var done bool

		if done, err = p.parseFieldAndValue(); err != nil {
			return err
		}
		if done {
			// check for a duration
			dur, err := p.readDuration()
			if err != nil {
				return err
			}
			p.Fields["duration_ms"] = dur
			break
		}
	}
	return nil
}

func (p *LogLineParser) parseFieldAndValue() (bool, error) {
	var fieldName string
	var fieldValue interface{}
	var err error

	p.eatWS()

	savedPosition := p.position
	if fieldName, err = p.readWhileNot([]interface{}{':', unicode.Space}); err != nil {
		p.position = savedPosition
		return true, nil // swallow the error to give our caller a change to backtrack
	}
	p.advance() // skip the ':'/WS
	p.eatWS()   // end eat any remaining WS

	// some known fields have a more complicated structure
	if fieldName == "planSummary" {
		if fieldValue, err = p.parsePlanSummary(); err != nil {
			return false, err
		}
	} else if fieldName == "command" {
		// >=2.6 has:  command: <command_name> <command_doc>?
		// <2.6 has:   command: <command_doc>
		firstCharInVal := p.lookahead(0)
		if firstCharInVal != '{' {
			name, err := p.readJSONIdentifier()
			if err != nil {
				return false, err
			}
			p.eatWS()
			p.Fields["command_type"] = name
		}

		if fieldValue, err = p.parseJSONMap(); err != nil {
			return false, err
		}
	} else if fieldName == "locks(micros)" {
		// < 2.8
		if fieldValue, err = p.parseLocksMicro(); err != nil {
			return false, err
		}
		p.eatWS()
	} else {
		if fieldValue, err = p.parseFieldValue(fieldName); err != nil {
			return false, err
		}
		if !p.validFieldName(fieldName) {
			return false, nil
		}
	}
	p.Fields[fieldName] = fieldValue
	return false, nil
}

func (p *LogLineParser) validFieldName(fieldName string) bool {
	if len(fieldName) == 0 {
		return false
	}
	for _, c := range fieldName {
		switch {
		case unicode.IsLetter(c):
			continue
		case unicode.IsDigit(c):
			continue
		case c == '_':
			continue
		case c == '$':
			continue
		default:
			return false
		}
	}
	return true
}

func (p *LogLineParser) parseFieldValue(fieldName string) (interface{}, error) {
	var fieldValue interface{}
	var err error

	firstCharInVal := p.lookahead(0)
	switch {
	case firstCharInVal == '{':
		if fieldValue, err = p.parseJSONMap(); err != nil {
			return nil, err
		}
	case unicode.IsDigit(firstCharInVal):
		if fieldValue, err = p.readNumber(); err != nil {
			return nil, err
		}
	case unicode.IsLetter(firstCharInVal):
		if fieldValue, err = p.readJSONIdentifier(); err != nil {
			return nil, err
		}
	case firstCharInVal == '"':
		if fieldValue, err = p.readStringValue(firstCharInVal); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected start character for value of field '%s'", fieldName)
	}
	return fieldValue, nil
}

func (p *LogLineParser) parseLocksMicro() (map[string]int64, error) {
	rv := make(map[string]int64)

	for {
		c := p.eatWS().lookahead(0)
		if c != 'r' && c != 'R' && c != 'w' && c != 'W' {
			return rv, nil
		} else if p.lookahead(1) != ':' {
			return rv, nil
		}

		p.advance()
		p.advance()

		// not strictly correct - the value here should be an integer, not a float
		var duration float64
		var err error
		if duration, err = p.readNumber(); err != nil {
			return nil, err
		}
		rv[string([]rune{c})] = int64(duration)
	}

}

func (p *LogLineParser) parsePlanSummary() (interface{}, error) {
	var rv []interface{}

	for {
		elem, err := p.parsePlanSummaryElement()
		if err != nil {
			return nil, err
		}
		if elem != nil {
			rv = append(rv, elem)
		}

		if p.eatWS().lookahead(0) != ',' {
			break
		} else {
			p.advance() // skip the ','
		}
	}

	return rv, nil
}

func (p *LogLineParser) parsePlanSummaryElement() (interface{}, error) {
	rv := make(map[string]interface{})

	p.eatWS()

	savedPosition := p.position

	var stage string
	var err error

	if stage, err = p.readUpcaseIdentifier(); err != nil {
		p.position = savedPosition
		return nil, nil
	}

	c := p.eatWS().lookahead(0)
	if c == '{' {
		if rv[stage], err = p.parseJSONMap(); err != nil {
			return nil, nil
		}
	} else {
		rv[stage] = true
	}

	return rv, nil
}

func (p *LogLineParser) readNumber() (float64, error) {
	startPosition := p.position
	endPosition := startPosition
	numberChecks := []interface{}{unicode.Digit, '.', '+', '-', 'e', 'E'}
	for check(p.runes[endPosition], numberChecks) {
		endPosition++
	}

	if p.runes[endPosition] == endRune {
		return 0, errors.New("found end of line before expected unicode range")
	}

	p.position = endPosition

	return strconv.ParseFloat(string(p.runes[startPosition:endPosition]), 64)
}

func (p *LogLineParser) readDuration() (float64, error) {
	startPosition := p.position
	endPosition := startPosition

	for unicode.IsDigit(p.runes[endPosition]) {
		endPosition++
	}

	if p.runes[endPosition] != 'm' || p.runes[endPosition+1] != 's' {
		return 0, errors.New("invalid duration specifier")
	}

	rv, err := strconv.ParseFloat(string(p.runes[startPosition:endPosition]), 64)
	p.position = endPosition + 2
	return rv, err
}

func (p *LogLineParser) parseJSONMap() (interface{}, error) {
	// we assume we're on the '{'
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	rv := make(map[string]interface{})

	for {
		var key string
		var value interface{}
		var err error

		// we support keys both of the form: { foo: ... } and { "foo": ... }
		fc := p.eatWS().lookahead(0)
		if fc == '"' || fc == '\'' {
			if key, err = p.readStringValue(fc); err != nil {
				return nil, err
			}
		} else {
			if key, err = p.readJSONIdentifier(); err != nil {
				return nil, err
			}
		}

		if key != "" {
			if err = p.eatWS().expect(':'); err != nil {
				return nil, err
			}
			if value, err = p.eatWS().parseJSONValue(); err != nil {
				return nil, err
			}
			rv[key] = value
		}

		commaOrRbrace := p.eatWS().lookahead(0)
		if commaOrRbrace == '}' {
			p.position++
			break
		} else if commaOrRbrace == ',' {
			p.position++
		} else {
			return nil, errors.New("expected '}' or ',' in json")
		}

	}

	return rv, nil
}

func (p *LogLineParser) parseJSONArray() (interface{}, error) {
	var rv []interface{}

	// we assume we're on the '['
	if err := p.expect('['); err != nil {
		return nil, err
	}

	if p.eatWS().lookahead(0) == ']' {
		p.advance()
		return rv, nil
	}

	for {
		var value interface{}
		var err error

		if value, err = p.eatWS().parseJSONValue(); err != nil {
			return nil, err
		}

		rv = append(rv, value)

		commaOrRbrace := p.eatWS().lookahead(0)
		if commaOrRbrace == ']' {
			p.position++
			break
		} else if commaOrRbrace == ',' {
			p.position++
		} else {
			return nil, errors.New("expected ']' or ',' in json")
		}
	}

	return rv, nil
}

func (p *LogLineParser) parseJSONValue() (interface{}, error) {
	var value interface{}
	var err error

	firstCharInVal := p.lookahead(0)
	switch {
	case firstCharInVal == '{':
		if value, err = p.parseJSONMap(); err != nil {
			return nil, err
		}
	case firstCharInVal == '[':
		if value, err = p.parseJSONArray(); err != nil {
			return nil, err
		}
	case check(firstCharInVal, []interface{}{unicode.Digit, '-', '+', '.'}):
		if value, err = p.readNumber(); err != nil {
			return nil, err
		}
	case firstCharInVal == '"':
		// mongo doesn't follow generally accepted rules on how to handle nested quotes
		// when the inner quote character matches the outer quote character (escaping the inner
		// quote with a \).

		// so we have to do something equally terrible to read these values.  we look ahead until we
		// find a value separator or an end to a json value - , ] }
		// that occurs after an even number of quotes.

		savedPosition := p.position + 1
		endPosition := savedPosition

		quoteCount := 1
		quotePosition := savedPosition - 1

		if endPosition < len(p.runes)-1 {
			lastRune := '"'

			for {
				r := p.runes[endPosition]
				if r == '"' {
					quoteCount++
					quotePosition = endPosition
				} else if (r == ',' || r == '}' || r == ']') && lastRune == '"' {
					if quoteCount%2 == 0 {
						value = string(p.runes[savedPosition:quotePosition])
						p.position = quotePosition + 1
						break
					}
				}

				if !unicode.IsSpace(r) {
					lastRune = r
				}

				endPosition++
				if endPosition == len(p.runes) {
					return nil, errors.New("unexpected end of line reading json value")
				}
			}
		}
	case unicode.IsLetter(firstCharInVal):
		if value, err = p.readJSONIdentifier(); err != nil {
			return nil, err
		}
		if value == "null" {
			value = nil
		} else if value == "true" {
			value = true
		} else if value == "false" {
			value = false
		} else if value == "new" {
			if value, err = p.eatWS().readJSONIdentifier(); err != nil {
				return nil, err
			}
			if value != "Date" {
				return nil, fmt.Errorf("unexpected constructor: %s", value)
			}
			// we expect "new Date(123456789)"
			if err = p.expect('('); err != nil {
				return nil, err
			}
			var dateNum float64
			if dateNum, err = p.readNumber(); err != nil {
				return nil, err
			}
			if err = p.expect(')'); err != nil {
				return nil, err
			}

			if math.Floor(dateNum) != dateNum {
				return nil, errors.New("expected int in `new Date()`")
			}
			unixSec := int64(dateNum) / 1000
			unixNS := int64(dateNum) % 1000 * 1000000
			value = time.Unix(unixSec, unixNS)
		} else if value == "Timestamp" {
			var ts string
			if p.lookahead(0) == '(' {
				if ts, err = p.readUntilRune(')'); err != nil {
					return nil, err
				}
			} else {
				if ts, err = p.eatWS().readWhile([]interface{}{unicode.Digit, '|'}); err != nil {
					return nil, err
				}
			}
			value = "Timestamp(" + ts + ")"
		} else if value == "ObjectId" {
			if err = p.expect('('); err != nil {
				return nil, err
			}
			quote := p.lookahead(0) // keep ahold of the quote so we can match it
			if p.lookahead(0) != '\'' && p.lookahead(0) != '"' {
				return nil, errors.New("expected ' or \" in ObjectId")
			}
			p.position++

			hexRunes := []interface{}{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'a', 'b', 'c', 'd', 'e', 'f'}
			var hex string
			if hex, err = p.readWhile(hexRunes); err != nil {
				return nil, err
			}
			if err = p.expect(quote); err != nil {
				return nil, err
			}
			if err = p.expect(')'); err != nil {
				return nil, err
			}
			value = "ObjectId(" + hex + ")"
		} else {
			return nil, fmt.Errorf("unexpected start of JSON value: %s", value)
		}
	default:
		return nil, fmt.Errorf("unexpected start character for JSON value of field: %c", firstCharInVal)
	}

	return value, nil
}

func (p *LogLineParser) readStringValue(quote rune) (string, error) {
	var s string
	var err error

	p.advance() // skip starting quote
	if s, err = p.readUntilRune(quote); err != nil {
		return "", err
	}
	p.advance() // skip ending quote

	return s, nil
}

func (p *LogLineParser) readJSONIdentifier() (string, error) {
	startPosition := p.position
	endPosition := startPosition

	for check(p.runes[endPosition], []interface{}{unicode.Letter, unicode.Digit, '$', '_', '.', '*'}) {
		endPosition++
	}

	p.position = endPosition
	return string(p.runes[startPosition:endPosition]), nil
}

func (p *LogLineParser) readUpcaseIdentifier() (string, error) {
	return p.readWhile([]interface{}{unicode.Upper, unicode.Digit, '_'})
}

func (p *LogLineParser) readAlphaIdentifier() (string, error) {
	return p.readWhile([]interface{}{unicode.Letter, unicode.Digit, '_'})
}

func (p *LogLineParser) readUntil(untilRangeTable *unicode.RangeTable) (string, error) {
	startPosition := p.position
	endPosition := startPosition
	for p.runes[endPosition] != endRune && !unicode.Is(untilRangeTable, p.runes[endPosition]) {
		endPosition++
	}

	if p.runes[endPosition] == endRune {
		return "", errors.New("found end of line before expected unicode range")
	}

	p.position = endPosition

	return string(p.runes[startPosition:endPosition]), nil
}

func (p *LogLineParser) readUntilRune(untilRune rune) (string, error) {
	startPosition := p.position
	endPosition := startPosition
	for p.runes[endPosition] != untilRune && p.runes[endPosition] != endRune {
		endPosition++
	}

	if p.runes[endPosition] == endRune && untilRune != endRune {
		return "", fmt.Errorf("found end of line before expected rune '%c'", untilRune)
	}

	p.position = endPosition

	return string(p.runes[startPosition:endPosition]), nil
}

func (p *LogLineParser) readUntilEOL() (string, error) {
	return p.readUntilRune(endRune)
}

func (p *LogLineParser) readWhile(checks []interface{}) (string, error) {
	return p._readWhile(checks, false)
}

func (p *LogLineParser) readWhileNot(checks []interface{}) (string, error) {
	return p._readWhile(checks, true)
}

func (p *LogLineParser) _readWhile(checks []interface{}, checkStopVal bool) (string, error) {
	startPosition := p.position
	endPosition := startPosition

	for p.runes[endPosition] != endRune {
		if check(p.runes[endPosition], checks) == checkStopVal {
			break
		}
		endPosition++
	}

	if p.runes[endPosition] == endRune {
		return "", errors.New("unexpected end of line")
	}

	p.position = endPosition

	return string(p.runes[startPosition:endPosition]), nil
}

func (p *LogLineParser) lookahead(amount int) rune {
	return p.runes[p.position+amount]
}

func (p *LogLineParser) matchAhead(startIdx int, s string) bool {
	runes := []rune(s)
	for i, r := range runes {
		if r != p.runes[startIdx+i] {
			return false
		}
	}
	return true
}

func (p *LogLineParser) advance() rune {
	r := p.runes[p.position]
	p.position++
	return r
}

func (p *LogLineParser) expect(c interface{}) error {
	r := p.advance()
	matches := doCheck(r, c)
	if !matches {
		return fmt.Errorf("unexpected '%c'", r)
	}
	return nil
}

func (p *LogLineParser) eatWS() *LogLineParser {
	for unicode.Is(unicode.Space, p.runes[p.position]) {
		p.position++
	}
	return p
}

func severityToString(sev rune) (string, error) {
	switch sev {
	case 'D':
		return "debug", nil
	case 'I':
		return "informational", nil
	case 'W':
		return "warning", nil
	case 'E':
		return "error", nil
	case 'F':
		return "fatal", nil
	default:
		return "", fmt.Errorf("unknown severity '%c'", sev)
	}
}

func check(r rune, checks []interface{}) bool {
	for _, c := range checks {
		if doCheck(r, c) {
			return true
		}
	}
	return false
}

func doCheck(r rune, c interface{}) bool {
	if rt, ok := c.(*unicode.RangeTable); ok {
		if unicode.Is(rt, r) {
			return true
		}
	} else if runeCheck, ok := c.(rune); ok {
		if r == runeCheck {
			return true
		}
	} else {
		panic("unhandled check in doCheck")
	}
	return false
}
func validDayOfWeek(dayOfWeek string, err error) (string, error) {
	if len(dayOfWeek) != 3 {
		return "", errors.New("invalid day of week")
	}
	// XXX(toshok) validate against a list?
	return dayOfWeek, nil
}

func validMonth(month string, err error) (string, error) {
	if len(month) != 3 {
		return "", errors.New("invalid month")
	}
	// XXX(toshok) validate against a list?
	return month, nil
}

func (p *LogLineParser) validOperationName(s string) bool {
	return s == "query" ||
		s == "getmore" ||
		s == "insert" ||
		s == "update" ||
		s == "remove" ||
		s == "command" ||
		s == "killcursors"
}

func (p *LogLineParser) validComponentName(s string) bool {
	return s == "ACCESS" ||
		s == "COMMAND" ||
		s == "CONTROL" ||
		s == "GEO" ||
		s == "INDEX" ||
		s == "NETWORK" ||
		s == "QUERY" ||
		s == "REPL" ||
		s == "SHARDING" ||
		s == "STORAGE" ||
		s == "JOURNAL" ||
		s == "WRITE" ||
		s == "TOTAL" ||
		s == "-"
}
