![Release status badege](https://git.calliope.rip/guarandoo/neko/actions/workflows/release.yaml/badge.svg)

# Neko

An(other) uptime monitor.

## Table of Contents

- [Installation](#installation)
  - [Building](#building)
  - [Binary](#binary)
  - [Helm](#helm)
- [Configuraton](#configuration)
  - [Distributed Mode](#distributed-mode)
  - [Monitors](#monitors)
  - [Notifiers](#notifiers)
    - [Discord Webhook](#discord-webhook)
    - [Gotify](#gotify)
    - [SMTP](#smtp)
  - [Probes](#probes)
    - [Ping](#ping)
    - [HTTP](#http)
    - [SSH](#ssh)
    - [SQL](#sql)
    - [Domain](#domain)

## Installation

### Building

To build a binary for your current platform:

```sh
make
```

To build binaries for all supported platforms:

```sh
make all-binaries
```

To build a Docker image for your current platform:

```sh
make docker-image
```

To build a multi-arch Docker image:

```sh
make docker-multiarch-image
```

### Binary

Grab prebuilt binaries from [Releases](/guarandoo/neko/releases)

### Helm

This repository includes a Helm chart for deployment on Kubernetes

```sh
helm install -n neko neko
```

### NixOS

This repository provides a Nix flake which can be used with a NixOS system:

```nix
{
  description = "My NixOS flake";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/unstable";
    # ...
    neko.url = "git+https://git.calliope.rip/guarandoo/neko"; # add flake as input
  };
  outputs = {
    nixpkgs,
    neko,
    ...
  }: {
    nixosConfigurations = {
      my-nixos-system = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          # ...
          neko.nixosModules.default # add module
        ];
        config = {
          # ...
          services.neko = {
            enable = true;
            settings = {
              instance = "my-neko-instance";
              notifiers = {
                # ...
              };
              monitors = [
                # ...
              ];
              metrics = {
                listenAddress = "127.0.0.1:9090";
              };
            };
          };
        };
      }
    };
  }
}
```

## Configuration

The configuration is loaded from a YAML file named `config.yaml` in the working directory.

> ❕ This path can be controlled by setting the `NEKO_CONFIG` environment variable.

| Key         | Required | Description                                                         |
| ----------- | -------- | ------------------------------------------------------------------- |
| `instance`  | No       | A unique instance identifier, defaults to hostname if not specified |
| `notifiers` | Yes      | A list of [notifier](#notifiers) configurations                     |
| `monitors`  | Yes      | A list of [probe](#probes) configurations                           |

See [config.example.yaml](config.example.yaml) for a full example.

### Distributed Mode

You can run multiple instances across machines to provide redundancy or monitor your targets from multiple vantage points.

### Monitors

A monitor is composed of a [probe](#probes) and zero or more [notifiers](#notifiers).

```yaml
interval: 1m
probe:
  type: ssh
  config:
    host: some-machine.example.com
notifiers:
  - my_discord_notifier
  - my_smtp_notifier
```

| Key                | Type           | Required | Description                                                     |
| ------------------ | -------------- | -------- | --------------------------------------------------------------- |
| `interval`         | Duration       | Yes      | Amount of time in between probe attempts                        |
| `probe`            | ProbeConfig    | Yes      | A [probe](#probes) configuration                                |
| `notifiers`        | NotifierConfig | No       | A list of [notifiers](#notifiers)                               |
| `considerAllTests` | bool           | No       | Whether to require all tests to pass to consider the monitor Up |

### Notifiers

Most notifiers offer the ability to customize the message in the notification. For those that do, the following variables are available.

| Variable       | Description                                     | Data Type     |
| -------------- | ----------------------------------------------- | ------------- |
| Instance       | Neko instance name                              | string        |
| Name           | Name of the probe that triggered the notifier   | string        |
| PreviousStatus | Status of the probe before transition           | string        |
| Status         | Current status of the probe                     | string        |
| TimeNotify     | Time the notifier was triggered                 | time.Time     |
| TimeNotifyUnix | Time the notifier was triggered (in Unix epoch) | int64         |
| Duration       | Duration since the last state transition        | time.Duration |

#### Discord Webhook

```yaml
type: discord_webhook
config:
  url: https://discord.com/api/webhooks/webhook_id/webhook_token
  messageTemplate: |-
    {{.Name}} is {{.Status}}
```

| Key               | Required | Description                |
| ----------------- | -------- | -------------------------- |
| `url`             | Yes      | URL of the Discord webhook |
| `messageTemplate` | No       | Message template           |

#### Gotify

#### SMTP

```yaml
type: smtp
config:
  host: smtp.example.com
  port: 587
  username: smtp_user
  password: smtp_secret
  sender: neko@example.com
  recipients:
    - john.doe@example.com
```

| Key          | Required | Description                                            |
| ------------ | -------- | ------------------------------------------------------ |
| `host`       | Yes      | SMTP Host                                              |
| `port`       | Yes      | SMTP Port                                              |
| `username`   | No       | SMTP Username                                          |
| `password`   | No       | SMTP Password                                          |
| `sender`     | Yes      | Mail address that will be used to send outgoing emails |
| `recipients` | No       | List of mail addresses to notify                       |

### Probes

Certain probes may execute multiple tests; for example, the [HTTP](#http) probe attempts the request using all resolved A and AAAA records.

You can combine this with a monitor's `considerAllTests` setting to check if the target is reachable from all hosts instead of just one.

#### Ping

```yaml
type: ping
config:
  address: some-machine.example.com
  # count: 6
  # packetLossThreshold: 0.5
```

| Key                   | Required | Description                        |
| --------------------- | -------- | ---------------------------------- |
| `address`             | Yes      |                                    |
| `count`               | No       | Number of ICMP packets to send out |
| `packetLossThreshold` | No       | Packet loss threshold (in percent) |

#### HTTP

```yaml
type: http
config:
  address: https://example.com
  maxRedirects: 1
```

| Key            | Required | Description                         |
| -------------- | -------- | ----------------------------------- |
| `address`      | Yes      |                                     |
| `maxRedirects` | No       | Maximum number of allowed redirects |

#### SSH

```yaml
type: ssh
config:
  host: some-machine.example.com
```

| Key    | Required | Description |
| ------ | -------- | ----------- |
| `host` | Yes      |             |

#### SQL

#### Domain

```yaml
type: domain
config:
  domain: https://example.com
```

| Key      | Required | Description |
| -------- | -------- | ----------- |
| `domain` | Yes      |             |

##### Extras

| Key         | Value                       |
| ----------- | --------------------------- |
| `remaining` | Amount of time until expiry |

#### DNS

Test DNS resolution capability, the probe succeeds if there is at least 1 record of the specified type returned for the target.

> ! This probe is primarily designed to test nameserver functionality, if you need to monitor records for a domain take a look at [domain](#domain).

```yaml
type: dns
config:
  server: 1.1.1.1
  port: 53
  timeout: 60
  target: my.domain.com
  type: A
```

| Key          | Required | Description                                                      |
| ------------ | -------- | ---------------------------------------------------------------- |
| `host`       | Yes      | DNS server to query                                              |
| `port`       | No       |                                                                  |
| `timeout`    | No       |                                                                  |
| `target`     | Yes      |                                                                  |
| `recordType` | No       | Must be one of `Host` (A/AAAA), `NS` or `MX`, defaults to `Host` |

> ❕ It is recommended that `target` be a stably resolvable domain otherwise this probe may produce false-positives.

## Metrics

### Prometheus

A Prometheus metrics endpoint is available at the address configured via `metrics.listenAddress` and exports the following metrics:

| Metric                           | Description                               |
| -------------------------------- | ----------------------------------------- |
| neko_up                          | Monitor status (`1` or `0`)               |
| neko_scrape_duration_nanoseconds | Amount of time it took to execute a probe |
| neko_probe_attempts_total        | Total number of probe attempts            |
| neko_probe_attempts_failed       | Number of probe attempts that failed      |
