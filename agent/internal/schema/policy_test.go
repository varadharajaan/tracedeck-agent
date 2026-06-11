package schema

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestGeneratePolicySchemaIsVersioned(t *testing.T) {
	t.Parallel()

	data, err := GeneratePolicy(PolicySchemaV1Alpha1)
	if err != nil {
		t.Fatalf("generate policy schema: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("schema should be valid JSON: %v", err)
	}

	if doc["$id"] != constants.PolicySchemaIDV1Alpha1 {
		t.Fatalf("unexpected schema id: %v", doc["$id"])
	}
	if !strings.Contains(string(data), constants.URLModeDomainOnly) {
		t.Fatalf("expected centralized enum value %q in schema", constants.URLModeDomainOnly)
	}
	if !strings.Contains(string(data), constants.SensitiveCapabilityDeny) {
		t.Fatalf("expected deny-only capability enum in schema")
	}
}

func TestBuildPolicyRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	_, err := BuildPolicy(PolicySchemaVersion("v9"))
	if err == nil {
		t.Fatal("expected unsupported schema version to fail")
	}
}
