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
	"time"

	"github.com/guarandoo/neko/pkg/core"
)

var (
	ErrInvalidUrlScheme = errors.New("invalid url scheme")
)

type httpProbe struct {
	timeout            time.Duration
	url                url.URL
	method             string
	maxRedirects       int
	successStatusCodes []int
	headers            map[string]string
}

func (p *httpProbe) Probe(ctx context.Context) (*core.Result, error) {
	req, err := http.NewRequestWithContext(ctx, p.method, p.url.String(), nil)
	if err != nil {
		return nil, err
	}

	for key, value := range p.headers {
		req.Header.Add(key, value)
	}

	r := net.Resolver{}
	ips, err := r.LookupIP(ctx, "ip", p.url.Hostname())
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{Target: ip.String(), Status: core.StatusUp}

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
		client.Transport = http.DefaultTransport.(*http.Transport).Clone()
		client.Timeout = time.Duration(p.timeout) * time.Second
		dialer := &net.Dialer{}
		client.Transport.(*http.Transport).DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
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
	Timeout            time.Duration
	Url                string
	Method             string
	MaxRedirects       int
	SuccessStatusCodes []int
	Headers            map[string]string
}

func NewHttpProbe(options HttpProbeOptions) (Probe, error) {
	u, err := url.ParseRequestURI(options.Url)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, ErrInvalidUrlScheme
	}

	instance := httpProbe{
		timeout:            options.Timeout,
		url:                *u,
		method:             options.Method,
		maxRedirects:       options.MaxRedirects,
		successStatusCodes: options.SuccessStatusCodes,
		headers:            options.Headers,
	}
	return &instance, nil
}
