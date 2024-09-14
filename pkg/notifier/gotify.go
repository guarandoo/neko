package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type gotifyNotifier struct {
	url   string
	token string
}

func (n *gotifyNotifier) Notify(name string, data map[string]interface{}) error {
	body := map[string]interface{}{
		"message":  fmt.Sprintf("%s: %s", name, data["Status"]),
		"priority": 2,
		"title":    "Monitor Status Change",
		"extras":   make(map[string]interface{}),
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
	Url   string
	Token string
}

func NewGotifyNotifier(options GotifyOptions) (Notifier, error) {
	return &gotifyNotifier{url: options.Url, token: options.Token}, nil
}
