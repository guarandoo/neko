package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type discordWebhookNotifier struct {
	url               string
	messageTemplate   string
	lastMessageId     *string
	persistentMessage bool
	state             map[string]string
}

type discordWebhookReply struct {
	Id string `json:"id"`
}

func (n *discordWebhookNotifier) editMessage(messageId string, content string) (*discordWebhookReply, error) {
	body := map[string]interface{}{
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

	defer res.Body.Close()

	j := discordWebhookReply{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return &j, nil
}

func sendMessage(url string, content string) (*discordWebhookReply, error) {
	body := map[string]interface{}{
		"content": content,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload: %w", err)
	}

	buf := bytes.NewBuffer(payload)

	url = fmt.Sprintf("%s?wait=true", url)
	res, err := http.Post(url, "application/json", buf)
	if err != nil {
		return nil, fmt.Errorf("unable to make Discord request: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("received non-success status code: %d", res.StatusCode)
	}

	defer res.Body.Close()
	j := discordWebhookReply{}
	err = json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body: %w", err)
	}

	return &j, nil
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
	now := time.Now()
	data["TimeNotify"] = now
	data["TimeNotifyUnix"] = now.Unix()

	var msgBuf bytes.Buffer
	if err := tpl.Execute(&msgBuf, data); err != nil {
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
			res, err = sendMessage(n.url, message)
			if err != nil {
				return err
			}
			n.lastMessageId = &res.Id
		}
	} else {
		res, err := sendMessage(n.url, message)
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
	return &discordWebhookNotifier{
		url:               options.Url,
		messageTemplate:   options.MessageTemplate,
		lastMessageId:     options.LastMessageId,
		persistentMessage: options.PersistentMessage,
		state:             make(map[string]string),
	}, nil
}
