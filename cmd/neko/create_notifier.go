package main

import (
	"fmt"

	"github.com/guarandoo/neko/pkg/notifier"
)

func createNotifier(nc *NotifierConfig) (notifier.Notifier, error) {
	var n notifier.Notifier
	var err error
	switch v := nc.Config.(type) {
	case SmtpNotifierConfig:
		n, err = notifier.NewSmtpNotifier(notifier.SmtpNotifierOptions{
			Host:            v.Host,
			Port:            v.Port,
			Username:        v.Username,
			Password:        v.Password,
			Sender:          v.Sender,
			Recipients:      v.Recipients,
			SubjectTemplate: v.SubjectTemplate,
			BodyTemplate:    v.BodyTemplate,
		})
	case DiscordWebhookNotifierConfig:
		reuseMessage := false
		var messageId *string = nil
		if v.ReuseMessage != nil {
			reuseMessage = v.ReuseMessage.Enable
			messageId = v.ReuseMessage.MessageId
		}

		n, err = notifier.NewDiscordWebhookNotifier(notifier.DiscordWebhookOptions{
			Url:               v.Url,
			MessageTemplate:   v.MessageTemplate,
			PersistentMessage: reuseMessage,
			LastMessageId:     messageId,
		})
	case GotifyNotifierConfig:
		n, err = notifier.NewGotifyNotifier(notifier.GotifyOptions{
			Url:             v.Url,
			Token:           v.Token,
			TitleTemplate:   v.TitleTemplate,
			MessageTemplate: v.MessageTemplate,
		})
	default:
		n = nil
		err = fmt.Errorf("unknown probe type: %s", nc.Type)
	}
	return n, err
}
