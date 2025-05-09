package probe

import (
	"context"

	"github.com/guarandoo/neko/pkg/core"
)

type Probe interface {
	Probe(context.Context, string, string) (*core.Result, error)
}

type ProbeOptions struct {
}
