package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/maps"

	"github.com/guarandoo/neko/pkg/core"
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
		maxRedirects := 20
		if v.MaxRedirects != nil {
			if *v.MaxRedirects < 0 {
				return nil, errors.New("MaxRedirects must be a positive number")
			}
			maxRedirects = *v.MaxRedirects
		}
		timeout := 60
		if v.Timeout != nil {
			if *v.Timeout < 0 {
				return nil, errors.New("timeout must be a positive number")
			}
			timeout = *v.Timeout
		}
		p, err = probe.NewHttpProbe(probe.HttpProbeOptions{
			Url:          v.Address,
			MaxRedirects: maxRedirects,
			Timeout:      timeout,
		})
	case SshProbeConfig:
		port := 22
		if v.Port != nil {
			port = *v.Port
		}
		p, err = probe.NewSshProbe(probe.SshProbeOptions{
			Host:    v.Host,
			Port:    port,
			HostKey: v.HostKey,
		})
	case DomainProbeConfig:
		timeout := 60
		if v.Timeout != nil {
			timeout = *v.Timeout
		}

		threshold := time.Duration(1)
		if v.Threshold != nil {
			threshold, err = time.ParseDuration(*v.Threshold)
			if err != nil {
				return nil, err
			}
		}
		p, err = probe.NewDomainProbe(probe.DomainProbeOptions{
			Domain:    v.Domain,
			Timeout:   timeout,
			Threshold: threshold,
		})
	case DnsProbeConfig:
		timeout := 60
		if v.Timeout != nil {
			timeout = *v.Timeout
			if timeout < 0 {
				return nil, errors.New("invalid timeout")
			}
		}

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

		p, err = probe.NewDnsProbe(probe.DnsProbeOptions{
			Server:     v.Server,
			Port:       uint16(port),
			Timeout:    time.Duration(timeout),
			Target:     v.Target,
			RecordType: recordType,
		})
	default:
		p = nil
		err = fmt.Errorf("unknown probe type: %s", pc.Type)
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
		messageTemplate := "{{.Name}} is now {{.Status}}, was {{.PreviousStatus}} for {{.Duration}}"
		if v.MessageTemplate != nil {
			messageTemplate = *v.MessageTemplate
		}
		reuseMessage := false
		var messageId *string = nil
		if v.ReuseMessage != nil {
			reuseMessage = v.ReuseMessage.Enable
			messageId = v.ReuseMessage.MessageId
		}

		n, err = notifier.NewDiscordWebhookNotifier(notifier.DiscordWebhookOptions{
			Url:               v.Url,
			MessageTemplate:   messageTemplate,
			PersistentMessage: reuseMessage,
			LastMessageId:     messageId,
		})
	case GotifyNotifierConfig:
		n, err = notifier.NewGotifyNotifier(notifier.GotifyOptions{Url: v.Url, Token: v.Token})
	default:
		n = nil
		err = fmt.Errorf("unknown probe type: %s", nc.Type)
	}
	return n, err
}

func Count[T any](ts []T, pred func(T) bool) int {
	count := 0
	for _, t := range ts {
		if pred(t) {
			count += 1
		}
	}
	return count
}

func main() {
	cfg := "config.yaml"
	cfgEnv := os.Getenv("NEKO_CONFIG")
	if len(cfgEnv) != 0 {
		cfg = cfgEnv
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

	instance := ""
	if hostname, err := os.Hostname(); err == nil {
		instance = hostname
	}
	if config.Instance != nil {
		instance = *config.Instance
	}
	instanceEnv := os.Getenv("NEKO_INSTANCE")
	if len(instanceEnv) != 0 {
		instance = instanceEnv
	}

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
			Name:          m.Name,
			Interval:      m.Interval,
			Probe:         p,
			Notifiers:     maps.Values(notifiers),
			Status:        core.StatusPending,
			Configuration: m,
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

			interval, err := time.ParseDuration(monitor.Interval)
			if err != nil {
				log.Fatal(err)
			}

			lastTransition := time.Now()

			ticker := time.NewTicker(interval)
			log.Printf("starting monitor %s", monitor.Name)
			for {
				<-ticker.C
				log.Printf("running monitor %v", monitor.Name)
				res, err := monitor.Probe.Probe()
				if err != nil {
					log.Printf("monitor %v failed: %s", monitor.Name, err)
					continue
				}
				log.Printf("Probe %v completed with result: %v", monitor.Name, res.Tests)

				if len(res.Tests) == 0 {
					continue
				}

				// calculate new state
				previousStatus := monitor.Status
				var status core.Status
				testCount := len(res.Tests)
				if testCount == 1 {
					status = res.Tests[0].Status
				} else {
					status = core.StatusDown
					count := Count(res.Tests, func(test core.Test) bool { return test.Status == core.StatusUp })
					if monitor.Configuration.ConsiderAllTests && count == testCount {
						status = core.StatusUp
					} else if !monitor.Configuration.ConsiderAllTests && count > 0 {
						status = core.StatusUp
					}
				}

				monitor.Status = status

				if previousStatus != status {
					now := time.Now()
					if previousStatus != core.StatusPending {
						data := make(map[string]interface{})
						data["Instance"] = instance
						data["Name"] = monitor.Name
						data["PreviousStatus"] = fmt.Sprintf("%v", previousStatus)
						data["Status"] = fmt.Sprintf("%v", status)
						data["TimeNotify"] = now
						data["TimeNotifyUnix"] = now.Unix()
						data["Duration"] = now.Sub(lastTransition).Round(time.Second)

						for _, n := range monitor.Notifiers {
							if err := n.Notify(monitor.Name, data); err != nil {
								log.Printf("unable to notify: %s", err)
							}
						}
					}
					lastTransition = now
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
