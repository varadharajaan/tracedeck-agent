package platform

import (
	"context"
	"testing"
)

func TestCurrentAdapterReportsCapabilities(t *testing.T) {
	t.Parallel()

	adapter := Current()
	if adapter.Name() == "" {
		t.Fatal("platform adapter name is required")
	}

	caps := adapter.Capabilities()
	if caps.OperatingSystem == "" {
		t.Fatal("operating system capability is required")
	}
	if !caps.LocalStorage {
		t.Fatal("local storage must be supported for the local agent")
	}

	if _, err := adapter.Hostname(context.Background()); err != nil {
		t.Fatalf("hostname: %v", err)
	}
}
