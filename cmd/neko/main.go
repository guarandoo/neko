package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/exp/maps"

	"github.com/guarandoo/neko/pkg/core"
	"github.com/guarandoo/neko/pkg/notifier"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"github.com/mxmauro/resetevent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type app struct {
	metricsProbeAttempts       *prometheus.CounterVec
	metricsProbeAttemptsFailed *prometheus.CounterVec
	metricsUp                  *prometheus.GaugeVec
	metricsScrapeDuration      *prometheus.HistogramVec
}

func (p *app) createRaft() error {
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

func (p *app) runMonitor(extraLabels []string, monitor *Monitor, context context.Context, lastTransition *time.Time, instance string) error {
	labels := lo.Union([]string{instance, monitor.Name, monitor.Configuration.Probe.Type}, extraLabels)
	p.metricsProbeAttempts.WithLabelValues(labels...).Add(1.0)
	start := time.Now()

	res, err := monitor.Probe.Probe(context, instance, monitor.Name)
	duration := time.Since(start)
	if err != nil {
		p.metricsProbeAttemptsFailed.WithLabelValues(labels...).Add(1.0)
		return fmt.Errorf("monitor %v failed: %s", monitor.Name, err)
	}
	log.Printf("probe %v completed with result: %v", monitor.Name, res.Tests)

	p.metricsScrapeDuration.WithLabelValues(labels...).Observe(float64(duration.Nanoseconds()))

	if len(res.Tests) == 0 {
		return nil
	}

	// calculate new state
	previousStatus := monitor.Status
	var status core.Status
	testCount := len(res.Tests)
	if testCount == 1 {
		status = res.Tests[0].Status
	} else {
		status = core.StatusDown
		count := lo.CountBy(res.Tests, func(test core.Test) bool { return test.Status == core.StatusUp })
		if monitor.Configuration.ConsiderAllTests && count == testCount {
			status = core.StatusUp
		} else if !monitor.Configuration.ConsiderAllTests && count > 0 {
			status = core.StatusUp
		}
	}

	if monitor.Configuration.Invert {
		switch status {
		case core.StatusUp:
			status = core.StatusDown
		case core.StatusDown:
			status = core.StatusUp
		}
	}

	monitor.Status = status
	gauge := p.metricsUp.WithLabelValues(labels...)
	if status == core.StatusUp {
		gauge.Set(1)
	} else {
		gauge.Set(0)
	}

	if previousStatus != status {
		now := time.Now()
		if previousStatus != core.StatusPending {
			data := make(map[string]any)
			data["Instance"] = instance
			data["Name"] = monitor.Name
			data["PreviousStatus"] = fmt.Sprintf("%v", previousStatus)
			data["Status"] = fmt.Sprintf("%v", status)
			data["TimeNotify"] = now
			data["TimeNotifyUnix"] = now.Unix()
			data["Duration"] = now.Sub(*lastTransition).Round(time.Second)

			for _, n := range monitor.Notifiers {
				if err := n.Notify(monitor.Name, data); err != nil {
					log.Printf("unable to notify: %s", err)
				}
			}
		}
		*lastTransition = now
	}

	return nil
}

func loadConfiguration(path string) (*Configuration, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get filename: %s", err)
	}

	log.Printf("loading configuration file from: %s", filename)

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config file does not exist: %s", err)
	}

	text, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %v", err)
	}

	var config Configuration
	err = yaml.Unmarshal(text, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config text: %v", err)
	}

	return &config, nil
}

func newApp() *app {
	instance := app{}

	return &instance
}

func (p *app) run() error {
	start := resetevent.NewManualResetEvent()

	cfgPaths := []string{}

	cfgEnv := os.Getenv("NEKO_CONFIG")
	if len(cfgEnv) != 0 {
		cfgPaths = append(cfgPaths, cfgEnv)
	}
	cfgPaths = append(cfgPaths, "config.yaml")
	cfgPaths = append(cfgPaths, "/etc/neko/config.yaml")

	var config *Configuration
	var err error
	for _, cfgPath := range cfgPaths {
		config, err = loadConfiguration(cfgPath)
		if err != nil {
			log.Printf("unable to load configuration from %v: %v", cfgPath, err)
		} else {
			log.Printf("successfully loaded configuration from: %v", cfgPath)
			break
		}
	}

	if config == nil {
		return errors.New("unable to load configuration")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	if len(config.IncludeNotifiers) > 0 {
		f, err := filepath.Abs(config.IncludeNotifiers)
		if err != nil {
			return fmt.Errorf("unable to get filename: %w", err)
		}
		if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config file does not exist: %w", err)
		}
		t, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("unable to read config file: %w", err)
		}
		var c map[string]NotifierConfig
		err = yaml.Unmarshal(t, &c)
		if err != nil {
			return fmt.Errorf("unable to unmarshal config text: %w", err)
		}
		for k, v := range c {
			config.Notifiers[k] = v
		}
	}

	if len(config.IncludeMonitors) > 0 {
		f, err := filepath.Abs(config.IncludeMonitors)
		if err != nil {
			return fmt.Errorf("unable to get filename: %s", err)
		}
		if _, err := os.Stat(f); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config file does not exist: %s", err)
		}
		t, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("unable to read config file: %s", err)
		}
		var c []MonitorConfig
		err = yaml.Unmarshal(t, &c)
		if err != nil {
			return fmt.Errorf("unable to unmarshal config text: %v", err)
		}
		config.Monitors = append(config.Monitors, c...)
	}

	var wg sync.WaitGroup

	if config.Metrics.Enable {
		keys := lo.Keys(config.Metrics.ExtraLabels)
		labels := lo.Union([]string{"instance", "monitor", "type"}, keys)
		p.metricsProbeAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "neko_probe_attempts_total",
			Help: "",
		}, labels)
		p.metricsProbeAttemptsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "neko_probe_attempts_failed",
			Help: "",
		}, labels)
		p.metricsUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "neko_up",
			Help: "",
		}, labels)
		p.metricsScrapeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "neko_scrape_duration_nanoseconds",
			Help: "",
		}, labels)

		log.Print("setting up metrics")

		metricsServerMux := http.NewServeMux()
		metricsServerMux.Handle("/metrics", promhttp.Handler())
		metricsServer := http.Server{
			Addr:    config.Metrics.ListenAddress,
			Handler: metricsServerMux,
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := metricsServer.ListenAndServe(); err != nil {
				log.Fatalf("unable to start metrics server: %v", err)
			}
		}()
	}

	clusterConfig := config.Cluster
	if clusterConfig.Enable {
		log.Print("setting up cluster")

		memberlistCfg := memberlist.DefaultWANConfig()
		memberlistCfg.Name = config.Instance

		serfCh := make(chan serf.Event, 16)

		serfCfg := serf.DefaultConfig()
		serfCfg.NodeName = config.Instance
		serfCfg.EventCh = serfCh
		serfCfg.MemberlistConfig = memberlistCfg
		serfCfg.LogOutput = os.Stdout

		s, err := serf.Create(serfCfg)
		if err != nil {
			log.Fatalf("failed to create serf: %v", err)
		}

		if len(clusterConfig.Join) > 0 {
			_, err = s.Join(clusterConfig.Join, false)
			if err != nil {
				log.Fatalf("unable to join cluster: %v", err)
			}
		}
	}

	notifiers := map[string]notifier.Notifier{}
	for k, v := range config.Notifiers {
		log.Printf("setting up notifier %s", k)

		n, err := createNotifier(&v)
		if err != nil {
			log.Fatalf("unable to create notifier: %s", err)
		}

		notifiers[k] = n
	}

	monitors := []Monitor{}
	for _, m := range config.Monitors {
		log.Printf("setting up monitor %s", m.Name)
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

	rootContext, _ := context.WithCancel(context.Background())

	// pool := pond.NewPool(config.ConcurrentTasks)

	for _, m := range monitors {
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

		outer:
			for {
				select {
				case <-ticker.C:
				case <-rootContext.Done():
					break outer
				}
				log.Printf("running monitor %v", monitor.Name)

				func() {
					context, cancel := context.WithTimeout(rootContext, monitor.Configuration.Probe.Timeout)
					defer cancel()

					extraLabels := lo.Values(config.Metrics.ExtraLabels)
					if err := p.runMonitor(extraLabels, &monitor, context, &lastTransition, config.Instance); err != nil {
						return
					}
				}()
			}
			log.Printf("stopping monitor %s", monitor.Name)
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

	ticker := time.NewTicker(3 * time.Second)
	ch := make(chan int)
	for {
		select {
		case <-ch:
		case <-ticker.C:
		}
	}
}

func main() {
	app := newApp()
	err := app.run()
	if err != nil {
		log.Fatalf("unable to run application: %v", err)
	}
}
