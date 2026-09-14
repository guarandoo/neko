package probe

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	server    *url.URL
	domains   []string
	client    *rdap.Client
	threshold time.Duration
}

func (p *domainProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	tests := []core.Test{}

	for _, domain := range p.domains {
		test := core.Test{
			Target: domain,
			Status: core.StatusUp,
			Error:  nil,
			Extras: make(map[string]any),
		}

		reqRaw := rdap.Request{
			Server: p.server,
			Type:   rdap.DomainRequest,
			Query:  domain,
		}
		req := reqRaw.WithContext(ctx)
		res, err := p.client.Do(req)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}
		if resErr, ok := res.Object.(*rdap.Error); ok {
			test.Status = core.StatusDown
			test.Error = fmt.Errorf("%v", resErr)
			tests = append(tests, test)
			continue
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
		test.Extras["remaining"] = remaining

		metricsDomainRemainingHours.WithLabelValues(instance, monitor, DomainProbeType, domain).Set(remaining.Hours())

		if remaining < p.threshold {
			test.Status = core.StatusDown
			test.Error = fmt.Errorf("domain expiration %s is below threshold %s", remaining, p.threshold)
			tests = append(tests, test)
			continue
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type DomainProbeOptions struct {
	ProbeOptions
	Server    string
	Domains   []string
	Threshold time.Duration
}

func NewDomainProbe(options DomainProbeOptions) (Probe, error) {
	onceInitDomainProbe.Do(initDomainProbe)

	var server *url.URL

	if options.Server != "" {
		var err error
		server, err = url.Parse(options.Server)
		if err != nil {
			return nil, err
		}
	}

	client := &rdap.Client{}
	instance := domainProbe{
		server:    server,
		domains:   options.Domains,
		client:    client,
		threshold: options.Threshold,
	}
	return &instance, nil
}
