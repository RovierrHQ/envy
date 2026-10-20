package api

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestMockOptimisticConcurrency(t *testing.T) {
	root := t.TempDir()
	m := NewMockBackend(root)
	ctx := context.Background()
	id, err := m.CreateProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	v1, err := m.PutSecret(ctx, id, "dev", []byte("blob-a"), 0)
	if err != nil || v1 != 1 {
		t.Fatalf("v1=%d err=%v", v1, err)
	}
	_, err = m.PutSecret(ctx, id, "dev", []byte("blob-b"), 0)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("want conflict got %v", err)
	}
	v2, err := m.PutSecret(ctx, id, "dev", []byte("blob-b"), 1)
	if err != nil || v2 != 2 {
		t.Fatalf("v2=%d err=%v", v2, err)
	}
	got, err := m.GetSecret(ctx, id, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 2 || string(got.Ciphertext) != "blob-b" {
		t.Fatalf("got %+v", got)
	}
	_, err = m.GetSecret(ctx, id, "prod")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want not found got %v", err)
	}
	_ = filepath.Join(root, "projects", id)
}
