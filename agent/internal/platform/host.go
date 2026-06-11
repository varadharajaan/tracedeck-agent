package platform

import (
	"context"
	"os"
)

func osHostname(_ context.Context) (string, error) {
	return os.Hostname()
}
