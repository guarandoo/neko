package notifier

type smtpNotifier struct {
}

func (n *smtpNotifier) Notify(name string, reason string) error {
	return nil
}

type SmtpNotifierOptions struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewSmtpNotifier(options SmtpNotifierOptions) (Notifier, error) {
	return &smtpNotifier{}, nil
}
