package schema

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type PolicySchemaVersion string

const (
	PolicySchemaV1Alpha1 PolicySchemaVersion = constants.PolicySchemaVersionV1Alpha1
)

func LatestPolicyVersion() PolicySchemaVersion {
	return PolicySchemaV1Alpha1
}

func SupportedPolicyVersions() []PolicySchemaVersion {
	return []PolicySchemaVersion{
		PolicySchemaV1Alpha1,
	}
}

func ParsePolicyVersion(value string) (PolicySchemaVersion, error) {
	version := PolicySchemaVersion(value)
	for _, supported := range SupportedPolicyVersions() {
		if version == supported {
			return version, nil
		}
	}
	return "", fmt.Errorf("unsupported policy schema version %q", value)
}

func GeneratePolicy(version PolicySchemaVersion) ([]byte, error) {
	doc, err := BuildPolicy(version)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal policy schema: %w", err)
	}
	return append(data, '\n'), nil
}

func BuildPolicy(version PolicySchemaVersion) (*jsonschema.Schema, error) {
	switch version {
	case PolicySchemaV1Alpha1:
		return buildPolicyV1Alpha1(), nil
	default:
		return nil, fmt.Errorf("unsupported policy schema version %q", version)
	}
}

func buildPolicyV1Alpha1() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            false,
	}

	doc := reflector.Reflect(&config.Policy{})
	doc.ID = jsonschema.ID(constants.PolicySchemaIDV1Alpha1)
	doc.Title = constants.PolicySchemaTitleV1Alpha1
	doc.Version = jsonschema.Version
	config.ApplySchemaEnums(doc)
	return doc
}
