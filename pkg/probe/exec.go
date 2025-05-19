package probe

import (
	"bytes"
	"context"
	"os/exec"
	"sync"

	"github.com/guarandoo/neko/pkg/core"
)

const ExecProbeType string = "exec"

var onceInitExecProbe sync.Once

func initExecProbe() {

}

type execProbe struct {
	name string
	args []string
}

func (p *execProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	tests := []core.Test{}
	test := core.Test{Status: core.StatusUp, Target: p.name}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, p.name, p.args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
	}

	test.Extras["Stdout"] = stdout.String()
	test.Extras["Stderr"] = stderr.String()

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
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
