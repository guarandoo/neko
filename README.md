# Neko

An(other) uptime monitor.

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

| Key         | Description                              |
| ----------- | ---------------------------------------- |
| `interval`  | Amount of time in between probe attempts |
| `probe`     | A [probe](#probes) configuration         |
| `notifiers` | A list of [notifiers](#notifiers)        |

### Notifiers

#### Discord Webhook

```yaml
type: discord_webhook
config:
  url: https://discord.com/api/webhooks/webhook_id/webhook_token
```

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

### Probes

#### Ping

```yaml
type: ping
config:
  address: some-machine.example.com
```

| Key              | Description |
| ---------------- | ----------- |
| `address` |             |

#### HTTP

```yaml
type: http
config:
  address: https://example.com
  maxRedirects: 1
```

| Key                   | Description                         |
| --------------------- | ----------------------------------- |
| `address`      |                                     |
| `maxRedirects` | Maximum number of allowed redirects |

#### SSH

```yaml
type: ssh
config:
  host: some-machine.example.com
```

#### SQL

#### Domain

##### Extras

| Key         | Value                       |
| ----------- | --------------------------- |
| `remaining` | Amount of time until expiry |
