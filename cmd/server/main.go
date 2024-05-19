package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/maps"

	"github.com/guarandoo/neko/pkg/notifier"
	"github.com/guarandoo/neko/pkg/probe"
	"github.com/hashicorp/raft"
	"github.com/mxmauro/resetevent"
	yaml "gopkg.in/yaml.v2"
)

func CreateRaft() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("ASD")

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1")
	if err != nil {
		return fmt.Errorf("unable to resolve raft bind addr: %w", err)
	}

	transport, err := raft.NewTCPTransport("", addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return fmt.Errorf("unable to create raft tcp transport: %w", err)
	}

	snapshots, err := raft.NewFileSnapshotStore("raft_dir", 10, os.Stderr)
	if err != nil {
		return fmt.Errorf("unable to create file snapshot store: %w", err)
	}

	var logStore raft.LogStore
	var stableStore raft.StableStore

	logStore = raft.NewInmemStore()
	stableStore = raft.NewInmemStore()

	ra, err := raft.NewRaft(config, nil, logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("unable to create raft: %w", err)
	}

	if ra == nil {
		return err
	}

	return nil
}

func createProbe(pc *ProbeConfig) (probe.Probe, error) {
	var p probe.Probe
	var err error
	switch v := pc.Config.(type) {
	case ExecProbeConfig:
		p, err = probe.NewExecProbe(probe.ExecProbeOptions{Name: v.Path, Args: v.Args})
	case PingProbeConfig:
		p, err = probe.NewPingProbe(probe.PingProbeOptions{Address: v.Address})
	case HttpProbeConfig:
		p, err = probe.NewHttpProbe(probe.HttpProbeOptions{Url: v.Address})
	default:
		p = nil
		err = errors.New(fmt.Sprintf("unknown probe type: %s", pc.Type))
	}
	return p, err
}

func createNotifier(nc *NotifierConfig) (notifier.Notifier, error) {
	var n notifier.Notifier
	var err error
	switch v := nc.Config.(type) {
	case SmtpNotifierCOnfig:
		n, err = notifier.NewSmtpNotifier(notifier.SmtpNotifierOptions{
			Host:       v.Host,
			Port:       v.Port,
			Username:   v.Username,
			Password:   v.Password,
			Sender:     v.Sender,
			Recipients: v.Recipients,
		})
	case DiscordWebhookNotifierConfig:
		n, err = notifier.NewDiscordWebhookNotifier(notifier.DiscordWebhookOptions{Url: v.Url})
	default:
		n = nil
		err = errors.New(fmt.Sprintf("unknown probe type: %s", nc.Type))
	}
	return n, err
}

func main() {
	cfg := os.Getenv("NEKO_CONFIG")
	if len(cfg) == 0 {
		cfg = "config.yaml"
	}

	quit := make(chan os.Signal, 1)
	start := resetevent.NewManualResetEvent()

	filename, err := filepath.Abs(cfg)
	if err != nil {
		log.Fatalf("unable to get filename: %s", err)
	}

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("config file does not exist: %s", err)
	}

	text, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("unable to read config file: %s", err)
	}

	var config Configuration
	err = yaml.Unmarshal(text, &config)

	if err != nil {
		log.Fatalf("unable to parse config: %s", err)
	}

	notifiers := map[string]notifier.Notifier{}
	for k, v := range config.Notifiers {
		n, err := createNotifier(&v)
		if err != nil {
			log.Fatalf("unable to create notifier: %s", err)
		}

		notifiers[k] = n
	}

	monitors := []Monitor{}
	for _, m := range config.Monitors {
		p, err := createProbe(&m.Probe)
		if err != nil {
			log.Fatalf("unable to create probe: %s", err)
		}
		monitor := Monitor{
			Name:      m.Name,
			Interval:  m.Interval,
			Probe:     p,
			Notifiers: maps.Values(notifiers),
		}
		monitors = append(monitors, monitor)
	}

	for _, m := range monitors {
		log.Printf("setting up monitor %s", m.Name)
		go func(monitor Monitor) {
			err := start.Wait(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			ticker := time.NewTicker(time.Duration(monitor.Interval) * time.Second)
			log.Printf("starting monitor %s", monitor.Name)
			for {
				<-ticker.C
				log.Printf("running monitor %v", monitor.Name)
				res, err := monitor.Probe.Probe()
				if err != nil {
					log.Printf("monitor %v failed: %s", monitor.Name, err)
				}
				log.Printf("Probe %v completed with result: %v", monitor.Name, res.Tests)

				json, err := json.Marshal(res.Tests)
				if err != nil {
					continue
				}
				for _, n := range monitor.Notifiers {
					if err := n.Notify(monitor.Name, string(json)); err != nil {
						log.Printf("unable to notify: %s", err)
					}
				}
			}
		}(m)
	}

	// err := CreateRaft()

	// storage := raft.NewMemoryStorage()
	// c := &raft.Config{
	// 	ID:              0x01,
	// 	ElectionTick:    10,
	// 	HeartbeatTick:   1,
	// 	Storage:         storage,
	// 	MaxSizePerMsg:   4096,
	// 	MaxInflightMsgs: 256,
	// }

	// n := raft.StartNode(c, nil)

	log.Println("initialization complete, releasing monitors")
	start.Set()
	<-quit
}
