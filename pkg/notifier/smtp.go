package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
)

type smtpNotifier struct {
	host       string
	port       int
	auth       smtp.Auth
	sender     string
	recipients []string
}

func (n *smtpNotifier) Notify(name string, reason string) error {
	msg := ""
	msg += fmt.Sprintf("From: %v\n", n.sender)
	msg += fmt.Sprintf("To: %v\n", strings.Join(n.recipients, ","))
	msg += fmt.Sprintf("Subject: %v\n", "Monitor Status Change")
	msg += "\n"
	msg += fmt.Sprintf("%s: %s", name, reason)
	if err := smtp.SendMail(fmt.Sprintf("%v:%v", n.host, n.port), n.auth, n.sender, n.recipients, []byte(msg)); err != nil {
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
