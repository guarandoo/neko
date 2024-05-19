package notifier

import (
	"fmt"
	"net/smtp"
)

type smtpNotifier struct {
	host       string
	port       int
	auth       smtp.Auth
	sender     string
	recipients []string
}

func (n *smtpNotifier) Notify(name string, reason string) error {
	msg := []byte(fmt.Sprintf("%v: %v", name, reason))
	if err := smtp.SendMail(fmt.Sprintf("%v:%v", n.host, n.port), n.auth, n.sender, n.recipients, msg); err != nil {
		return err
	}
	return nil
}

type SmtpNotifierOptions struct {
	Host       string
	Port       int
	Username   string
	Password   string
	Sender     string
	Recipients []string
}

func NewSmtpNotifier(options SmtpNotifierOptions) (Notifier, error) {
	auth := smtp.PlainAuth("", options.Username, options.Password, options.Host)
	return &smtpNotifier{
		host:       options.Host,
		port:       options.Port,
		auth:       auth,
		sender:     options.Sender,
		recipients: options.Recipients,
	}, nil
}
