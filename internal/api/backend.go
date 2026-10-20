package api

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("api: not found")
	ErrConflict     = errors.New("api: version conflict")
	ErrUnauthorized = errors.New("api: unauthorized")
)

type SecretBundle struct {
	Ciphertext []byte
	Version    int
}

type StorageBackend interface {
	CreateProject(ctx context.Context) (projectID string, err error)
	PutSecret(ctx context.Context, projectID, env string, ciphertext []byte, baseVersion int) (newVersion int, err error)
	GetSecret(ctx context.Context, projectID, env string) (*SecretBundle, error)
}

type ConflictError struct {
	RemoteVersion int
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("stale state: someone else pushed first (remote version %d), please pull", e.RemoteVersion)
}

func (e *ConflictError) Unwrap() error { return ErrConflict }
