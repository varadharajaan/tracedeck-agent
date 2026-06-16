package platform

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

var ErrUnsupportedCapability = errors.New("unsupported platform capability")
var ErrNoForegroundApp = errors.New("foreground app unavailable")

type Adapter interface {
	Name() string
	Hostname(ctx context.Context) (string, error)
	Capabilities() Capabilities
	ForegroundApp(ctx context.Context) (ForegroundApp, error)
	SoftwareInventory(ctx context.Context) ([]InstalledSoftware, error)
}

type ForegroundApp struct {
	AppName        string
	ProcessID      int32
	ExecutablePath string
}

type InstalledSoftware struct {
	ID        string
	Name      string
	Version   string
	Publisher string
	Source    string
}

type Capabilities struct {
	OperatingSystem   string
	ServiceManager    string
	ProcessCollection bool
	LocalStorage      bool
	Features          []CapabilitySupport
}

type CapabilitySupport struct {
	ID                 string `json:"id"`
	Status             string `json:"status"`
	PermissionRequired bool   `json:"permission_required"`
	Notes              string `json:"notes"`
}

type CapabilityError struct {
	OperatingSystem string
	CapabilityID    string
	Status          string
	Reason          string
}

func (e CapabilityError) Error() string {
	return fmt.Sprintf("%s on %s is %s: %s", e.CapabilityID, e.OperatingSystem, e.Status, e.Reason)
}

func (e CapabilityError) Unwrap() error {
	return ErrUnsupportedCapability
}

func (c Capabilities) SupportFor(capabilityID string) (CapabilitySupport, bool) {
	capabilityID = strings.TrimSpace(capabilityID)
	for _, feature := range c.Features {
		if feature.ID == capabilityID {
			return feature, true
		}
	}
	return CapabilitySupport{}, false
}

func (c Capabilities) Require(capabilityID string) error {
	feature, ok := c.SupportFor(capabilityID)
	if ok && feature.Status == constants.PlatformSupportSupported {
		return nil
	}
	if !ok {
		feature = CapabilitySupport{
			ID:     strings.TrimSpace(capabilityID),
			Status: constants.PlatformSupportUnsupported,
			Notes:  "capability is not declared by this platform adapter",
		}
	}
	return CapabilityError{
		OperatingSystem: c.OperatingSystem,
		CapabilityID:    feature.ID,
		Status:          feature.Status,
		Reason:          feature.Notes,
	}
}

func unsupportedForegroundApp(capabilities Capabilities) (ForegroundApp, error) {
	return ForegroundApp{}, capabilities.Require(constants.PlatformCapabilityForegroundApp)
}

func unsupportedSoftwareInventory(capabilities Capabilities) ([]InstalledSoftware, error) {
	return nil, capabilities.Require(constants.PlatformCapabilitySoftwareInventory)
}
