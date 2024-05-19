package probe

import "github.com/guarandoo/neko/pkg/core"

type Probe interface {
	Probe() (*core.Result, error)
}
