package core

import "fmt"

type Result struct {
	Tests []Test
}

type Test struct {
	Target string
	Status Status
	Error  error
	Extras map[string]any
}

func (t Test) String() string {
	return fmt.Sprintf("Test{Target: %s, Status: %s, Error: %v}", t.Target, t.Status, t.Error)
}
