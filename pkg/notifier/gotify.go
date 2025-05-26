package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
)

type gotifyNotifier struct {
	url             string
	token           string
	titleTemplate   *template.Template
	messageTemplate *template.Template
}

func (n *gotifyNotifier) Notify(ctx context.Context, name string, data map[string]any) error {
	var titleBuf bytes.Buffer
	if err := n.titleTemplate.Execute(&titleBuf, data); err != nil {
		return err
	}

	var messageBuf bytes.Buffer
	if err := n.messageTemplate.Execute(&messageBuf, data); err != nil {
		return err
	}

	body := map[string]any{
		"message":  messageBuf.String(),
		"priority": 2,
		"title":    titleBuf.String(),
		"extras":   make(map[string]any),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)
	url := fmt.Sprintf("%s%s", n.url, "/message")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return fmt.Errorf("unable to create Gotify request: %w", err)
	}

	req.Header.Add("X-Gotify-Key", n.token)
	req.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to make Gotify request: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errors.New("unable to make Discord request")
	}

	return nil
}

type GotifyOptions struct {
	Url             string
	Token           string
	TitleTemplate   string
	MessageTemplate string
}

func NewGotifyNotifier(options GotifyOptions) (Notifier, error) {
	titleTemplate, err := template.New("").Parse(options.TitleTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse title template: %w", err)
	}

	messageTemplate, err := template.New("").Parse(options.MessageTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse message template: %w", err)
	}

	return &gotifyNotifier{
		url:             options.Url,
		token:           options.Token,
		titleTemplate:   titleTemplate,
		messageTemplate: messageTemplate,
	}, nil
}
