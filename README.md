# Neko

An(other) uptime monitor.

## Installation

### Binary

### Helm

This repository includes a Helm chart for deployment on Kubernetes

```sh
helm install -n neko neko
```

## Configuration

The configuration is loaded from a YAML file named `config.yaml` in the working directory.

> ❕ This path can be controlled by setting the `NEKO_CONFIG` environment variable.

See [config.example.yaml](config.example.yaml) for an example.

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

| Key         | Required | Description                              |
| ----------- | -------- | ---------------------------------------- |
| `interval`  | Yes      | Amount of time in between probe attempts |
| `probe`     | Yes      | A [probe](#probes) configuration         |
| `notifiers` | No       | A list of [notifiers](#notifiers)        |

### Notifiers

#### Discord Webhook

```yaml
type: discord_webhook
config:
  url: https://discord.com/api/webhooks/webhook_id/webhook_token
```

| Key   | Required | Description                |
| ----- | -------- | -------------------------- |
| `url` | Yes      | URL of the Discord webhook |

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

#### Ping

```yaml
type: ping
config:
  address: some-machine.example.com
```

| Key       | Required | Description |
| --------- | -------- | ----------- |
| `address` | Yes      |             |

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

| Key    | Required | Description |
| ------ | -------- | ----------- |
| `domain` | Yes      |             |

##### Extras

| Key         | Value                       |
| ----------- | --------------------------- |
| `remaining` | Amount of time until expiry |
