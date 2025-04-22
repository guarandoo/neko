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
	titleTemplate   *template.Template
	messageTemplate *template.Template
}

func (n *gotifyNotifier) Notify(name string, data map[string]any) error {
	var titleBuf bytes.Buffer
	var messageBuf bytes.Buffer

	body := map[string]any{
		"message":  n.messageTemplate.Execute(&messageBuf, data),
		"priority": 2,
		"title":    n.titleTemplate.Execute(&titleBuf, data),
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
	titleTemplate := template.New("title")
	titleTemplate.Parse(options.TitleTemplate)
	messageTemplate := template.New("message")
	messageTemplate.Parse(options.MessageTemplate)

	return &gotifyNotifier{
		url:             options.Url,
		token:           options.Token,
		titleTemplate:   titleTemplate,
		messageTemplate: messageTemplate,
	}, nil
}
