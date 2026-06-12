package redaction

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func HashValue(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func HashPath(path string) string {
	return HashValue(path)
}
