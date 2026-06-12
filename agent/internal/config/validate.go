package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

var hourMinutePattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func (p Policy) Validate() error {
	var errs []error

	requiredString(&errs, constants.ConfigFieldTenantID, p.TenantID)
	requiredString(&errs, constants.ConfigFieldDeviceID, p.DeviceID)
	requiredString(&errs, constants.ConfigFieldProfile, p.Profile)

	requireEnum(&errs, constants.ConfigFieldTransparencyMode, p.Collection.TransparencyMode, transparencyModes)
	requireEnum(&errs, constants.ConfigFieldBrowserURLMode, p.Collection.Browser.URLMode, urlModes)
	requireEnum(&errs, constants.ConfigFieldYouTubeClassification, p.Collection.Browser.YouTubeClassification, featureModes)
	requireEnum(&errs, constants.ConfigFieldYouTubeVideoIDMode, p.Collection.Browser.YouTubeVideoIDMode, videoIDModes)
	requireEnum(&errs, constants.ConfigFieldMediaPathMode, p.Collection.Media.PathMode, pathModes)

	requireDenyOnly(&errs, constants.SensitiveCapabilityCredentials, p.Collection.SensitiveCapabilities.Credentials)
	requireDenyOnly(&errs, constants.SensitiveCapabilityKeystrokes, p.Collection.SensitiveCapabilities.Keystrokes)
	requireDenyOnly(&errs, constants.SensitiveCapabilityCookies, p.Collection.SensitiveCapabilities.Cookies)
	requireDenyOnly(&errs, constants.SensitiveCapabilityTokens, p.Collection.SensitiveCapabilities.Tokens)
	requireDenyOnly(&errs, constants.SensitiveCapabilityPrivateMessages, p.Collection.SensitiveCapabilities.PrivateMessages)
	requireDenyOnly(&errs, constants.SensitiveCapabilityScreenshots, p.Collection.SensitiveCapabilities.Screenshots)

	if p.Retention.LocalTTLDays <= 0 {
		errs = append(errs, fieldError(constants.ConfigFieldLocalTTLDays, constants.ConfigErrorMustBeGreaterThanZero))
	}
	if p.Retention.MaxLocalStorageMB <= 0 {
		errs = append(errs, fieldError(constants.ConfigFieldMaxLocalStorageMB, constants.ConfigErrorMustBeGreaterThanZero))
	}

	requireEnum(&errs, constants.ConfigFieldArchiveProvider, p.Archive.Provider, archiveProviders)
	if p.Archive.Enabled {
		if p.Archive.Provider == ArchiveProvider(constants.ArchiveProviderNone) {
			errs = append(errs, fieldError(constants.ConfigFieldArchiveProvider, constants.ConfigErrorArchiveProviderRequired))
		}
		requiredString(&errs, constants.ConfigFieldArchiveBucket, p.Archive.Bucket)
		requiredString(&errs, constants.ConfigFieldArchivePrefixTemplate, p.Archive.PrefixTemplate)
		requiredString(&errs, constants.ConfigFieldArchiveUploadInterval, p.Archive.UploadInterval)
		validateDuration(&errs, constants.ConfigFieldArchiveUploadInterval, p.Archive.UploadInterval)
	}
	if p.Archive.StorageClassDays.StandardIAUntil <= p.Archive.StorageClassDays.Standard {
		errs = append(errs, fieldError(constants.ConfigFieldArchiveStandardIAUntil, constants.ConfigErrorStandardIAAfterStandard))
	}
	if p.Archive.StorageClassDays.ArchiveAfter < p.Archive.StorageClassDays.StandardIAUntil {
		errs = append(errs, fieldError(constants.ConfigFieldArchiveAfter, constants.ConfigErrorArchiveAfterStandardIA))
	}

	requireEnum(&errs, constants.ConfigFieldEmailProvider, p.Alerts.Email.Provider, emailProviders)
	requireEnum(&errs, constants.ConfigFieldEmailMinSeverity, p.Alerts.Email.MinSeverity, severities)
	if p.Alerts.Enabled {
		if p.Alerts.Email.Provider == EmailProvider(constants.EmailProviderNone) {
			errs = append(errs, fieldError(constants.ConfigFieldEmailProvider, constants.ConfigErrorEmailProviderRequired))
		}
		if strings.TrimSpace(p.Alerts.Email.From) == "" {
			errs = append(errs, fieldError(constants.ConfigFieldEmailFrom, constants.ConfigErrorEmailSenderRequired))
		}
		if len(p.Alerts.Email.To) == 0 {
			errs = append(errs, fieldError(constants.ConfigFieldEmailTo, constants.ConfigErrorEmailRecipientRequired))
		}
		if p.Alerts.Email.CooldownMinutes <= 0 {
			errs = append(errs, fieldError(constants.ConfigFieldEmailCooldownMinutes, constants.ConfigErrorMustBeGreaterThanZero))
		}
	}

	validateThresholds(&errs, p.Thresholds)

	for ruleName, rule := range p.AlertRules {
		if strings.TrimSpace(ruleName) == "" {
			errs = append(errs, fieldError(constants.ConfigFieldAlertRules, constants.ConfigErrorRuleNameRequired))
		}
		requireEnum(&errs, constants.ConfigFieldAlertRuleSeverity, rule.Severity, severities)
		if rule.ThresholdMinutesPerDay < 0 {
			errs = append(errs, fieldError(constants.ConfigFieldAlertRuleThreshold, constants.ConfigErrorMustNotBeNegative))
		}
	}

	return errors.Join(errs...)
}

func requiredString(errs *[]error, fieldName, value string) {
	if strings.TrimSpace(value) == "" {
		*errs = append(*errs, fieldError(fieldName, constants.ConfigErrorMustNotBeEmpty))
	}
}

func requireEnum[T ~string](errs *[]error, fieldName string, value T, allowed map[T]struct{}) {
	if !isAllowed(value, allowed) {
		*errs = append(*errs, fieldError(fieldName, fmt.Sprintf(constants.ConfigErrorUnsupportedValueFormat, value)))
	}
}

func requireDenyOnly(errs *[]error, fieldName string, value SensitiveCapabilityMode) {
	if value != SensitiveCapabilityMode(constants.SensitiveCapabilityDeny) {
		*errs = append(*errs, fieldError(fieldName, constants.ConfigErrorSensitiveCapabilityDenyOnly))
	}
}

func fieldError(fieldName, message string) error {
	return fmt.Errorf("%s: %s", fieldName, message)
}

func validateThresholds(errs *[]error, thresholds ThresholdPolicy) {
	if thresholds.MaxVideoMinutesPerDay < 0 {
		*errs = append(*errs, fieldError(constants.ConfigFieldMaxVideoMinutes, constants.ConfigErrorMustNotBeNegative))
	}
	if thresholds.MaxSocialMinutesPerDay < 0 {
		*errs = append(*errs, fieldError(constants.ConfigFieldMaxSocialMinutes, constants.ConfigErrorMustNotBeNegative))
	}
	if thresholds.MaxUnknownAppMinutesPerDay < 0 {
		*errs = append(*errs, fieldError(constants.ConfigFieldMaxUnknownAppMinutes, constants.ConfigErrorMustNotBeNegative))
	}
	validateOptionalHourMinute(errs, constants.ConfigFieldLateNightUsageStart, thresholds.LateNightUsageStart)
	validateOptionalHourMinute(errs, constants.ConfigFieldLateNightUsageEnd, thresholds.LateNightUsageEnd)
}

func validateOptionalHourMinute(errs *[]error, fieldName, value string) {
	if value == "" {
		return
	}
	if !hourMinutePattern.MatchString(value) {
		*errs = append(*errs, fieldError(fieldName, constants.ConfigErrorTimeMustUseHourMinute))
	}
}

func validateDuration(errs *[]error, fieldName, value string) {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		*errs = append(*errs, fieldError(fieldName, constants.ConfigErrorDurationRequired))
	}
}
