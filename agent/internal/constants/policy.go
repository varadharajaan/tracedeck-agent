package constants

const (
	PolicySchemaVersionV1Alpha1 = "v1alpha1"
	PolicySchemaIDV1Alpha1      = "https://schema.tracedeck.io/policy/v1alpha1/schema.json"
	PolicySchemaTitleV1Alpha1   = "TraceDeck Policy v1alpha1"
)

const (
	TransparencyVisibleIndicatorRequired = "visible_indicator_required"
)

const (
	URLModeDomainOnly = "domain_only"
	URLModeFullURL    = "full_url"
)

const (
	FeatureEnabled  = "enabled"
	FeatureDisabled = "disabled"
)

const (
	VideoIDModeNone   = "none"
	VideoIDModeHashed = "hashed"
)

const (
	PathModeNone     = "none"
	PathModeHashOnly = "hash_only"
	PathModeFullPath = "full_path"
)

const (
	SensitiveCapabilityCredentials     = "credentials"
	SensitiveCapabilityKeystrokes      = "keystrokes"
	SensitiveCapabilityCookies         = "cookies"
	SensitiveCapabilityTokens          = "tokens"
	SensitiveCapabilityPrivateMessages = "private_messages"
	SensitiveCapabilityScreenshots     = "screenshots"
)

const (
	SensitiveCapabilityDeny = "deny"
)

const (
	ArchiveProviderNone = "none"
	ArchiveProviderS3   = "s3"
)

const (
	EmailProviderNone = "none"
	EmailProviderSES  = "ses"
	EmailProviderSMTP = "smtp"
)

const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

const (
	DefaultTenantID = "family-varadha"
	DefaultDeviceID = "laptop-cousin-001"
	DefaultProfile  = "ai-btech-student"
)

const (
	DefaultArchiveBucket = "tracedeck-agent-family-varadha-996335889295-ap-south-1"
	DefaultAlertEmail    = "varathu09@gmail.com"
)

const (
	DefaultLocalTTLDays      = 90
	DefaultMaxLocalStorageMB = 2048
	DefaultUploadInterval    = "1h"
	DefaultAlertCooldownMins = 30
)
