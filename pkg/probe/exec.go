package probe

import (
	"os/exec"

	"github.com/guarandoo/neko/pkg/core"
)

type execProbe struct {
	name string
	args []string
}

func (p *execProbe) Probe() (*core.Result, error) {
	tests := []core.Test{}
	test := core.Test{Status: core.StatusUp, Target: p.name}

	cmd := exec.Command(p.name, p.args...)
	err := cmd.Run()
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
	}

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
}

type ExecProbeOptions struct {
	Name string
	Args []string
}

func NewExecProbe(options ExecProbeOptions) (Probe, error) {
	return &execProbe{name: options.Name, args: options.Args}, nil
}
