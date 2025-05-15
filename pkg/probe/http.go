package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/guarandoo/neko/pkg/core"
)

const HttpProbeType string = "http"

var onceInitHttpProbe sync.Once

func initHttpProbe() {

}

var (
	ErrInvalidUrlScheme = errors.New("invalid url scheme")
)

type httpProbe struct {
	socketPath         string
	timeout            time.Duration
	url                url.URL
	method             string
	maxRedirects       int
	successStatusCodes []int
	headers            map[string]string
}

func (p *httpProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	req, err := http.NewRequestWithContext(ctx, p.method, p.url.String(), nil)
	if err != nil {
		return nil, err
	}

	for key, value := range p.headers {
		req.Header.Add(key, value)
	}

	tests := []core.Test{}

	type Target struct {
		target    string
		transport *http.Transport
	}

	targets := []Target{}

	if len(p.socketPath) > 0 {
		transport := &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", p.socketPath)
			},
		}
		targets = append(targets, Target{
			target:    p.socketPath,
			transport: transport,
		})
	} else {
		getTransportIp := func(ip net.IP) *http.Transport {
			dialer := &net.Dialer{}
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
				parts := strings.Split(addr, ":")
				if parts[0] == p.url.Host {
					if ip.To4() != nil {
						addr = fmt.Sprintf("%s:%s", ip, parts[1])
					} else {
						addr = fmt.Sprintf("[%s]:%s", ip, parts[1])
					}
				}
				return dialer.DialContext(ctx, network, addr)
			}

			return transport
		}

		r := net.Resolver{}
		ips, err := r.LookupIP(ctx, "ip", p.url.Hostname())
		if err != nil {
			return nil, fmt.Errorf("unable to lookup domain: %w", err)
		}
		for _, ip := range ips {
			targets = append(targets, Target{
				target:    ip.String(),
				transport: getTransportIp(ip),
			})
		}
	}

	for _, target := range targets {
		test := core.Test{Target: target.target, Status: core.StatusUp}

		redirectCount := 0
		client := http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				redirectCount++
				if p.maxRedirects >= 0 && redirectCount > p.maxRedirects {
					return fmt.Errorf("exceeded allowed redirect count %v", p.maxRedirects)
				}

				return nil
			},
		}
		client.Transport = target.transport
		client.Timeout = time.Duration(p.timeout) * time.Second

		res, err := client.Do(req)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}

		if !slices.Contains(p.successStatusCodes, res.StatusCode) {
			test.Status = core.StatusDown
			test.Error = fmt.Errorf("return code was %v", res.StatusCode)
			tests = append(tests, test)
			continue
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type HttpProbeOptions struct {
	ProbeOptions
	SocketPath         string
	Timeout            time.Duration
	Url                string
	Method             string
	MaxRedirects       int
	SuccessStatusCodes []int
	Headers            map[string]string
}

func NewHttpProbe(options HttpProbeOptions) (Probe, error) {
	onceInitHttpProbe.Do(initHttpProbe)

	u, err := url.ParseRequestURI(options.Url)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrInvalidUrlScheme
	}

	instance := httpProbe{
		socketPath:         options.SocketPath,
		timeout:            options.Timeout,
		url:                *u,
		method:             options.Method,
		maxRedirects:       options.MaxRedirects,
		successStatusCodes: options.SuccessStatusCodes,
		headers:            options.Headers,
	}
	return &instance, nil
}
