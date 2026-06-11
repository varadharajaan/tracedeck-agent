package event

import "time"

type Event struct {
	ID        string
	Type      string
	Source    string
	Timestamp time.Time
	TenantID  string
	DeviceID  string
	HostName  string
	AppName   string
	ProcessID int32
	PathHash  string
	Metadata  map[string]string
}
