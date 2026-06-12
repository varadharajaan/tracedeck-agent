package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestSupportedPolicyVersions(t *testing.T) {
	t.Parallel()

	versions := SupportedPolicyVersions()
	if len(versions) != 1 || versions[0] != PolicySchemaV1Alpha1 {
		t.Fatalf("unexpected supported versions: %+v", versions)
	}
	if LatestPolicyVersion() != PolicySchemaV1Alpha1 {
		t.Fatalf("unexpected latest version: %s", LatestPolicyVersion())
	}
	parsed, err := ParsePolicyVersion(constants.PolicySchemaVersionV1Alpha1)
	if err != nil {
		t.Fatalf("parse supported version: %v", err)
	}
	if parsed != PolicySchemaV1Alpha1 {
		t.Fatalf("unexpected parsed version: %s", parsed)
	}
}

func TestGeneratedPolicySchemaMatchesCheckedInFile(t *testing.T) {
	t.Parallel()

	generated, err := GeneratePolicy(LatestPolicyVersion())
	if err != nil {
		t.Fatalf("generate policy schema: %v", err)
	}

	checkedInPath := filepath.Join("..", "..", "..", constants.GeneratedPolicySchemaPath)
	checkedIn, err := os.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("read checked-in schema: %v", err)
	}
	if normalizeSchemaText(string(generated)) != normalizeSchemaText(string(checkedIn)) {
		t.Fatalf("checked-in policy schema is stale; run schema generation for %s", LatestPolicyVersion())
	}
}

func TestBuildPolicyRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	_, err := BuildPolicy(PolicySchemaVersion("v9"))
	if err == nil {
		t.Fatal("expected unsupported schema version to fail")
	}
}

func TestParsePolicyVersionRejectsUnsupportedVersion(t *testing.T) {
	t.Parallel()

	_, err := ParsePolicyVersion("v9")
	if err == nil {
		t.Fatal("expected unsupported schema version to fail")
	}
}

func normalizeSchemaText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.TrimSpace(value)
}
