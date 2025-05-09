package probe

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/guarandoo/neko/pkg/core"
	"github.com/openrdap/rdap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const DomainProbeType string = "domain"

var (
	onceInitDomainProbe         sync.Once
	metricsDomainRemainingHours *prometheus.GaugeVec
)

func initDomainProbe() {
	metricsDomainRemainingHours = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_domain_remaining_hours",
	}, []string{"instance", "monitor", "type", "domain"})
}

type domainProbe struct {
	domain    string
	client    *rdap.Client
	threshold time.Duration
}

func (p *domainProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	reqRaw := rdap.Request{
		Type:  rdap.DomainRequest,
		Query: p.domain,
	}
	req := reqRaw.WithContext(ctx)
	res, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}
	if resErr, ok := res.Object.(*rdap.Error); ok {
		return nil, fmt.Errorf("error response received: %v", resErr)
	}

	info, ok := res.Object.(*rdap.Domain)
	if !ok {
		return nil, errors.New("response returned unexpected type")
	}

	index := slices.IndexFunc(info.Events, func(e rdap.Event) bool { return e.Action == "expiration" })
	if index == -1 {
		err := errors.New("")
		return nil, err
	}

	expiration, err := time.Parse(time.RFC3339, info.Events[index].Date)
	if err != nil {
		return nil, err
	}

	remaining := time.Since(expiration).Abs()

	extras := make(map[string]any)
	extras["remaining"] = remaining

	metricsDomainRemainingHours.WithLabelValues(instance, monitor, DomainProbeType, p.domain).Set(remaining.Hours())

	test := core.Test{
		Target: p.domain,
		Status: core.StatusDown,
		Error:  nil,
		Extras: extras,
	}

	if remaining > p.threshold {
		test.Status = core.StatusUp
	}

	tests := []core.Test{}
	tests = append(tests, test)

	return &core.Result{Tests: tests}, nil
}

type DomainProbeOptions struct {
	ProbeOptions
	Domain    string
	Threshold time.Duration
}

func NewDomainProbe(options DomainProbeOptions) (Probe, error) {
	onceInitDomainProbe.Do(initDomainProbe)

	client := &rdap.Client{}
	instance := domainProbe{
		domain:    options.Domain,
		client:    client,
		threshold: options.Threshold,
	}
	return &instance, nil
}
