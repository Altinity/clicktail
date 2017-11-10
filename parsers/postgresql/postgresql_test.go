package postgresql

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/honeycombio/honeytail/event"
	"github.com/stretchr/testify/assert"
)

// Test parsing individual log statements with different prefix formats.
func TestSingleQueryParsing(t *testing.T) {
	testcases := []struct {
		description  string
		in           string
		prefixFormat string
		expected     event.Event
	}{
		{
			description: "parse a multiline log statement from a default postgres 9.5 log",
			in: `2017-11-07 23:05:16 UTC [3053-3] postgres@postgres LOG:  duration: 0.681 ms  statement: SELECT d.datname as "Name",
	       pg_catalog.pg_get_userbyid(d.datdba) as "Owner",
	       pg_catalog.pg_encoding_to_char(d.encoding) as "Encoding",
	       d.datcollate as "Collate",
	       d.datctype as "Ctype",
	       pg_catalog.array_to_string(d.datacl, E'\n') AS "Access privileges"
	FROM pg_catalog.pg_database d
	ORDER BY 1;`,
			prefixFormat: "%t [%p-%l] %u@%d",
			expected: event.Event{
				Timestamp: time.Date(2017, 11, 7, 23, 5, 16, 0, time.UTC),
				Data: map[string]interface{}{
					"user":     "postgres",
					"database": "postgres",
					"duration": 0.681,
					"pid":      3053,
					"session_line_number": 3,
					"query":               "SELECT d.datname as \"Name\", pg_catalog.pg_get_userbyid(d.datdba) as \"Owner\", pg_catalog.pg_encoding_to_char(d.encoding) as \"Encoding\", d.datcollate as \"Collate\", d.datctype as \"Ctype\", pg_catalog.array_to_string(d.datacl, E'\\n') AS \"Access privileges\" FROM pg_catalog.pg_database d ORDER BY 1;",
					"normalized_query":    "select d.datname as ?, pg_catalog.pg_get_userbyid(d.datdba) as ?, pg_catalog.pg_encoding_to_char(d.encoding) as ?, d.datcollate as ?, d.datctype as ?, pg_catalog.array_to_string(d.datacl, e?) as ? from pg_catalog.pg_database d order by ?;",
				},
			},
		},
		{
			description:  "extract everything you can put in a line prefix",
			in:           `2017-11-08 03:02:49.314 UTC [8544-1] postgres@test (3/0) (0) (00000) (2017-11-08 03:02:38 UTC) (psql)LOG:  duration: 2.753 ms  statement: select * from test;`,
			prefixFormat: `%m [%p-%l] %q%u@%d (%v) (%x) (%e) (%s) (%a)`,
			expected: event.Event{
				Timestamp: time.Date(2017, 11, 8, 3, 2, 49, 314000000, time.UTC),
				Data: map[string]interface{}{
					"user":     "postgres",
					"database": "test",
					"duration": 2.753,
					"pid":      8544,
					"session_line_number":    1,
					"virtual_transaction_id": "3/0",
					"transaction_id":         "0",
					"sql_state":              "00000",
					"session_start":          "2017-11-08 03:02:38 UTC",
					"application":            "psql",
					"query":                  "select * from test;",
					"normalized_query":       "select * from test;",
				},
			},
		},
		{
			description:  "extract the event timestamp from a unix time",
			in:           `1510258541402 [8544-1] postgres@test LOG:  duration: 2.753 ms  statement: select * from test;`,
			prefixFormat: "%n [%p-%l] %u@%d",
			expected: event.Event{
				Timestamp: time.Date(2017, 11, 9, 20, 15, 41, 402000000, time.UTC),
				Data: map[string]interface{}{
					"user":     "postgres",
					"database": "test",
					"duration": 2.753,
					"pid":      8544,
					"session_line_number": 1,
					"query":               "select * from test;",
					"normalized_query":    "select * from test;",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.description, func(t *testing.T) {
			in := make(chan []string)
			out := make(chan event.Event)
			p := Parser{}
			p.Init(&Options{LogLinePrefix: tc.prefixFormat})
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go p.handleEvents(in, out, wg)
			in <- strings.Split(tc.in, "\n")
			close(in)
			got := <-out
			assert.Equal(t, got, tc.expected)
		})
	}
}

// Test grouping and parsing multiple log statements.
func TestMultipleQueryParsing(t *testing.T) {
	in := `
2017-11-07 01:43:18 UTC [3542-5] postgres@test LOG:  duration: 9.263 ms  statement: INSERT INTO test (id, name, value) VALUES (1, 'Alice', 'foo');
2017-11-07 01:43:27 UTC [3542-6] postgres@test LOG:  duration: 0.841 ms  statement: INSERT INTO test (id, name, value) VALUES (2, 'Bob', 'bar');
2017-11-07 01:43:39 UTC [3542-7] postgres@test LOG:  duration: 15.577 ms  statement: SELECT * FROM test
	WHERE id=1;
2017-11-07 01:43:42 UTC [3542-8] postgres@test LOG:  duration: 0.501 ms  statement: SELECT * FROM test
	WHERE id=2;
`
	out := []event.Event{
		event.Event{
			Timestamp: time.Date(2017, 11, 7, 1, 43, 18, 0, time.UTC),
			Data: map[string]interface{}{
				"user":     "postgres",
				"database": "test",
				"duration": 9.263,
				"pid":      3542,
				"session_line_number": 5,
				"query":               "INSERT INTO test (id, name, value) VALUES (1, 'Alice', 'foo');",
				"normalized_query":    "insert into test (id, name, value) values (?, ?, ?);",
			},
		},
		event.Event{
			Timestamp: time.Date(2017, 11, 7, 1, 43, 27, 0, time.UTC),
			Data: map[string]interface{}{
				"user":     "postgres",
				"database": "test",
				"duration": 0.841,
				"pid":      3542,
				"session_line_number": 6,
				"query":               "INSERT INTO test (id, name, value) VALUES (2, 'Bob', 'bar');",
				"normalized_query":    "insert into test (id, name, value) values (?, ?, ?);",
			},
		},
		event.Event{
			Timestamp: time.Date(2017, 11, 7, 1, 43, 39, 0, time.UTC),
			Data: map[string]interface{}{
				"user":     "postgres",
				"database": "test",
				"duration": 15.577,
				"pid":      3542,
				"session_line_number": 7,
				"query":               "SELECT * FROM test WHERE id=1;",
				"normalized_query":    "select * from test where id=?;",
			},
		},
		event.Event{
			Timestamp: time.Date(2017, 11, 7, 1, 43, 42, 0, time.UTC),
			Data: map[string]interface{}{
				"user":     "postgres",
				"database": "test",
				"duration": 0.501,
				"pid":      3542,
				"session_line_number": 8,
				"query":               "SELECT * FROM test WHERE id=2;",
				"normalized_query":    "select * from test where id=?;",
			},
		},
	}

	parser := Parser{}
	parser.Init(nil)
	inChan := make(chan string)
	sendChan := make(chan event.Event, 4)
	go parser.ProcessLines(inChan, sendChan, nil)
	for _, line := range strings.Split(in, "\n") {
		inChan <- line
	}
	close(inChan)
	for _, expected := range out {
		got := <-sendChan
		assert.Equal(t, got, expected)
	}
}

// Test handling log statements that aren't slow query logs
func TestSkipNonQueryLogLines(t *testing.T) {
	parser := Parser{}
	parser.Init(nil)
	testcases := []string{
		"la la la",
		"2017-11-06 19:20:32 UTC [11534-2] LOG:  autovacuum launcher shutting down",
		"2017-11-07 01:43:39 UTC [3542-7] postgres@test ERROR: relation \"test\" does not exist at character 15",
	}

	for _, tc := range testcases {
		lineGroup := []string{tc}
		ev := parser.handleEvent(lineGroup)
		assert.Nil(t, ev)
	}
}
