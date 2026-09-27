package util

import (
	"context"
	"net"
)

type Resolver interface {
	LookupIP(ctx context.Context, network string, host string) ([]net.IP, error)
}

type systemResolver struct {
	overrides map[string][]net.IP
	resolver  *net.Resolver
}

func NewDefaultResolver(overrides map[string][]net.IP) *systemResolver {
	resolver := &systemResolver{
		overrides: overrides,
	}
	return resolver
}

func (r *systemResolver) LookupIP(ctx context.Context, network string, host string) ([]net.IP, error) {
	if ips, found := r.overrides[host]; found {
		return ips, nil
	}

	return r.LookupIP(ctx, network, host)
}
