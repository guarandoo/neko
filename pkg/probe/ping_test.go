package probe

import (
	"context"
	"testing"
	"time"

	"github.com/guarandoo/neko/pkg/core"
)

func TestPingProbe(t *testing.T) {
	probe, err := NewPingProbe(PingProbeOptions{
		Address:             "127.0.0.1",
		Count:               3,
		PacketLossThreshold: 0.0,
		Interval:            time.Second * 1,
		ProbeOptions:        ProbeOptions{},
		Privileged:          true,
	})
	if err != nil {
		t.Errorf("NewPingProbe failed: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*4)
	defer cancel()

	res, err := probe.Probe(ctx)
	if err != nil {
		t.Errorf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Error("unexpected test count")
	}

	test := res.Tests[0]
	if test.Status != core.StatusUp {
		t.Errorf("probe returned unexpected status, found: %v expecting %v", test.Status, core.StatusUp)
	}
}

func TestPingProbeFail(t *testing.T) {
	probe, err := NewPingProbe(PingProbeOptions{
		Address:             "192.0.2.0",
		Count:               3,
		PacketLossThreshold: 0.0,
		Interval:            time.Second * 1,
		ProbeOptions:        ProbeOptions{},
	})
	if err != nil {
		t.Errorf("NewPingProbe failed: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*4)
	defer cancel()

	res, err := probe.Probe(ctx)
	if err != nil {
		t.Errorf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Error("unexpected test count")
	}

	test := res.Tests[0]
	if test.Status != core.StatusDown {
		t.Error("probe returned unexpected status")
	}
}

func TestPingProbeTimeout(t *testing.T) {
	probe, err := NewPingProbe(PingProbeOptions{
		Address:             "192.0.2.0",
		Count:               1,
		PacketLossThreshold: 0.0,
		Interval:            time.Second * 3,
		ProbeOptions:        ProbeOptions{},
	})
	if err != nil {
		t.Errorf("NewPingProbe failed: %v", err)
	}

	ctx, cancel := getContextWithTimeout(context.Background(), time.Second*4)
	defer cancel()

	res, err := probe.Probe(ctx)
	if err != nil {
		t.Errorf("probe failed: %v", err)
	}

	if len(res.Tests) != 1 {
		t.Error("unexpected test count")
	}

	test := res.Tests[0]
	if test.Status != core.StatusDown {
		t.Errorf("probe returned unexpected status, found: %v expecting %v", test.Status, core.StatusDown)
	}
}
