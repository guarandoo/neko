package core

import (
	"time"
)

type Datapoint struct {
	MeasuredAt time.Time
	Status     Status
}

func (s *Datapoint) blah() {
	time.Now()
}
