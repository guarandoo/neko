package notifier

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

type smtpNotifier struct {
	host            string
	port            int
	username        string
	password        string
	sender          string
	recipients      []string
	subjectTemplate *template.Template
	bodyTemplate    *template.Template
}

func sendMail(ctx context.Context, addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	errCh := make(chan error, 1)

	go func() {
		err := smtp.SendMail(addr, a, from, to, msg)
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (n *smtpNotifier) Notify(ctx context.Context, name string, data map[string]any) error {
	var subjectBuf bytes.Buffer
	if err := n.subjectTemplate.Execute(&subjectBuf, data); err != nil {
		return err
	}

	var bodyBuf bytes.Buffer
	if err := n.bodyTemplate.Execute(&bodyBuf, data); err != nil {
		return err
	}

	msg := ""
	msg += fmt.Sprintf("From: %v\r\n", n.sender)
	msg += fmt.Sprintf("To: %v\r\n", strings.Join(n.recipients, ","))
	msg += fmt.Sprintf("Subject: %v\r\n", subjectBuf.String())
	msg += "\r\n"
	msg += bodyBuf.String()
	addr := fmt.Sprintf("%v:%v", n.host, n.port)
	auth := smtp.PlainAuth("", n.username, n.password, n.host)
	if err := sendMail(ctx, addr, auth, n.sender, n.recipients, []byte(msg)); err != nil {
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
	subjectTemplate, err := template.New("").Parse(options.SubjectTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse subject template: %w", err)
	}

	bodyTemplate, err := template.New("").Parse(options.BodyTemplate)
	if err != nil {
		return nil, fmt.Errorf("unable to parse body template: %w", err)
	}

	return &smtpNotifier{
		host:            options.Host,
		port:            options.Port,
		username:        options.Username,
		password:        options.Password,
		sender:          options.Sender,
		recipients:      options.Recipients,
		subjectTemplate: subjectTemplate,
		bodyTemplate:    bodyTemplate,
	}, nil
}
