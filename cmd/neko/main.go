package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/exp/maps"

	"github.com/guarandoo/neko/pkg/core"
	"github.com/guarandoo/neko/pkg/notifier"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
	"github.com/mxmauro/resetevent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type Application interface {
	Run(context.Context) error
	Reload() error
}

type application struct {
	metricsProbeAttempts       *prometheus.CounterVec
	metricsProbeAttemptsFailed *prometheus.CounterVec
	metricsUp                  *prometheus.GaugeVec
	metricsScrapeDuration      *prometheus.HistogramVec
	metricsServer              MetricsServer
	configurationFile          string
	configuration              *Configuration
}

func (p *application) Reload() error {
	oldConfig := p.configuration

	newConfig, err := loadConfiguration(p.configurationFile)
	if err != nil {
		return err
	}

	p.configuration = newConfig

	{
		listenAddressChanged := newConfig.Metrics.ListenAddress != oldConfig.Metrics.ListenAddress

		shouldShutdown := !newConfig.Metrics.Enable || listenAddressChanged
		if shouldShutdown {
			if err := p.metricsServer.Close(); err != nil {
				log.Printf("unable to shut down metrics server: %v", err)
			}
		}

		shouldStart := (!oldConfig.Metrics.Enable && newConfig.Metrics.Enable) || (newConfig.Metrics.Enable && listenAddressChanged)
		if shouldStart {
			if err := p.metricsServer.Listen(newConfig.Metrics.ListenAddress); err != nil {
				log.Printf("unable to shart metrics server: %v", err)
			}
		}
	}

	return nil
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-time.After(d):
	case <-ctx.Done():
	}
}

func (p *application) runMonitor(ctx context.Context, extraLabels []string, monitor *Monitor, lastTransition *time.Time, instance string) error {
	labels := lo.Union([]string{instance, monitor.Name, monitor.Configuration.Probe.Type}, extraLabels)

	previousStatus := monitor.Status
	var status core.Status

	attempts := 0
	for {
		p.metricsProbeAttempts.WithLabelValues(labels...).Add(1.0)
		start := time.Now()

		probeCtx, cancelProbeCtx := context.WithTimeout(ctx, monitor.Configuration.Probe.Timeout)
		defer cancelProbeCtx()

		res, err := monitor.Probe.Probe(probeCtx, instance, monitor.Name)
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

		attempts += 1

		if status != core.StatusDown {
			break
		}

		if attempts > monitor.Configuration.Retry.MaxAttempts {
			break
		}

		log.Printf("monitor %v failed, retrying after %v %v/%v", monitor.Name, monitor.Configuration.Retry.Interval, attempts, monitor.Configuration.Retry.MaxAttempts)
		sleep(ctx, monitor.Configuration.Retry.Interval)
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
				if err := n.Notify(ctx, monitor.Name, data); err != nil {
					log.Printf("unable to notify: %s", err)
				}
			}
		}
		*lastTransition = now
	}

	return nil
}

func (p *application) Run(ctx context.Context) error {
	start := resetevent.NewManualResetEvent()

	cfgPaths := []string{}

	cfgEnv := os.Getenv("NEKO_CONFIG")
	if len(cfgEnv) != 0 {
		cfgPaths = append(cfgPaths, cfgEnv)
	}
	cfgPaths = append(cfgPaths, "config.yaml")
	cfgPaths = append(cfgPaths, "/etc/neko/config.yaml")

	for _, cfgPath := range cfgPaths {
		config, err := loadConfiguration(cfgPath)
		if err != nil {
			log.Printf("unable to load configuration from %v: %v", cfgPath, err)
		} else {
			log.Printf("successfully loaded configuration from: %v", cfgPath)
			p.configurationFile = cfgPath
			p.configuration = config
			break
		}
	}

	if p.configuration == nil {
		return errors.New("unable to load configuration")
	}

	if err := p.configuration.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	log.Printf("running as instance: %v", p.configuration.Instance)

	if len(p.configuration.IncludeNotifiers) > 0 {
		f, err := filepath.Abs(p.configuration.IncludeNotifiers)
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
			p.configuration.Notifiers[k] = v
		}
	}

	if len(p.configuration.IncludeMonitors) > 0 {
		f, err := filepath.Abs(p.configuration.IncludeMonitors)
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
		p.configuration.Monitors = append(p.configuration.Monitors, c...)
	}

	keys := lo.Keys(p.configuration.Metrics.ExtraLabels)
	labels := lo.Union([]string{"instance", "monitor", "type"}, keys)
	p.metricsProbeAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "neko_probe_attempts_total",
		Help: "Total number of probe attempts.",
	}, labels)
	p.metricsProbeAttemptsFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "neko_probe_attempts_failed",
		Help: "Number of probe attempts that failed.",
	}, labels)
	p.metricsUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_up",
		Help: "",
	}, labels)
	p.metricsScrapeDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "neko_scrape_duration_nanoseconds",
		Help: "The amount of time the probe took.",
	}, labels)

	p.metricsServer = newMetricsServer(promhttp.Handler())
	if p.configuration.Metrics.Enable {
		log.Print("setting up metrics")

		if err := p.metricsServer.Listen(p.configuration.Metrics.ListenAddress); err != nil {
			return fmt.Errorf("unable to start metrics server: %v", err)
		}
	}

	clusterConfig := p.configuration.Cluster
	if clusterConfig.Enable {
		log.Print("setting up cluster")

		memberlistCfg := memberlist.DefaultWANConfig()
		memberlistCfg.Name = p.configuration.Instance
		memberlistCfg.BindAddr = p.configuration.Cluster.Memberlist.BindAddress
		memberlistCfg.BindPort = p.configuration.Cluster.Memberlist.BindPort
		memberlistCfg.AdvertisePort = p.configuration.Cluster.Memberlist.AdvertisePort
		memberlistCfg.AdvertiseAddr = p.configuration.Cluster.Memberlist.AdvertiseAddress

		serfCh := make(chan serf.Event, 16)

		serfCfg := serf.DefaultConfig()
		serfCfg.NodeName = p.configuration.Instance
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
	for k, v := range p.configuration.Notifiers {
		log.Printf("setting up notifier %s", k)

		n, err := createNotifier(&v)
		if err != nil {
			log.Fatalf("unable to create notifier: %s", err)
		}

		notifiers[k] = n
	}

	monitors := []Monitor{}
	for _, m := range p.configuration.Monitors {
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
				case <-ctx.Done():
					break outer
				}
				log.Printf("running monitor %v", monitor.Name)

				func() {
					extraLabels := lo.Values(p.configuration.Metrics.ExtraLabels)
					if err := p.runMonitor(ctx, extraLabels, &monitor, &lastTransition, p.configuration.Instance); err != nil {
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

	return nil
}

func newApp() Application {
	instance := application{}

	return &instance
}

func sighandler(cancel context.CancelFunc, app Application) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

outer:
	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGQUIT:
			log.Printf("received signal: %v", sig)
			cancel()
			break outer

		case syscall.SIGHUP:
			log.Printf("received signal: %v", sig)
			err := app.Reload()
			if err != nil {
				log.Printf("unable to reload configuration: %v", err)
			}
		}
	}
}

func main() {
	log.Printf("starting neko %v %v %v", Version, Commit, BuildTime)

	ctx, cancel := context.WithCancel(context.Background())

	app := newApp()
	err := app.Run(ctx)
	if err != nil {
		log.Fatalf("unable to run application: %v", err)
	}

	go sighandler(cancel, app)
	<-ctx.Done()

	log.Printf("terminating")
}
