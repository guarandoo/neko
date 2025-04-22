package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
)

type gotifyNotifier struct {
	url             string
	token           string
	titleTemplate   string
	messageTemplate string
}

func (n *gotifyNotifier) Notify(name string, data map[string]any) error {
	titleTemplate := template.New("")
	titleTemplate.Parse(n.titleTemplate)

	messageTemplate := template.New("")
	messageTemplate.Parse(n.messageTemplate)

	var titleBuf bytes.Buffer
	var messageBuf bytes.Buffer

	body := map[string]any{
		"message":  messageTemplate.Execute(&messageBuf, data),
		"priority": 2,
		"title":    titleTemplate.Execute(&titleBuf, data),
		"extras":   make(map[string]any),
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)
	url := fmt.Sprintf("%s%s", n.url, "/message")
	req, err := http.NewRequest(http.MethodPost, url, buf)
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
	return &gotifyNotifier{
		url:             options.Url,
		token:           options.Token,
		titleTemplate:   options.TitleTemplate,
		messageTemplate: options.MessageTemplate,
	}, nil
}
