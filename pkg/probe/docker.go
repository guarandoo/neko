package probe

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/guarandoo/neko/pkg/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const DockerProbeType string = "docker"

var (
	onceInitDockerProbe             sync.Once
	metricDockerContainers          *prometheus.GaugeVec
	metricDockerSwarmServiceRunning *prometheus.GaugeVec
	metricDockerSwarmServiceDesired *prometheus.GaugeVec
	metricDockerSwarmNodes          *prometheus.GaugeVec
)

func initDockerProbe() {
	metricDockerContainers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_docker_containers",
	}, []string{"instance", "monitor", "type"})
	metricDockerSwarmServiceRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_docker_swarm_service_running",
	}, []string{"instance", "monitor", "type", "service", "mode"})
	metricDockerSwarmServiceDesired = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_docker_swarm_service_desired",
	}, []string{"instance", "monitor", "type", "service", "mode"})
	metricDockerSwarmNodes = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "neko_docker_swarm_nodes",
	}, []string{"instance", "monitor", "type", "state", "availability", "role"})
}

type dockerProbe struct {
	host string
}

func (p *dockerProbe) Probe(ctx context.Context, instance string, monitor string) (res *core.Result, err error) {
	test := core.Test{
		Target: p.host,
		Status: core.StatusUp,
		Error:  nil,
		Extras: nil,
	}
	result := core.Result{
		Tests: []core.Test{test},
	}

	extras := make(map[string]any)

	client, err := client.NewClientWithOpts(client.WithHost(p.host), client.WithAPIVersionNegotiation())
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &result, nil
	}
	defer func() { client.Close() }()

	_, err = client.Ping(ctx)
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &result, nil
	}

	info, err := client.Info(ctx)
	if err != nil {
		test.Status = core.StatusDown
		test.Error = err
		return &result, nil
	}

	if info.Swarm.LocalNodeState == "active" {
		if info.Swarm.ControlAvailable {
			nodes, err := client.NodeList(ctx, swarm.NodeListOptions{})
			if err != nil {
				test.Status = core.StatusDown
				test.Error = err
				return &core.Result{Tests: []core.Test{test}}, nil
			}

			test.Extras["nodes"] = nodes

			type nodeKey struct {
				state        string
				availability string
				role         string
			}
			nodeCounts := make(map[nodeKey]int)
			for _, node := range nodes {
				key := nodeKey{
					state:        string(node.Status.State),
					availability: string(node.Spec.Availability),
					role:         string(node.Spec.Role),
				}
				nodeCounts[key]++
			}
			for key, count := range nodeCounts {
				metricDockerSwarmNodes.WithLabelValues(instance, monitor, DockerProbeType, key.state, key.availability, key.role).Set(float64(count))
			}

			services, err := client.ServiceList(ctx, swarm.ServiceListOptions{})
			if err != nil {
				test.Status = core.StatusDown
				test.Error = err
				return &core.Result{Tests: []core.Test{test}}, nil
			}

			test.Extras["services"] = services

			for _, service := range services {
				isGlobal := service.Spec.Mode.Global != nil
				mode := ""
				if isGlobal {
					mode = "global"
				} else {
					mode = "replicated"
				}

				filter := filters.NewArgs()
				filter.Add("service", service.ID)
				filter.Add("desired-state", "running")

				tasks, err := client.TaskList(ctx, swarm.TaskListOptions{Filters: filter})
				if err != nil {
					test.Status = core.StatusDown
					test.Error = err
					return &core.Result{Tests: []core.Test{test}}, nil
				}

				running := uint64(len(tasks))
				metricDockerSwarmServiceRunning.WithLabelValues(instance, monitor, DockerProbeType, service.Spec.Name, mode).Set(float64(running))

				desired := uint64(0)
				if !isGlobal {
					desired = *service.Spec.Mode.Replicated.Replicas
				}
				metricDockerSwarmServiceDesired.WithLabelValues(instance, monitor, DockerProbeType, service.Spec.Name, mode).Set(float64(desired))
			}
		}
	}

	extras["info"] = info
	test.Extras = extras

	metricDockerContainers.WithLabelValues(instance, monitor, DockerProbeType).Set(float64(info.Containers))

	return &core.Result{Tests: []core.Test{test}}, nil
}

type DockerProbeOptions struct {
	ProbeOptions
	Host string
}

func getDockerHost() string {
	if runtime.GOOS != "linux" {
		return client.DefaultDockerHost
	}

	xdgRuntime := os.Getenv("XDG_RUNTIME_DIR")
	if xdgRuntime != "" {
		socket := fmt.Sprintf("unix://%s/docker.sock", xdgRuntime)
		if _, err := os.Stat(strings.TrimPrefix(socket, "unix://")); err == nil {
			return socket
		}
	}

	uid := os.Getuid()
	userSocket := fmt.Sprintf("unix:///run/user/%d/docker.sock", uid)
	if _, err := os.Stat(strings.TrimPrefix(userSocket, "unix://")); err == nil {
		return userSocket
	}

	return client.DefaultDockerHost
}

func NewDockerProbe(options DockerProbeOptions) (Probe, error) {
	onceInitDockerProbe.Do(initDockerProbe)

	host := options.Host
	if host == "" {
		host = getDockerHost()
	}
	return &dockerProbe{
		host: host,
	}, nil
}
