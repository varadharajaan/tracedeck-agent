package redaction

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashPath(path string) string {
	normalized := strings.TrimSpace(strings.ToLower(path))
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
