package config

import (
	"github.com/invopop/jsonschema"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func ApplySchemaEnums(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	applyPropertyEnum(schema, constants.SchemaDefArchivePolicy, constants.SchemaPropProvider, enumValues(archiveProviders))
	applyPropertyEnum(schema, constants.SchemaDefBrowserCollection, constants.SchemaPropURLMode, enumValues(urlModes))
	applyPropertyEnum(schema, constants.SchemaDefBrowserCollection, constants.SchemaPropYouTubeClassification, enumValues(featureModes))
	applyPropertyEnum(schema, constants.SchemaDefBrowserCollection, constants.SchemaPropYouTubeVideoIDMode, enumValues(videoIDModes))
	applyPropertyEnum(schema, constants.SchemaDefCollectionPolicy, constants.SchemaPropTransparencyMode, enumValues(transparencyModes))
	applyPropertyEnum(schema, constants.SchemaDefEmailPolicy, constants.SchemaPropProvider, enumValues(emailProviders))
	applyPropertyEnum(schema, constants.SchemaDefEmailPolicy, constants.SchemaPropMinSeverity, enumValues(severities))
	applyPropertyEnum(schema, constants.SchemaDefForegroundCollection, constants.SchemaPropWindowTitleMode, enumValues(windowTitleModes))
	applyPropertyEnum(schema, constants.SchemaDefMediaCollection, constants.SchemaPropPathMode, enumValues(pathModes))
	applyPropertyEnum(schema, constants.SchemaDefOpenTelemetryPolicy, constants.SchemaPropProtocol, enumValues(openTelemetryProtocols))
	applyPropertyEnum(schema, constants.SchemaDefPushPolicy, constants.SchemaPropProvider, enumValues(pushProviders))
	applyPropertyEnum(schema, constants.SchemaDefPushPolicy, constants.SchemaPropMinSeverity, enumValues(severities))
	applyPropertyEnum(schema, constants.SchemaDefRuleSpec, constants.SchemaPropSeverity, enumValues(severities))
	applyPropertyEnum(schema, constants.SchemaDefSoftwareCollection, constants.SchemaPropInventoryMode, enumValues(softwareInventoryModes))
	applySensitiveCapabilityEnums(schema)
}

func applyPropertyEnum(schema *jsonschema.Schema, definitionName string, propertyName string, values []string) {
	definition := schema.Definitions[definitionName]
	if definition == nil || definition.Properties == nil {
		return
	}

	property, ok := definition.Properties.Get(propertyName)
	if !ok || property == nil {
		return
	}

	property.Enum = make([]any, 0, len(values))
	for _, value := range values {
		property.Enum = append(property.Enum, value)
	}
}

func applySensitiveCapabilityEnums(schema *jsonschema.Schema) {
	for _, propertyName := range []string{
		constants.SchemaPropCredentials,
		constants.SchemaPropKeystrokes,
		constants.SchemaPropCookies,
		constants.SchemaPropTokens,
		constants.SchemaPropPrivateMessages,
		constants.SchemaPropScreenshots,
	} {
		applyPropertyEnum(schema, constants.SchemaDefSensitiveCapabilities, propertyName, enumValues(sensitiveCapabilityModes))
	}
}
