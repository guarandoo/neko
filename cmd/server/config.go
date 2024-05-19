package main

import (
	"fmt"
)

type SmtpNotifierCOnfig struct {
	Host       string   `yaml:"host"`
	Port       int      `yaml:"port"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	Sender     string   `yaml:"sender"`
	Recipients []string `yaml:"recipients"`
}

type DiscordWebhookNotifierConfig struct {
	Url string `yaml:"url"`
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
	Address string `yaml:"address"`
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
	default:
		return fmt.Errorf("unknown probe type: %s", f.Type)
	}
	return nil
}

type ProbeNotifierConfig struct {
	Name string
}

type MonitorConfig struct {
	Name      string                `yaml:"name"`
	Interval  int                   `yaml:"interval"`
	Probe     ProbeConfig           `yaml:"probe"`
	Notifiers []ProbeNotifierConfig `yaml:"notifiers"`
}

type Configuration struct {
	Notifiers map[string]NotifierConfig `yaml:"notifiers"`
	Monitors  []MonitorConfig           `yaml:"monitors"`
}
