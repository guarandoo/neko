package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"text/template"
)

type discordWebhookNotifier struct {
	url             string
	messageTemplate string
}

func (n *discordWebhookNotifier) Notify(instance string, name string, reason string) error {
	tpl, err := template.New(name).Parse(n.messageTemplate)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["Instance"] = instance
	data["Name"] = name
	data["Reason"] = reason

	var msgBuf bytes.Buffer
	if err := tpl.Execute(&msgBuf, data); err != nil {
		return err
	}

	message := msgBuf.String()

	body := map[string]interface{}{
		"content": message,
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

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errors.New("unable to make Discord request")
	}

	return nil
}

type DiscordWebhookOptions struct {
	Url             string
	MessageTemplate string
}

func NewDiscordWebhookNotifier(options DiscordWebhookOptions) (Notifier, error) {
	return &discordWebhookNotifier{url: options.Url, messageTemplate: options.MessageTemplate}, nil
}
