package htjson

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tsts := []struct {
		strftime string
		expected string
	}{
		{"%Y-%m-%d %H:%M:%S", "2006-01-02 15:04:05"},
		{"%Y-%m-%d %H:%M", "2006-01-02 15:04"},
		{"%Y-%m-%d %k:%M", "2006-01-02 15:04"},
		{"%m/%d/%y %I:%M %p", "01/02/06 03:04 PM"},
		{"%m/%d/%y %I:%M %P%t%z", "01/02/06 03:04 pm\t-0700"},
		{"%a %B %d %C %r", "Mon January 02 06 03:04:05 PM"},
		{"%c %G %g %O %u %V %w %X", "       "},
		{"%H:%M:%S.%f", "15:04:05.999"},
	}

	for _, tt := range tsts {
		gotime := convertTimeFormat(tt.strftime)
		if gotime != tt.expected {
			t.Errorf("strftime format '%s' was parsed to go time '%s', expected '%s'",
				tt.strftime, gotime, tt.expected)
		}
	}
}
