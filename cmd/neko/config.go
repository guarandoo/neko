package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/guarandoo/neko/pkg/probe"
	"gopkg.in/yaml.v3"
)

const (
	EnvPrefix                      = "NEKO_"
	DefaultNotifierTitleTemplate = "Monitor Status Change"
	DefaultNotifierMessageTemplate = "{{.Name}} is now {{.Status}}, was {{.PreviousStatus}} for {{.Duration}}"
)

func getEnv(key string) (string, bool) {
	return os.LookupEnv(fmt.Sprintf("%v%v", EnvPrefix, key))
}

// region notifiers

// region smtpnotifier
type SmtpNotifierConfig struct {
	Host            string   `yaml:"host"`
	Port            int      `yaml:"port"`
	Username        string   `yaml:"username"`
	Password        string   `yaml:"password"`
	Sender          string   `yaml:"sender"`
	Recipients      []string `yaml:"recipients"`
	SubjectTemplate string   `yaml:"subjectTemplate"`
	BodyTemplate    string   `yaml:"bodyTemplate"`
}

func (t *SmtpNotifierConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt SmtpNotifierConfig
	value := rt{
		Port:            587,
		SubjectTemplate: DefaultNotifierTitleTemplate,
		BodyTemplate:    DefaultNotifierMessageTemplate,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = SmtpNotifierConfig(value)
	return nil
}

// endregion

// region discordwebhooknotifier
type DiscordWebhookNotifierReuseMessageConfig struct {
	Enable    bool    `yaml:"enable"`
	MessageId *string `yaml:"messageId"`
}

type DiscordWebhookNotifierConfig struct {
	Url             string                                    `yaml:"url"`
	MessageTemplate string                                    `yaml:"messageTemplate"`
	ReuseMessage    *DiscordWebhookNotifierReuseMessageConfig `yaml:"reuseMessage"`
}

func (t *DiscordWebhookNotifierConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt DiscordWebhookNotifierConfig
	value := rt{
		MessageTemplate: DefaultNotifierMessageTemplate,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = DiscordWebhookNotifierConfig(value)
	return nil
}

// endregion

// region gotifynotifier
type GotifyNotifierConfig struct {
	Url             string `yaml:"url"`
	Token           string `yaml:"token"`
	TitleTemplate   string `yaml:"titleTemplate"`
	MessageTemplate string `yaml:"messsageTemplate"`
}

func (t *GotifyNotifierConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt GotifyNotifierConfig
	value := rt{
		TitleTemplate:   DefaultNotifierTitleTemplate,
		MessageTemplate: DefaultNotifierMessageTemplate,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = GotifyNotifierConfig(value)
	return nil
}

// endregion

// region notifier
type NotifierConfig struct {
	Type   string `yaml:"type"`
	Config any    `yaml:"config"`
}

func (f *NotifierConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var t struct {
		Type string `yaml:"type"`
	}
	err := unmarshal(&t)
	if err != nil {
		return err
	}
	f.Type = t.Type
	switch t.Type {
	case "smtp":
		var c struct {
			Config SmtpNotifierConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "discord_webhook":
		var c struct {
			Config DiscordWebhookNotifierConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "gotify":
		var c struct {
			Config GotifyNotifierConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	default:
		return fmt.Errorf("unknown notifier type: %s", f.Type)
	}
	return nil
}

// endregion

// endregion

// region probes
type ProbeTypeConfig struct {
}

// region execprobetype
type ExecProbeTypeConfig struct {
	ProbeTypeConfig
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

// endregion

// region pingprobetype
type PingProbeTypeConfig struct {
	ProbeTypeConfig
	Address             string        `yaml:"address"`
	Count               int           `yaml:"count"`
	PacketLossThreshold float64       `yaml:"packetLossThreshold"`
	Privileged          bool          `yaml:"privileged"`
	Interval            time.Duration `yaml:"interval"`
}

func (t *PingProbeTypeConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt PingProbeTypeConfig
	value := rt{
		Count:               1,
		PacketLossThreshold: 1.0,
		Privileged:          false,
		Interval:            time.Second * 1,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = PingProbeTypeConfig(value)
	return nil
}

// endregion

// region httpprobetype
type HttpProbeTypeConfig struct {
	ProbeTypeConfig
	Address            string            `yaml:"address"`
	SocketPath         string            `yaml:"socketPath"`
	MaxRedirects       int               `yaml:"maxRedirects"`
	SuccessStatusCodes []int             `yaml:"successStatusCodes"`
	Headers            map[string]string `yaml:"headers"`
}

func (t *HttpProbeTypeConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt HttpProbeTypeConfig
	value := rt{
		MaxRedirects:       20,
		SuccessStatusCodes: []int{200},
		Headers:            map[string]string{},
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = HttpProbeTypeConfig(value)
	return nil
}

// endregion

// region sshprobetype
type SshProbeTypeConfig struct {
	ProbeTypeConfig
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	HostKey string `yaml:"hostKey"`
}

func (t *SshProbeTypeConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt SshProbeTypeConfig
	value := rt{
		Port: 22,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = SshProbeTypeConfig(value)
	return nil
}

// endregion

// region domainprobetype
type DomainProbeTypeConfig struct {
	ProbeTypeConfig
	Domain    string        `yaml:"domain"`
	Threshold time.Duration `yaml:"threshold"`
}

func (t *DomainProbeTypeConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt DomainProbeTypeConfig
	value := rt{
		Threshold: 0,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = DomainProbeTypeConfig(value)
	return nil
}

// endregion

// region dnsprobetype
type DnsProbeTypeConfig struct {
	ProbeTypeConfig
	Server     string           `yaml:"server"`
	Port       int              `yaml:"port"`
	Target     string           `yaml:"target"`
	RecordType probe.RecordType `yaml:"recordType"`
}

func (t *DnsProbeTypeConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt DnsProbeTypeConfig
	value := rt{
		Port:       53,
		RecordType: probe.Host,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = DnsProbeTypeConfig(value)
	return nil
}

// endregion

// region probe
type ProbeConfig struct {
	Type    string
	Timeout time.Duration
	Config  any
}

func (t *ProbeConfig) UnmarshalYAML(n *yaml.Node) error {
	{
		var c struct {
			Type string `yaml:"type"`
		}
		err := n.Decode(&c)
		if err != nil {
			return fmt.Errorf("unable to unmarshal config: %w", err)
		}
		t.Type = c.Type
	}

	switch t.Type {
	case probe.ExecProbeType:
		type rt struct {
			Timeout time.Duration       `yaml:"timeout"`
			Config  ExecProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Minute * 1,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	case probe.PingProbeType:
		type rt struct {
			Timeout time.Duration       `yaml:"timeout"`
			Config  PingProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Second * 4,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	case probe.HttpProbeType:
		type rt struct {
			Timeout time.Duration       `yaml:"timeout"`
			Config  HttpProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Minute * 1,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	case probe.SshProbeType:
		type rt struct {
			Timeout time.Duration      `yaml:"timeout"`
			Config  SshProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Second * 30,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	case probe.DomainProbeType:
		type rt struct {
			Timeout time.Duration         `yaml:"timeout"`
			Config  DomainProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Second * 30,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	case probe.DnsProbeType:
		type rt struct {
			Timeout time.Duration      `yaml:"timeout"`
			Config  DnsProbeTypeConfig `yaml:"config"`
		}
		c := rt{
			Timeout: time.Second * 5,
		}
		err := n.Decode(&c)
		if err != nil {
			return err
		}
		t.Timeout = c.Timeout
		t.Config = c.Config

	default:
		return fmt.Errorf("unknown probe type: %s", t.Type)
	}

	return nil
}

// endregion

// region probenotifier
type ProbeNotifierConfig struct {
	Name string
}

// endregion

// region monitor
type MonitorConfig struct {
	Name             string                `yaml:"name"`
	Interval         string                `yaml:"interval"`
	Probe            ProbeConfig           `yaml:"probe"`
	Notifiers        []ProbeNotifierConfig `yaml:"notifiers"`
	ConsiderAllTests bool                  `yaml:"considerAllTests"`
	Invert           bool                  `yaml:"invert"`
}

func (t *MonitorConfig) UnmarshalYAML(n *yaml.Node) error {
	type rt MonitorConfig
	value := rt{
		Interval: "1m",
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = MonitorConfig(value)
	return nil
}

// endregion

// region metrics
type MetricsConfiguration struct {
	Enable        bool              `yaml:"enable"`
	ListenAddress string            `yaml:"listenAddress"`
	ExtraLabels   map[string]string `yaml:"extraLabels"`
}

func (t *MetricsConfiguration) UnmarshalYAML(n *yaml.Node) error {
	type rt MetricsConfiguration
	value := rt{
		ListenAddress: ":3000",
		ExtraLabels:   map[string]string{},
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = MetricsConfiguration(value)
	return nil
}

// endregion

// region memberlist
type MemberlistConfiguration struct {
	BindAddress      string `yaml:"bindAddress"`
	BindPort         int    `yaml:"bindPort"`
	AdvertiseAddress string `yaml:"advertiseAddress"`
	AdvertisePort    int    `yaml:"advertisePort"`
}

func (t *MemberlistConfiguration) UnmarshalYAML(n *yaml.Node) error {
	type rt MemberlistConfiguration
	value := rt{
		BindAddress:      "0.0.0.0",
		BindPort:         7946,
		AdvertiseAddress: "0.0.0.0",
		AdvertisePort:    7946,
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = MemberlistConfiguration(value)
	return nil
}

// endregion

// region cluster
type ClusterConfiugration struct {
	Enable     bool                    `yaml:"enable"`
	Memberlist MemberlistConfiguration `yaml:"memberlist"`
	Join       []string                `yaml:"join"`
}

// endregion

// region configuration
type Configuration struct {
	Instance         string                    `yaml:"instance"`
	Metrics          MetricsConfiguration      `yaml:"metrics"`
	Cluster          ClusterConfiugration      `yaml:"cluster"`
	ConcurrentTasks  int                       `yaml:"concurrentTasks"`
	Notifiers        map[string]NotifierConfig `yaml:"notifiers"`
	IncludeNotifiers string                    `yaml:"includeNotifiers"`
	Monitors         []MonitorConfig           `yaml:"monitors"`
	IncludeMonitors  string                    `yaml:"includeMonitors"`
}

func (t *Configuration) UnmarshalYAML(n *yaml.Node) error {
	type rt Configuration

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "neko"
	}

	value := rt{
		Instance:        hostname,
		ConcurrentTasks: runtime.NumCPU(),
	}
	if err := n.Decode(&value); err != nil {
		return err
	}
	*t = Configuration(value)

	if value, found := getEnv("INSTANCE"); found {
		t.Instance = value
	}

	return nil
}

func (c *Configuration) Validate() error {
	if c.ConcurrentTasks < 1 {
		return fmt.Errorf("invalid concurrentTasks value: %v", c.ConcurrentTasks)
	}
	return nil
}

// endregion

// endregion
