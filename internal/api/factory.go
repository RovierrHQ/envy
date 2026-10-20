package api

import (
	"fmt"
	"os"
	"path/filepath"
)

func DefaultMockRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".envy", "mock"), nil
}

func NewBackend(kind, apiURL, token string) (StorageBackend, error) {
	switch kind {
	case "mock", "":
		root, err := DefaultMockRoot()
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(root, 0o755); err != nil {
			return nil, err
		}
		return NewMockBackend(root), nil
	case "http":
		if apiURL == "" {
			return nil, fmt.Errorf("api: http backend requires api_url")
		}
		return NewHTTPBackend(apiURL, token), nil
	default:
		return nil, fmt.Errorf("api: unknown backend %q", kind)
	}
}
