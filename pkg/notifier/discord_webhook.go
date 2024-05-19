package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type discordWebhookNotifier struct {
	url string
}

func (n *discordWebhookNotifier) Notify(name string, reason string) error {
	body := map[string]interface{}{
		"content": fmt.Sprintf("%s: %s", name, reason),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)
	res, err := http.Post(n.url, "application/json", buf)
	if err != nil {
		return fmt.Errorf("unable to make Discord request: %w", err)
	}

	if res.StatusCode != 200 {
		return errors.New("unable to make Discord request")
	}

	return nil
}

type DiscordWebhookOptions struct {
	Url string
}

func NewDiscordWebhookNotifier(options DiscordWebhookOptions) (Notifier, error) {
	return &discordWebhookNotifier{url: options.Url}, nil
}
