package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/guarandoo/neko/pkg/probe"
)

func createProbe(pc *ProbeConfig) (probe.Probe, error) {
	var p probe.Probe
	var err error

	switch v := pc.Config.(type) {
	case ExecProbeTypeConfig:
		// #region defaults
		// endregion

		p, err = probe.NewExecProbe(probe.ExecProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Name:         v.Path,
			Args:         v.Args,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case PingProbeTypeConfig:
		// #region defaults
		count := 1
		if v.Count != nil {
			count = *v.Count
		}

		packetLossThreshold := 0.0
		if v.PacketLossThreshold != nil {
			packetLossThreshold = *v.PacketLossThreshold
		}

		privileged := false
		if v.Privileged != nil {
			privileged = *v.Privileged
		}

		interval := time.Second * 1
		if v.Interval != nil {
			interval, err = time.ParseDuration(*v.Interval)
			if err != nil {
				return nil, err
			}
		}
		// #endregion

		p, err = probe.NewPingProbe(probe.PingProbeOptions{
			ProbeOptions:        probe.ProbeOptions{},
			Address:             v.Address,
			Count:               count,
			PacketLossThreshold: packetLossThreshold,
			Privileged:          privileged,
			Interval:            interval,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case HttpProbeTypeConfig:
		// region defaults
		maxRedirects := 20
		if v.MaxRedirects != nil {
			if *v.MaxRedirects < 0 {
				return nil, errors.New("MaxRedirects must be a positive number")
			}
			maxRedirects = *v.MaxRedirects
		}

		successStatusCodes := []int{200}
		if v.SuccessStatusCodes != nil {
			successStatusCodes = *v.SuccessStatusCodes
		}

		headers := make(map[string]string)
		if v.Headers != nil {
			headers = *v.Headers
		}
		// endregion

		p, err = probe.NewHttpProbe(probe.HttpProbeOptions{
			ProbeOptions:       probe.ProbeOptions{},
			SocketPath:         v.SocketPath,
			Url:                v.Address,
			MaxRedirects:       maxRedirects,
			SuccessStatusCodes: successStatusCodes,
			Headers:            headers,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case SshProbeTypeConfig:
		// #region defaults
		port := 22
		if v.Port != nil {
			port = *v.Port
		}
		// endregion

		p, err = probe.NewSshProbe(probe.SshProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Host:         v.Host,
			Port:         port,
			HostKey:      v.HostKey,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case DomainProbeTypeConfig:
		// #region defaults
		threshold := time.Hour * 24
		if v.Threshold != nil {
			threshold, err = time.ParseDuration(*v.Threshold)
			if err != nil {
				return nil, err
			}
		}
		// endregion

		p, err = probe.NewDomainProbe(probe.DomainProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Domain:       v.Domain,
			Threshold:    threshold,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create probe: %w", err)
		}

	case DnsProbeTypeConfig:
		// #region defaults
		port := 53
		if v.Port != nil {
			port = *v.Port
			if !(port > 0 && port <= 65535) {
				return nil, errors.New("invalid port")
			}
		}

		recordType := probe.Host
		if v.RecordType != nil {
			recordType = *v.RecordType
		}
		// endregion

		p, err = probe.NewDnsProbe(probe.DnsProbeOptions{
			ProbeOptions: probe.ProbeOptions{},
			Server:       v.Server,
			Port:         uint16(port),
			Target:       v.Target,
			RecordType:   recordType,
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
