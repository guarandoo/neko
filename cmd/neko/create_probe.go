package main

import (
	"errors"
	"fmt"

	"github.com/guarandoo/neko/pkg/probe"
)

func createProbe(pc *ProbeConfig) (probe.Probe, error) {
	var p probe.Probe
	var err error

	switch v := pc.Config.(type) {
	case ExecProbeTypeConfig:
		p, err = probe.NewExecProbe(probe.ExecProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Name:         v.Path,
			Args:         v.Args,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case PingProbeTypeConfig:
		p, err = probe.NewPingProbe(probe.PingProbeOptions{
			ProbeOptions:        probe.ProbeOptions{},
			Address:             v.Address,
			Count:               v.Count,
			PacketLossThreshold: v.PacketLossThreshold,
			Privileged:          v.Privileged,
			Interval:            v.Interval,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case HttpProbeTypeConfig:
		p, err = probe.NewHttpProbe(probe.HttpProbeOptions{
			ProbeOptions:       probe.ProbeOptions{},
			SocketPath:         v.SocketPath,
			Url:                v.Address,
			MaxRedirects:       v.MaxRedirects,
			SuccessStatusCodes: v.SuccessStatusCodes,
			Headers:            v.Headers,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case SshProbeTypeConfig:
		var auth any
		if c, ok := v.Authentication.Configuration.(SshProbeTypePasswordAuthConfig); ok {
			auth = probe.SshProbePasswordAuthMethodOptions{
				Password: c.Password,
			}
		} else if c, ok := v.Authentication.Configuration.(SshProbeTypePubkeyAuthConfig); ok {
			auth = probe.SshProbeKeyAuthMethodOptions{
				PrivateKey: []byte(c.PrivKey),
			}
		} else {
			return nil, errors.New("unknown ssh authentication type")
		}

		p, err = probe.NewSshProbe(probe.SshProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Host:         v.Host,
			Port:         v.Port,
			HostKey:      v.HostKey,
			Authentication: probe.SshProbeAuthOptions{
				User:   v.Authentication.User,
				Method: auth,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case DomainProbeTypeConfig:
		p, err = probe.NewDomainProbe(probe.DomainProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Domain:       v.Domain,
			Threshold:    v.Threshold,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case DnsProbeTypeConfig:
		p, err = probe.NewDnsProbe(probe.DnsProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Server:       v.Server,
			Port:         uint16(v.Port),
			Target:       v.Target,
			RecordType:   v.RecordType,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case SqlProbeTypeConfig:
		p, err = probe.NewSqlProbe(probe.SqlProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Driver:       v.Driver,
			DSN:          v.DSN,
			Query:        v.Query,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case SmbProbeTypeConfig:
		p, err = probe.NewSmbProbe(probe.SmbProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Host:         v.Host,
			User:         v.User,
			Password:     v.Password,
			Share:        v.Share,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	default:
		p = nil
		err = fmt.Errorf("unknown probe type: %s", pc.Type)
	}

	return p, err
}
