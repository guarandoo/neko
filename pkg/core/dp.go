package core

import (
	"time"
)

type Datapoint struct {
	MeasuredAt time.Time
	Status     Status
}
