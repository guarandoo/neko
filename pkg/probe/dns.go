package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/guarandoo/neko/pkg/core"
)

type RecordType string

const (
	Host RecordType = "Host"
	NS   RecordType = "NS"
	MX   RecordType = "MX"
)

type dnsProbe struct {
	server     net.IP
	port       uint16
	timeout    time.Duration
	target     string
	recordType RecordType
}

func (p *dnsProbe) Probe() (*core.Result, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network string, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: p.timeout}
			return d.DialContext(ctx, network, fmt.Sprintf("%s:%d", p.server, p.port))
		},
	}

	tests := []core.Test{}

	test := core.Test{Target: p.target, Status: core.StatusUp}

	switch p.recordType {
	case Host:
		addrs, err := r.LookupHost(context.Background(), p.target)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			return nil, err
		}

		if len(addrs) == 0 {
			test.Status = core.StatusDown
			test.Error = errors.New("no records returned")

			tests = append(tests, test)
			return &core.Result{Tests: tests}, nil
		}
	case NS:
		addrs, err := r.LookupNS(context.Background(), p.target)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err

			tests = append(tests, test)
			return &core.Result{Tests: tests}, nil
		}

		if len(addrs) == 0 {
			test.Status = core.StatusDown
			test.Error = errors.New("no records returned")

			tests = append(tests, test)
			return &core.Result{Tests: tests}, nil
		}
	case MX:
		addrs, err := r.LookupMX(context.Background(), p.target)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err

			tests = append(tests, test)
			return &core.Result{Tests: tests}, nil
		}

		if len(addrs) == 0 {
			test.Status = core.StatusDown
			test.Error = errors.New("no records returned")

			tests = append(tests, test)
			return &core.Result{Tests: tests}, nil
		}
	}

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
}

type DnsProbeOptions struct {
	Server     string
	Port       uint16
	Timeout    time.Duration
	Target     string
	RecordType RecordType
}

func NewDnsProbe(options DnsProbeOptions) (Probe, error) {
	server := net.ParseIP(options.Server)
	if server == nil {
		return nil, fmt.Errorf("unable to parse DNS server address")
	}

	instance := dnsProbe{
		server:     server,
		port:       options.Port,
		timeout:    options.Timeout,
		target:     options.Target,
		recordType: options.RecordType,
	}
	return &instance, nil
}
