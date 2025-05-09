package probe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/guarandoo/neko/pkg/core"
)

const PingProbeType string = "ping"

var onceInitPingProbe sync.Once

func initPingProbe() {

}

type pingProbe struct {
	address             string
	count               int
	interval            time.Duration
	packetLossThreshold float64
	privileged          bool
}

func (p *pingProbe) Probe(ctx context.Context, instance string, monitor string) (*core.Result, error) {
	ips, err := net.LookupIP(p.address)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{
			Target: ip.String(),
			Status: core.StatusUp,
			Extras: make(map[string]any),
		}

		pinger, err := probing.NewPinger(ip.String())
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}

		pinger.Interval = p.interval
		pinger.OnFinish = func(stats *probing.Statistics) {
			test.Extras["rtt_avg"] = stats.AvgRtt
			test.Extras["rtt_max"] = stats.MaxRtt
			test.Extras["rtt_min"] = stats.MinRtt
			test.Extras["rtt_stdev"] = stats.StdDevRtt
			test.Extras["packet_loss"] = stats.PacketLoss
			test.Extras["packets_received"] = stats.PacketsRecv
			test.Extras["packets_received_duplicates"] = stats.PacketsRecvDuplicates
			test.Extras["packets_sent"] = stats.PacketsSent

			if stats.PacketLoss > p.packetLossThreshold {
				test.Status = core.StatusDown
				test.Error = errors.New("packet loss")
			}
		}

		pinger.Count = p.count
		pinger.SetPrivileged(p.privileged)
		err = pinger.RunWithContext(ctx)
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}

		tests = append(tests, test)
	}

	return &core.Result{Tests: tests}, nil
}

type PingProbeOptions struct {
	ProbeOptions
	Address             string
	Count               int
	Interval            time.Duration
	PacketLossThreshold float64
	Privileged          bool
}

func NewPingProbe(options PingProbeOptions) (Probe, error) {
	onceInitPingProbe.Do(initPingProbe)

	return &pingProbe{
		address:             options.Address,
		count:               options.Count,
		interval:            options.Interval,
		packetLossThreshold: options.PacketLossThreshold,
		privileged:          options.Privileged,
	}, nil
}
