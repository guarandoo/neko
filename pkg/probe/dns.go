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
	A    RecordType = "A"
	AAAA RecordType = "AAAA"
	NS   RecordType = "NS"
)

type dnsProbe struct {
	server  net.IP
	port    uint16
	timeout time.Duration
	target  string
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

	tests = append(tests, test)
	return &core.Result{Tests: tests}, nil
}

type DnsProbeOptions struct {
	Server  string
	Port    uint16
	Timeout time.Duration
	Target  string
}

func NewDnsProbe(options DnsProbeOptions) (Probe, error) {
	server := net.ParseIP(options.Server)
	if server == nil {
		return nil, fmt.Errorf("unable to parse DNS server address")
	}

	instance := dnsProbe{
		server:  server,
		port:    options.Port,
		timeout: options.Timeout,
		target:  options.Target,
	}
	return &instance, nil
}
