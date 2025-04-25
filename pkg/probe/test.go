package probe

import (
	"context"
	"os"
	"time"
)

func isUnderDebugger() bool {
	return os.Getenv("GODEBUG") != ""
}

func getContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if isUnderDebugger() {
		return context.WithCancel(parent)
	} else {
		return context.WithTimeout(parent, timeout)
	}
}
