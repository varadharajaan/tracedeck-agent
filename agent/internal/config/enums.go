package config

import (
	"sort"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type TransparencyMode string
type URLMode string
type FeatureMode string
type VideoIDMode string
type PathMode string
type WindowTitleMode string
type SoftwareInventoryMode string
type SensitiveCapabilityMode string
type ArchiveProvider string
type EmailProvider string
type PushProvider string
type Severity string
type OpenTelemetryProtocol string

var transparencyModes = enumSet[TransparencyMode](
	constants.TransparencyVisibleIndicatorRequired,
)

var urlModes = enumSet[URLMode](
	constants.URLModeDomainOnly,
	constants.URLModeFullURL,
)

var featureModes = enumSet[FeatureMode](
	constants.FeatureEnabled,
	constants.FeatureDisabled,
)

var videoIDModes = enumSet[VideoIDMode](
	constants.VideoIDModeNone,
	constants.VideoIDModeHashed,
)

var pathModes = enumSet[PathMode](
	constants.PathModeNone,
	constants.PathModeHashOnly,
	constants.PathModeFullPath,
)

var windowTitleModes = enumSet[WindowTitleMode](
	constants.WindowTitleModeNone,
)

var softwareInventoryModes = enumSet[SoftwareInventoryMode](
	constants.SoftwareInventoryModeMetadataOnly,
)

var sensitiveCapabilityModes = enumSet[SensitiveCapabilityMode](
	constants.SensitiveCapabilityDeny,
)

var archiveProviders = enumSet[ArchiveProvider](
	constants.ArchiveProviderNone,
	constants.ArchiveProviderS3,
)

var emailProviders = enumSet[EmailProvider](
	constants.EmailProviderNone,
	constants.EmailProviderSES,
	constants.EmailProviderSMTP,
)

var pushProviders = enumSet[PushProvider](
	constants.PushProviderNone,
	constants.PushProviderWebPush,
)

var openTelemetryProtocols = enumSet[OpenTelemetryProtocol](
	constants.OpenTelemetryProtocolOTLPHTTPJSON,
)

var severities = enumSet[Severity](
	constants.SeverityLow,
	constants.SeverityMedium,
	constants.SeverityHigh,
	constants.SeverityCritical,
)

func enumSet[T ~string](values ...string) map[T]struct{} {
	out := make(map[T]struct{}, len(values))
	for _, value := range values {
		out[T(value)] = struct{}{}
	}
	return out
}

func enumValues[T ~string](set map[T]struct{}) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, string(value))
	}
	sort.Strings(values)
	return values
}

func isAllowed[T ~string](value T, set map[T]struct{}) bool {
	_, ok := set[value]
	return ok
}
