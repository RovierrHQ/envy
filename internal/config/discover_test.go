package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeriveEnvName(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{".env", "dev", true},
		{".env.staging", "staging", true},
		{".env.local", "local", true},
		{".env.example", "", false},
		{".env.sample", "", false},
		{".env.template", "", false},
		{".env.dist", "", false},
		{"env", "", false},
		{".envrc", "", false},
		{"foo.env", "", false},
	}
	for _, tc := range cases {
		got, ok := DeriveEnvName(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("DeriveEnvName(%q)=(%q,%v) want (%q,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestDiscoverEnvFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{".env", ".env.staging", ".env.example", ".env.sample", "README.md", ".env.production"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, ".env.dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := DiscoverEnvFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{".env": true, ".env.staging": true, ".env.production": true}
	if len(got) != len(want) {
		t.Fatalf("got %v want keys %v", got, want)
	}
	for _, g := range got {
		if !want[g] {
			t.Errorf("unexpected %q", g)
		}
	}
}

func TestMergeDiscovered(t *testing.T) {
	c := &Config{Files: DefaultFiles()}
	changed, warnings := c.MergeDiscovered([]string{".env", ".env.ci", ".env.staging"})
	if !changed {
		t.Fatal("expected change for .env.ci")
	}
	if c.Files["ci"] != ".env.ci" {
		t.Fatalf("ci=%q", c.Files["ci"])
	}
	if len(warnings) != 0 {
		t.Fatalf("warnings=%v", warnings)
	}
	// collision: .env already mapped via dev, HasPath skips; invent collision by mapping
	c.Files["other"] = ".env.other"
	changed, warnings = c.MergeDiscovered([]string{".env.other"})
	if changed {
		t.Fatal("path already tracked should not change")
	}
	c2 := &Config{Files: map[string]string{"dev": ".env.custom"}}
	changed, warnings = c2.MergeDiscovered([]string{".env"})
	if changed {
		t.Fatal("should not overwrite env key")
	}
	if len(warnings) != 1 {
		t.Fatalf("want 1 warning got %v", warnings)
	}
	if c2.Files["dev"] != ".env.custom" {
		t.Fatalf("dev overwritten: %q", c2.Files["dev"])
	}
}

func TestEnvNamesSorted(t *testing.T) {
	c := &Config{Files: map[string]string{"prod": ".env.production", "dev": ".env", "staging": ".env.staging"}}
	got := c.EnvNames()
	want := []string{"dev", "prod", "staging"}
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v != %v", got, want)
		}
	}
}
