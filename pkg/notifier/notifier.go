package notifier

type Notifier interface {
	Notify(instance string, name string, reason string) error
}
