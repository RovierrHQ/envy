package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteAndHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", ".env")
	data := []byte("A=1\n")
	if err := AtomicWriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("got %q", got)
	}
	h, err := HashFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if h != ContentHash(data) {
		t.Fatal("hash mismatch")
	}
	if err := BackupFile(path); err != nil {
		t.Fatal(err)
	}
	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(bak) != string(data) {
		t.Fatal("backup mismatch")
	}
}
