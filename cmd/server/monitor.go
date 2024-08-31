package main

import (
	"github.com/guarandoo/neko/pkg/core"
	"github.com/guarandoo/neko/pkg/notifier"
	"github.com/guarandoo/neko/pkg/probe"
)

type Monitor struct {
	Name      string
	Interval  string
	Probe     probe.Probe
	Status    core.Status
	Notifiers []notifier.Notifier
}
