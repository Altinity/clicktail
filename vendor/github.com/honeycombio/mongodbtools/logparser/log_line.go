package logparser

import "github.com/honeycombio/mongodbtools/logparser/internal/logparser"

// ParseLogLine attempts to parse a MongoDB log line into a structured representation
func ParseLogLine(input string) (map[string]interface{}, error) {
	return logparser.ParseLogLine(input)
}

func IsPartialLogLine(err error) bool {
	return logparser.IsPartialLogLine(err)
}

// parse just a mongodb query as it exists in the log
func ParseQuery(query string) (map[string]interface{}, error) {
	return logparser.ParseQuery(query)
}
