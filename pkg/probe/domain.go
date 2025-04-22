package probe

import (
	"errors"
	"slices"
	"time"

	"github.com/guarandoo/neko/pkg/core"
	"github.com/openrdap/rdap"
)

type domainProbe struct {
	domain    string
	timeout   int
	client    *rdap.Client
	threshold time.Duration
}

func (p *domainProbe) Probe() (*core.Result, error) {
	info, err := p.client.QueryDomain(p.domain)
	if err != nil {
		return nil, err
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
	Domain    string
	Timeout   int
	Threshold time.Duration
}

func NewDomainProbe(options DomainProbeOptions) (Probe, error) {
	client := &rdap.Client{}
	instance := domainProbe{
		domain:    options.Domain,
		timeout:   options.Timeout,
		client:    client,
		threshold: options.Threshold,
	}
	return &instance, nil
}
