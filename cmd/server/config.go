package main

import (
	"fmt"
	"time"

	"github.com/guarandoo/neko/pkg/probe"
)

type SmtpNotifierCOnfig struct {
	Host       string   `yaml:"host"`
	Port       int      `yaml:"port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	Sender     string   `yaml:"sender"`
	Recipients []string `yaml:"recipients"`
}

type DiscordWebhookNotifierReuseMessageConfig struct {
	Enable    bool    `yaml:"enable"`
	MessageId *string `yaml:"messageId"`
}

type DiscordWebhookNotifierConfig struct {
	Url             string                                    `yaml:"url"`
	MessageTemplate *string                                   `yaml:"messageTemplate"`
	ReuseMessage    *DiscordWebhookNotifierReuseMessageConfig `yaml:"reuseMessage"`
}

type GotifyNotifierConfig struct {
	Url             string  `yaml:"url"`
	Token           string  `yaml:"token"`
	TitleTemplate   *string `yaml:"titleTemplate"`
	MessageTemplate *string `yaml:"messsageTemplate"`
}

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
			Config SmtpNotifierCOnfig `yaml:"config"`
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

type ProbeTypeConfig struct {
}

type ExecProbeTypeConfig struct {
	ProbeTypeConfig
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

type PingProbeTypeConfig struct {
	ProbeTypeConfig
	Address             string   `yaml:"address"`
	Count               *int     `yaml:"count"`
	PacketLossThreshold *float64 `yaml:"packetLossThreshold"`
}

type HttpProbeTypeConfig struct {
	ProbeTypeConfig
	Address            string             `yaml:"address"`
	MaxRedirects       *int               `yaml:"maxRedirects"`
	SuccessStatusCodes *[]int             `yaml:"successStatusCodes"`
	Headers            *map[string]string `yaml:"headers"`
}

type SshProbeTypeConfig struct {
	ProbeTypeConfig
	Host    string  `yaml:"host"`
	Port    *int    `yaml:"port"`
	HostKey *string `yaml:"hostKey"`
}

type DomainProbeTypeConfig struct {
	ProbeTypeConfig
	Domain    string  `yaml:"domain"`
	Threshold *string `yaml:"threshold"`
}

type DnsProbeTypeConfig struct {
	ProbeTypeConfig
	Server     string            `yaml:"server"`
	Port       *int              `yaml:"port"`
	Target     string            `yaml:"target"`
	RecordType *probe.RecordType `yaml:"recordType"`
}

type ProbeConfig struct {
	Type    string
	Timeout *time.Duration
	Config  any
}

func (f *ProbeConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var t struct {
		Type    string  `yaml:"type"`
		Timeout *string `yaml:"timeout"`
	}
	err := unmarshal(&t)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config: %w", err)
	}
	f.Type = t.Type
	if t.Timeout == nil {
		duration, err := time.ParseDuration(*t.Timeout)
		if err == nil {
			f.Timeout = &duration
		}
	}

	switch t.Type {
	case "exec":
		var c struct {
			Config ExecProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout == nil {
			duration := time.Second * 30
			f.Timeout = &duration
		}

	case "ping":
		var c struct {
			Config PingProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout != nil {
			duration := time.Second * 4
			f.Timeout = &duration
		}

	case "http":
		var c struct {
			Config HttpProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout == nil {
			duration := time.Second * 60
			f.Timeout = &duration
		}

	case "ssh":
		var c struct {
			Config SshProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout == nil {
			duration := time.Second * 30
			f.Timeout = &duration
		}

	case "domain":
		var c struct {
			Config DomainProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout == nil {
			duration := time.Second * 30
			f.Timeout = &duration
		}

	case "dns":
		var c struct {
			Config DnsProbeTypeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
		if f.Timeout == nil {
			duration := time.Second * 5
			f.Timeout = &duration
		}

	default:
		return fmt.Errorf("unknown probe type: %s", f.Type)
	}
	return nil
}

type ProbeNotifierConfig struct {
	Name string
}

type MonitorConfig struct {
	Name             string                `yaml:"name"`
	Interval         string                `yaml:"interval"`
	Probe            ProbeConfig           `yaml:"probe"`
	Notifiers        []ProbeNotifierConfig `yaml:"notifiers"`
	ConsiderAllTests bool                  `yaml:"considerAllTests"`
	Invert           bool                  `yaml:"invert"`
}

type MetricsConfiguration struct {
	ListenAddress string `yaml:"listenAddress"`
}

type Configuration struct {
	Instance         *string                   `yaml:"instance"`
	Notifiers        map[string]NotifierConfig `yaml:"notifiers"`
	IncludeNotifiers *string                   `yaml:"includeNotifiers"`
	Monitors         []MonitorConfig           `yaml:"monitors"`
	IncludeMonitors  *string                   `yaml:"includeMonitors"`
	Metrics          MetricsConfiguration      `yaml:"metrics"`
}
