package probe

import (
	"github.com/guarandoo/neko/pkg/core"
)

type groupProbe struct {
}

func (p *groupProbe) Probe() (*core.Result, error) {
	tests := []core.Test{}
	return &core.Result{Tests: tests}, nil
}

type GroupProbeOptions struct {
}

func NewGroupProbe(options GroupProbeOptions) (Probe, error) {
	instance := groupProbe{}
	return &instance, nil
}
