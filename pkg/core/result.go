package core

type Result struct {
	Tests []Test
}

type Test struct {
	Target string
	Status Status
	Error  error
}
