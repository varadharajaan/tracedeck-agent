package platform

import "context"

type Adapter interface {
	Name() string
	Hostname(ctx context.Context) (string, error)
	Capabilities() Capabilities
}

type Capabilities struct {
	OperatingSystem   string
	ProcessCollection bool
	LocalStorage      bool
}
