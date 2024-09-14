package notifier

type Notifier interface {
	Notify(name string, data map[string]interface{}) error
}
