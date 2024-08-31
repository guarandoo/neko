# Neko

An(other) uptime monitor.

## Configuration

The configuration is loaded from a YAML file named `config.yaml` in the working directory.

> ❕ This path can be controlled by setting the `NEKO_CONFIG` environment variable.

See [values.yaml](values.yaml) for an example.

### Distributed Mode

You can run multiple instances across machines to provide redundancy or monitor your targets from multiple vantage points.

### Notifiers

#### Discord Webhook

```yaml
my_discord_notifier:
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
interval: 60
probe:
  type: ping
  config:
    address: some-machine.example.com
notifiers:
  - my_discord_notifier
  - my_smtp_notifier
```

| Key              | Description |
| ---------------- | ----------- |
| `config.address` |             |

#### HTTP

```yaml
interval: 10
probe:
  type: http
  config:
    address: https://example.com
    maxRedirects: 1
notifiers:
  - my_discord_notifier
  - my_smtp_notifier
```

| Key                   | Description                         |
| --------------------- | ----------------------------------- |
| `config.address`      |                                     |
| `config.maxRedirects` | Maximum number of allowed redirects |

#### SSH

```yaml
interval: 60
probe:
  type: ssh
  config:
    host: some-machine.example.com
notifiers:
  - my_discord_notifier
  - my_smtp_notifier
```

#### SQL

#### Domain

##### Extras

| Key         | Value                       |
| ----------- | --------------------------- |
| `remaining` | Amount of time until expiry |
