package notifier

import "context"

type Notifier interface {
	Notify(context.Context, string, map[string]any) error
}
