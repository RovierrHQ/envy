package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var discoverExcludeExact = map[string]struct{}{
	".env.example":  {},
	".env.sample":   {},
	".env.template": {},
	".env.dist":     {},
}

// DiscoverEnvFiles lists .env / .env.* filenames in root (non-recursive).
func DiscoverEnvFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !isDiscoverableEnvFile(name) {
			continue
		}
		out = append(out, name)
	}
	return out, nil
}

func isDiscoverableEnvFile(name string) bool {
	if name == ".env" {
		return true
	}
	if !strings.HasPrefix(name, ".env.") {
		return false
	}
	if _, skip := discoverExcludeExact[name]; skip {
		return false
	}
	if strings.HasSuffix(name, ".example") {
		return false
	}
	// single segment after .env. (no further path sep; ReadDir already flat)
	rest := strings.TrimPrefix(name, ".env.")
	if rest == "" || strings.Contains(rest, string(filepath.Separator)) {
		return false
	}
	return true
}

// DeriveEnvName maps a discovered filename to an env key.
// .env → dev; .env.staging → staging. ok=false if not a discoverable name.
func DeriveEnvName(filename string) (string, bool) {
	base := filepath.Base(filename)
	if !isDiscoverableEnvFile(base) {
		return "", false
	}
	if base == ".env" {
		return "dev", true
	}
	return strings.TrimPrefix(base, ".env."), true
}

// MergeDiscovered adds untracked discovered paths into Files.
// Returns whether Files changed and any collision warnings.
func (c *Config) MergeDiscovered(paths []string) (changed bool, warnings []string) {
	for _, path := range paths {
		base := filepath.Base(path)
		if c.HasPath(base) {
			continue
		}
		env, ok := DeriveEnvName(base)
		if !ok {
			continue
		}
		if !c.AddFile(env, base) {
			warnings = append(warnings, fmt.Sprintf("skip %s: env %q already maps to %s", base, env, c.Files[env]))
			continue
		}
		changed = true
	}
	return changed, warnings
}
