package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadFile(path string) (*Policy, error) {
	cleanPath := filepath.Clean(path)
	root, err := os.OpenRoot(filepath.Dir(cleanPath))
	if err != nil {
		return nil, fmt.Errorf("open config root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	file, err := root.Open(filepath.Base(cleanPath))
	if err != nil {
		return nil, fmt.Errorf("read policy config %q: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read policy config %q: %w", path, err)
	}
	return Load(data)
}

func Load(data []byte) (*Policy, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return nil, fmt.Errorf("decode policy config: %w", err)
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &policy, nil
}
