package core

import "fmt"

type Status int

const (
	StatusPending Status = iota
	StatusUp
	StatusDown
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "Unknown"
	case StatusUp:
		return "Up"
	case StatusDown:
		return "Down"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}
