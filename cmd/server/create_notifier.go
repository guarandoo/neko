package main

import (
	"fmt"

	"github.com/guarandoo/neko/pkg/notifier"
)

func createNotifier(nc *NotifierConfig) (notifier.Notifier, error) {
	var n notifier.Notifier
	var err error
	switch v := nc.Config.(type) {
	case SmtpNotifierCOnfig:
		n, err = notifier.NewSmtpNotifier(notifier.SmtpNotifierOptions{
			Host:       v.Host,
			Port:       v.Port,
			Username:   v.Username,
			Password:   v.Password,
			Sender:     v.Sender,
			Recipients: v.Recipients,
		})
	case DiscordWebhookNotifierConfig:
		messageTemplate := "{{.Name}} is now {{.Status}}, was {{.PreviousStatus}} for {{.Duration}}"
		if v.MessageTemplate != nil {
			messageTemplate = *v.MessageTemplate
		}
		reuseMessage := false
		var messageId *string = nil
		if v.ReuseMessage != nil {
			reuseMessage = v.ReuseMessage.Enable
			messageId = v.ReuseMessage.MessageId
		}

		n, err = notifier.NewDiscordWebhookNotifier(notifier.DiscordWebhookOptions{
			Url:               v.Url,
			MessageTemplate:   messageTemplate,
			PersistentMessage: reuseMessage,
			LastMessageId:     messageId,
		})
	case GotifyNotifierConfig:
		titleTemplate := "Monitor Status Change"
		if v.TitleTemplate != nil {
			titleTemplate = *v.TitleTemplate
		}
		messageTemplate := "{{.Name}}: {{.Status}}"
		if v.MessageTemplate != nil {
			messageTemplate = *v.MessageTemplate
		}
		n, err = notifier.NewGotifyNotifier(notifier.GotifyOptions{
			Url:             v.Url,
			Token:           v.Token,
			TitleTemplate:   titleTemplate,
			MessageTemplate: messageTemplate,
		})
	default:
		n = nil
		err = fmt.Errorf("unknown probe type: %s", nc.Type)
	}
	return n, err
}
