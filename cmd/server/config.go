package main

import (
	"fmt"

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
	Url   string `yaml:"url"`
	Token string `yaml:"token"`
}

type NotifierConfig struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config"`
}

func (f *NotifierConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

type ExecProbeConfig struct {
	Path string   `yaml:"path"`
	Args []string `yaml:"args"`
}

type PingProbeConfig struct {
	Address string `yaml:"address"`
}

type HttpProbeConfig struct {
	Address      string `yaml:"address"`
	MaxRedirects *int   `yaml:"maxRedirects"`
	Timeout      *int   `yaml:"timeout"`
}

type SshProbeConfig struct {
	Host    string  `yaml:"host"`
	Port    *int    `yaml:"port"`
	HostKey *string `yaml:"hostKey"`
}

type DomainProbeConfig struct {
	Domain    string  `yaml:"domain"`
	Timeout   *int    `yaml:"timeout"`
	Threshold *string `yaml:"threshold"`
}

type DnsProbeConfig struct {
	Server     string            `yaml:"server"`
	Timeout    *int              `yaml:"timeout"`
	Port       *int              `yaml:"port"`
	Target     string            `yaml:"target"`
	RecordType *probe.RecordType `yaml:"recordType"`
}

type ProbeConfig struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config"`
}

func (f *ProbeConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t struct {
		Type string `yaml:"type"`
	}
	err := unmarshal(&t)
	if err != nil {
		return err
	}
	f.Type = t.Type
	switch t.Type {
	case "exec":
		var c struct {
			Config ExecProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "ping":
		var c struct {
			Config PingProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "http":
		var c struct {
			Config HttpProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "ssh":
		var c struct {
			Config SshProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "domain":
		var c struct {
			Config DomainProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
	case "dns":
		var c struct {
			Config DnsProbeConfig `yaml:"config"`
		}
		err := unmarshal(&c)
		if err != nil {
			return err
		}
		f.Config = c.Config
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
}

type Configuration struct {
	Instance  *string                   `yaml:"instance"`
	Notifiers map[string]NotifierConfig `yaml:"notifiers"`
	Monitors  []MonitorConfig           `yaml:"monitors"`
}
