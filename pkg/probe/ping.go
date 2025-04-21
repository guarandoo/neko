package probe

import (
	"errors"
	"fmt"
	"net"

	probing "github.com/prometheus-community/pro-bing"

	"github.com/guarandoo/neko/pkg/core"
)

type pingProbe struct {
	address string
}

func (p *pingProbe) Probe() (*core.Result, error) {
	ips, err := net.LookupIP(p.address)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup domain: %w", err)
	}

	tests := []core.Test{}
	for _, ip := range ips {
		test := core.Test{Target: ip.String(), Status: core.StatusUp}

		pinger, err := probing.NewPinger(ip.String())
		if err != nil {
			test.Status = core.StatusDown
			test.Error = err
			tests = append(tests, test)
			continue
		}

		pinger.OnFinish = func(stats *probing.Statistics) {
			test.Extras["rtt_avg"] = stats.AvgRtt
			test.Extras["rtt_max"] = stats.MaxRtt
			test.Extras["rtt_min"] = stats.MinRtt
			test.Extras["rtt_stdev"] = stats.StdDevRtt
			test.Extras["packet_loss"] = stats.PacketLoss
			test.Extras["packets_received"] = stats.PacketsRecv
			test.Extras["packets_received_duplicates"] = stats.PacketsRecvDuplicates
			test.Extras["packets_sent"] = stats.PacketsSent

			if stats.PacketLoss > 0.0 {
				test.Status = core.StatusDown
				test.Error = errors.New("packet loss")
				tests = append(tests, test)
			}
		}

		pinger.Count = 1
		pinger.SetPrivileged(true)
		err = pinger.Run()
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
	Address string
}

func NewPingProbe(options PingProbeOptions) (Probe, error) {
	return &pingProbe{address: options.Address}, nil
}
