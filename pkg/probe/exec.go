package probe

import (
	"context"
	"os/exec"
	"sync"

	"github.com/guarandoo/neko/pkg/core"
)

type execProbe struct {
	name string
	args []string
}

func (p *execProbe) Probe(ctx context.Context) (*core.Result, error) {
	tests := []core.Test{}
	test := core.Test{Status: core.StatusUp, Target: p.name}

	cmd := exec.CommandContext(ctx, p.name, p.args...)
	err := cmd.Run()
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
	}

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
}

var onceInitExecProbe sync.Once

func initExecProbe() {

}

type ExecProbeOptions struct {
	ProbeOptions
	Name string
	Args []string
}

func NewExecProbe(options ExecProbeOptions) (Probe, error) {
	onceInitExecProbe.Do(initExecProbe)

	return &execProbe{
		name: options.Name,
		args: options.Args,
	}, nil
}
