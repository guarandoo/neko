package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
)

type discordWebhookNotifier struct {
	url               string
	messageTemplate   *template.Template
	lastMessageId     *string
	persistentMessage bool
	state             map[string]string
}

type discordWebhookReply struct {
	Id string `json:"id"`
}

func (n *discordWebhookNotifier) editMessage(messageId string, content string) (*discordWebhookReply, error) {
	var err error

	body := map[string]any{
		"content": content,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)
	url := fmt.Sprintf("%s/messages/%s", n.url, messageId)
	req, err := http.NewRequest(http.MethodPatch, url, buf)
	if err != nil {
		return nil, fmt.Errorf("unable to make Discord request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make Discord request: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("received non-success status code: %d", res.StatusCode)
	}

	defer func() { err = res.Body.Close() }()

	j := discordWebhookReply{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return &j, err
}

func sendMessage(ctx context.Context, url string, content string) (*discordWebhookReply, error) {
	var err error

	body := map[string]any{
		"content": content,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)

	url = fmt.Sprintf("%s?wait=true", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to make Discord request: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("received non-success status code: %d", res.StatusCode)
	}

	defer func() { err = res.Body.Close() }()

	j := discordWebhookReply{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return &j, err
}

func (n *discordWebhookNotifier) Notify(ctx context.Context, name string, data map[string]any) error {
	var msgBuf bytes.Buffer
	if err := n.messageTemplate.Execute(&msgBuf, data); err != nil {
		return err
	}

	n.state[name] = msgBuf.String()

	message := n.state[name]
	if n.persistentMessage {
		builder := strings.Builder{}
		for _, v := range n.state {
			builder.WriteString(v)
			builder.WriteString("\n")
		}
		message = builder.String()
	}

	if n.persistentMessage && n.lastMessageId != nil {
		res, err := n.editMessage(*n.lastMessageId, message)
		if err == nil {
			n.lastMessageId = &res.Id
		} else {
			res, err = sendMessage(ctx, n.url, message)
			if err != nil {
				return err
			}
			n.lastMessageId = &res.Id
		}
	} else {
		res, err := sendMessage(ctx, n.url, message)
		if err != nil {
			return err
		}
		n.lastMessageId = &res.Id
	}

	return nil
}

type DiscordWebhookOptions struct {
	Url               string
	MessageTemplate   string
	LastMessageId     *string
	PersistentMessage bool
}

func NewDiscordWebhookNotifier(options DiscordWebhookOptions) (Notifier, error) {
	messageTemplate, err := template.New("").Parse(options.MessageTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse message template: %w", err)
	}

	return &discordWebhookNotifier{
		url:               options.Url,
		messageTemplate:   messageTemplate,
		lastMessageId:     options.LastMessageId,
		persistentMessage: options.PersistentMessage,
		state:             make(map[string]string),
	}, nil
}
