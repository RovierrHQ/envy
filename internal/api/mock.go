package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/RovierrHQ/envy/internal/fsutil"
)

type mockRecord struct {
	Ciphertext string `json:"ciphertext"`
	Version    int    `json:"version"`
}

type MockBackend struct {
	Root string
	mu   sync.Mutex
}

func NewMockBackend(root string) *MockBackend {
	return &MockBackend{Root: root}
}

func (m *MockBackend) projectDir(id string) string {
	return filepath.Join(m.Root, "projects", id)
}

func (m *MockBackend) secretPath(projectID, env string) string {
	return filepath.Join(m.projectDir(projectID), "secrets", env+".json")
}

func (m *MockBackend) CreateProject(ctx context.Context) (string, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	id, err := newProjectID()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(m.projectDir(id), "secrets"), 0o755); err != nil {
		return "", err
	}
	meta := map[string]string{"project_id": id}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := fsutil.AtomicWriteFile(filepath.Join(m.projectDir(id), "meta.json"), append(data, '\n'), 0o644); err != nil {
		return "", err
	}
	return id, nil
}

func (m *MockBackend) PutSecret(ctx context.Context, projectID, env string, ciphertext []byte, baseVersion int) (int, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, err := os.Stat(m.projectDir(projectID)); err != nil {
		if os.IsNotExist(err) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	path := m.secretPath(projectID, env)
	var cur mockRecord
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &cur); err != nil {
			return 0, err
		}
		if cur.Version != baseVersion {
			return 0, &ConflictError{RemoteVersion: cur.Version}
		}
	} else if !os.IsNotExist(err) {
		return 0, err
	} else if baseVersion != 0 {
		return 0, &ConflictError{RemoteVersion: 0}
	}
	next := baseVersion + 1
	rec := mockRecord{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Version:    next,
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return 0, err
	}
	if err := fsutil.AtomicWriteFile(path, append(data, '\n'), 0o600); err != nil {
		return 0, err
	}
	return next, nil
}

func (m *MockBackend) GetSecret(ctx context.Context, projectID, env string) (*SecretBundle, error) {
	_ = ctx
	m.mu.Lock()
	defer m.mu.Unlock()
	data, err := os.ReadFile(m.secretPath(projectID, env))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var rec mockRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	ct, err := base64.StdEncoding.DecodeString(rec.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("mock: decode ciphertext: %w", err)
	}
	return &SecretBundle{Ciphertext: ct, Version: rec.Version}, nil
}

func newProjectID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
