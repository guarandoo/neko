package probe

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"sync"

	"github.com/guarandoo/neko/pkg/core"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
)

type ExecMetricCollector struct {
	mu            sync.RWMutex
	cachedMetrics []prometheus.Metric
}

func (c *ExecMetricCollector) Describe(ch chan<- *prometheus.Desc) {}

func (c *ExecMetricCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, m := range c.cachedMetrics {
		ch <- m
	}
}

const ExecProbeType string = "exec"

var onceInitExecProbe sync.Once

func initExecProbe() {
}

type execProbe struct {
	name            string
	args            []string
	enableMetrics   bool
	metricCollector *ExecMetricCollector
}

func (p *execProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	tests := []core.Test{}
	test := core.Test{Status: core.StatusUp, Target: p.name}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, p.name, p.args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	cmdErr := cmd.Run()
	if cmdErr != nil {
		test.Status = core.StatusDown
		test.Error = cmdErr
	}

	extras := make(map[string]any)

	stdoutStr := stdout.String()
	extras["Stdout"] = stdoutStr
	extras["Stderr"] = stderr.String()

	test.Extras = extras

	if p.enableMetrics && cmdErr == nil {
		parser := expfmt.NewTextParser(model.UTF8Validation)
		families, err := parser.TextToMetricFamilies(strings.NewReader(stdoutStr))
		if err == nil {
			metrics := []prometheus.Metric{}
			for _, family := range families {
				for _, m := range family.GetMetric() {
					var metricType prometheus.ValueType
					var metricValue float64
					switch family.GetType() {
					case dto.MetricType_COUNTER:
						metricType = prometheus.CounterValue
						metricValue = m.GetCounter().GetValue()
					case dto.MetricType_GAUGE:
						metricType = prometheus.GaugeValue
						metricValue = m.GetGauge().GetValue()
					default:
						continue
					}

					metric, err := prometheus.NewConstMetric(
						prometheus.NewDesc(family.GetName(), family.GetHelp(), nil, nil),
						metricType,
						metricValue,
					)
					if err == nil {
						metrics = append(metrics, metric)
					}
				}
			}

			p.metricCollector.mu.Lock()
			p.metricCollector.cachedMetrics = metrics
			p.metricCollector.mu.Unlock()
		}
	}

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
}

type ExecProbeOptions struct {
	ProbeOptions
	Name          string
	Args          []string
	EnableMetrics bool
}

func NewExecProbe(options ExecProbeOptions) (Probe, error) {
	onceInitExecProbe.Do(initExecProbe)

	metricCollector := &ExecMetricCollector{}
	prometheus.MustRegister(metricCollector)

	return &execProbe{
		name:            options.Name,
		args:            options.Args,
		enableMetrics:   options.EnableMetrics,
		metricCollector: metricCollector,
	}, nil
}
