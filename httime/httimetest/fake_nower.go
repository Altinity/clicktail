package httimetest

import (
	"time"
)

type FakeNower struct {
	FakeNow time.Time
}

func (f *FakeNower) Now() time.Time {
	if !f.FakeNow.IsZero() {
		return f.FakeNow
	}

	// default fake time to return
	fakeNow, _ := time.Parse(time.RFC3339, "2010-06-21T15:04:05Z")
	f.FakeNow = fakeNow
	return f.FakeNow
}
