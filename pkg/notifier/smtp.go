package notifier

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

type smtpNotifier struct {
	host            string
	port            int
	auth            smtp.Auth
	sender          string
	recipients      []string
	subjectTemplate *template.Template
	bodyTemplate    *template.Template
}

func (n *smtpNotifier) Notify(name string, data map[string]any) error {
	var subjectBuf bytes.Buffer
	n.subjectTemplate.Execute(&subjectBuf, data)

	var bodyBuf bytes.Buffer
	n.bodyTemplate.Execute(&bodyBuf, data)

	msg := ""
	msg += fmt.Sprintf("From: %v\r\n", n.sender)
	msg += fmt.Sprintf("To: %v\r\n", strings.Join(n.recipients, ","))
	msg += fmt.Sprintf("Subject: %v\r\n", subjectBuf.String())
	msg += "\r\n"
	msg += bodyBuf.String()
	addr := fmt.Sprintf("%v:%v", n.host, n.port)
	if err := smtp.SendMail(addr, n.auth, n.sender, n.recipients, []byte(msg)); err != nil {
		return err
	}
	return nil
}

type SmtpNotifierOptions struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Sender          string
	Recipients      []string
	SubjectTemplate string
	BodyTemplate    string
}

func NewSmtpNotifier(options SmtpNotifierOptions) (Notifier, error) {
	subjectTemplate := template.New("")
	if _, err := subjectTemplate.Parse(options.SubjectTemplate); err != nil {
		return nil, fmt.Errorf("unable to parse subject template: %w", err)
	}

	bodyTemplate := template.New("")
	if _, err := bodyTemplate.Parse(options.BodyTemplate); err != nil {
		return nil, fmt.Errorf("unable to parse body template: %w", err)
	}

	auth := smtp.PlainAuth("", options.Username, options.Password, options.Host)
	return &smtpNotifier{
		host:            options.Host,
		port:            options.Port,
		auth:            auth,
		sender:          options.Sender,
		recipients:      options.Recipients,
		subjectTemplate: subjectTemplate,
		bodyTemplate:    bodyTemplate,
	}, nil
}
