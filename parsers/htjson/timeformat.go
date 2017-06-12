package htjson

import "strings"

const (
	StrftimeChar     = "%"
	UnixTimestampFmt = "%s"
)

var (
	// reference: http://man7.org/linux/man-pages/man3/strftime.3.html
	convertMapping = map[string]string{
		"%a": "Mon",
		"%A": "Monday",
		"%b": "Jan",
		"%B": "January",
		"%c": "", // locale not supported
		"%C": "06",
		"%d": "02",
		"%D": "01/02/06",
		"%e": "_2",
		"%E": "", // modifiers not supported
		"%f": "999",
		"%F": "2006-01-02",
		"%G": "", // week-based year not supported
		"%g": "", // week-based year not supported
		"%h": "Jan",
		"%H": "15",
		"%I": "03",
		"%j": "",   // day of year not supported
		"%k": "15", // same case as %H but accepts leading space instead of 0
		"%l": "_3",
		"%L": "999", // milliseconds
		"%m": "01",
		"%M": "04",
		"%n": "\n",
		"%O": "", // modifiers not supported
		"%p": "PM",
		"%P": "pm",
		"%r": "03:04:05 PM",
		"%R": "15:04",
		"%S": "05",
		"%t": "\t",
		"%T": "15:04:05",
		"%u": "", // day of week not supported
		"%U": "", // week number of the current year not supported
		"%V": "", // ISO 8601 week number not supported
		"%w": "", // day of week not supported
		"%W": "", // day of week not supported
		"%x": "", // date-only not supported
		"%X": "", // date-only not supported
		"%y": "06",
		"%Y": "2006",
		"%z": "-0700",
		"%Z": "MST",
		"%+": "Mon Jan _2 15:04:05 MST 2006",
	}
)

// maybeConvertTimeFormat tries to handle C-style time formats alongside Go's
// existing time.Parse behavior.
func convertTimeFormat(layout string) string {
	for fmt, conv := range convertMapping {
		layout = strings.Replace(layout, fmt, conv, -1)
	}
	return layout
}
