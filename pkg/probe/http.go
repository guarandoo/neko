package probe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/guarandoo/neko/pkg/core"
)

type httpProbe struct {
	url          url.URL
	method       string
	maxRedirects int
}

func (p *httpProbe) Probe() (*core.Result, error) {
	req, err := http.NewRequest(p.method, p.url.String(), nil)
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(p.url.Host)
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
				log.Printf("following redirect %v", redirectCount)
				if redirectCount > p.maxRedirects {
					return errors.New(fmt.Sprintf("Exceeded allowed redirect count %v", p.maxRedirects))
				}

				return nil
			},
		}
		client.Transport = http.DefaultTransport.(*http.Transport).Clone()
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}
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

		if res.StatusCode != 200 {
			test.Status = core.StatusDown
			test.Error = errors.New(fmt.Sprintf("return code was %v", res.StatusCode))
			tests = append(tests, test)
			continue
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type HttpProbeOptions struct {
	Url          string
	Method       string
	MaxRedirects int
}

func NewHttpProbe(options HttpProbeOptions) (Probe, error) {
	u, err := url.ParseRequestURI(options.Url)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("Unknown url scheme")
	}

	instance := httpProbe{
		url:          *u,
		method:       options.Method,
		maxRedirects: options.MaxRedirects,
	}
	return &instance, nil
}
