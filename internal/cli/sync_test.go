package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/crypto"
	"github.com/RovierrHQ/envy/internal/fsutil"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, crypto.KeySize)
	for i := range k {
		k[i] = byte(i + 1)
	}
	return k
}

func setupSyncFixture(t *testing.T) (root string, cfg *config.Config, backend *api.MockBackend, key []byte, st *config.State) {
	t.Helper()
	root = t.TempDir()
	key = testKey(t)
	backend = api.NewMockBackend(t.TempDir())
	ctx := context.Background()
	pid, err := backend.CreateProject(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cfg = &config.Config{
		ProjectID: pid,
		Files:     map[string]string{"dev": ".env", "staging": ".env.staging"},
	}
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("A=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.staging"), []byte("B=2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	st = &config.State{Envs: map[string]config.EnvState{}}
	return root, cfg, backend, key, st
}

func TestPushOneSkipClean(t *testing.T) {
	root, cfg, backend, key, st := setupSyncFixture(t)
	if err := pushOne(root, cfg, backend, key, st, "dev", false, false); err != nil {
		t.Fatal(err)
	}
	if st.Get("dev").Version != 1 {
		t.Fatalf("version=%d", st.Get("dev").Version)
	}
	// unchanged + skipClean → no version bump
	if err := pushOne(root, cfg, backend, key, st, "dev", true, false); err != nil {
		t.Fatal(err)
	}
	if st.Get("dev").Version != 1 {
		t.Fatalf("want still v1 got %d", st.Get("dev").Version)
	}
	// force push (skipClean=false) bumps
	if err := pushOne(root, cfg, backend, key, st, "dev", false, false); err != nil {
		t.Fatal(err)
	}
	if st.Get("dev").Version != 2 {
		t.Fatalf("want v2 got %d", st.Get("dev").Version)
	}
}

func TestPushOneSoftMissing(t *testing.T) {
	root, cfg, backend, key, st := setupSyncFixture(t)
	cfg.Files["ghost"] = ".env.ghost"
	if err := pushOne(root, cfg, backend, key, st, "ghost", true, true); err != nil {
		t.Fatal(err)
	}
	if err := pushOne(root, cfg, backend, key, st, "ghost", false, false); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPullOneSkipClean(t *testing.T) {
	root, cfg, backend, key, st := setupSyncFixture(t)
	if err := pushOne(root, cfg, backend, key, st, "dev", false, false); err != nil {
		t.Fatal(err)
	}
	// remove local file and clear hash so pull writes
	_ = os.Remove(filepath.Join(root, ".env"))
	st.Set("dev", config.EnvState{Version: 0, ContentHash: ""})
	if err := pullOne(root, cfg, backend, key, st, "dev", false, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".env"))
	if err != nil || string(data) != "A=1\n" {
		t.Fatalf("got %q err=%v", data, err)
	}
	ver := st.Get("dev").Version
	hash := st.Get("dev").ContentHash
	if hash != fsutil.ContentHash([]byte("A=1\n")) {
		t.Fatalf("hash=%s", hash)
	}
	if err := pullOne(root, cfg, backend, key, st, "dev", true, true); err != nil {
		t.Fatal(err)
	}
	if st.Get("dev").Version != ver {
		t.Fatal("skip clean should not change state")
	}
}

func TestResolveEnvs(t *testing.T) {
	cfg := &config.Config{Files: map[string]string{"staging": ".env.staging", "dev": ".env"}}
	all, err := resolveEnvs(cfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || all[0] != "dev" || all[1] != "staging" {
		t.Fatalf("%v", all)
	}
	one, err := resolveEnvs(cfg, "staging")
	if err != nil || len(one) != 1 || one[0] != "staging" {
		t.Fatalf("%v %v", one, err)
	}
	if _, err := resolveEnvs(cfg, "nope"); err == nil {
		t.Fatal("expected unknown env error")
	}
}

func TestDiscoverAndMergeOnPushPath(t *testing.T) {
	root := t.TempDir()
	cfg := &config.Config{
		ProjectID: "p",
		Files:     map[string]string{"dev": ".env"},
	}
	if err := config.Save(root, cfg); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("A=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env.ci"), []byte("C=3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := discoverAndMerge(root, cfg); err != nil {
		t.Fatal(err)
	}
	reloaded, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Files["ci"] != ".env.ci" {
		t.Fatalf("files=%v", reloaded.Files)
	}
}
