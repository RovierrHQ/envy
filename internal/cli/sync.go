package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/crypto"
	"github.com/RovierrHQ/envy/internal/fsutil"
	"github.com/RovierrHQ/envy/internal/keyring"
)

func resolveEnvs(cfg *config.Config, envFlag string) ([]string, error) {
	if envFlag != "" {
		if _, err := cfg.EnvFile(envFlag); err != nil {
			return nil, err
		}
		return []string{envFlag}, nil
	}
	names := cfg.EnvNames()
	if len(names) == 0 {
		return nil, fmt.Errorf("config: no environments in files map")
	}
	return names, nil
}

func projectKey(cfg *config.Config) ([]byte, error) {
	key, err := keyring.GetProjectKey(cfg.ProjectID)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, fmt.Errorf("project key not in keyring; run envy unlock")
		}
		return nil, err
	}
	return key, nil
}

// pushOne encrypts and uploads one env.
// skipClean: skip when local hash matches last push (all-mode).
// softMissing: warn and skip when file missing (all-mode); else return error.
func pushOne(root string, cfg *config.Config, backend api.StorageBackend, key []byte, st *config.State, env string, skipClean, softMissing bool) error {
	rel, err := cfg.EnvFile(env)
	if err != nil {
		return err
	}
	path := filepath.Join(root, rel)
	plaintext, err := os.ReadFile(path)
	if err != nil {
		if softMissing && os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "skip %s: missing local file\n", rel)
			return nil
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	hash := fsutil.ContentHash(plaintext)
	if skipClean && st.Get(env).ContentHash == hash {
		fmt.Printf("skip %s (env=%s): unchanged\n", rel, env)
		return nil
	}
	blob, err := crypto.Seal(plaintext, key)
	if err != nil {
		return err
	}
	base := st.Get(env).Version
	newVer, err := backend.PutSecret(context.Background(), cfg.ProjectID, env, blob, base)
	if err != nil {
		var ce *api.ConflictError
		if errors.As(err, &ce) {
			return ce
		}
		return err
	}
	st.Set(env, config.EnvState{
		Version:     newVer,
		ContentHash: hash,
	})
	fmt.Printf("Pushed %s → env=%s version=%d\n", rel, env, newVer)
	return nil
}

// pullOne downloads and writes one env.
// skipClean: skip when remote version matches local and file hash matches (all-mode).
func pullOne(root string, cfg *config.Config, backend api.StorageBackend, key []byte, st *config.State, env string, skipClean, force bool) error {
	rel, err := cfg.EnvFile(env)
	if err != nil {
		return err
	}
	path := filepath.Join(root, rel)
	bundle, err := backend.GetSecret(context.Background(), cfg.ProjectID, env)
	if err != nil {
		if errors.Is(err, api.ErrNotFound) {
			if skipClean {
				fmt.Fprintf(os.Stderr, "skip env=%s: no remote secret\n", env)
				return nil
			}
			return fmt.Errorf("no remote secret for env %q", env)
		}
		return err
	}
	localState := st.Get(env)
	if skipClean && bundle.Version == localState.Version {
		if existing, err := os.ReadFile(path); err == nil {
			if localState.ContentHash != "" && fsutil.ContentHash(existing) == localState.ContentHash {
				fmt.Printf("skip %s (env=%s): up to date\n", rel, env)
				return nil
			}
		}
	}

	plaintext, err := crypto.Open(bundle.Ciphertext, key)
	if err != nil {
		return err
	}

	if existing, err := os.ReadFile(path); err == nil {
		localHash := fsutil.ContentHash(existing)
		if localState.ContentHash != "" && localHash != localState.ContentHash && !force {
			ok, err := confirm(fmt.Sprintf("%s has local changes since last pull; overwrite?", rel))
			if err != nil {
				return err
			}
			if !ok {
				if skipClean {
					fmt.Fprintf(os.Stderr, "skip %s: local changes kept\n", rel)
					return nil
				}
				return fmt.Errorf("aborted; pass --force to overwrite")
			}
		}
		if err := fsutil.BackupFile(path); err != nil {
			return err
		}
	}

	if err := fsutil.AtomicWriteFile(path, plaintext, 0o600); err != nil {
		return err
	}
	st.Set(env, config.EnvState{
		Version:     bundle.Version,
		ContentHash: fsutil.ContentHash(plaintext),
	})
	fmt.Printf("Pulled env=%s version=%d → %s\n", env, bundle.Version, rel)
	return nil
}

func discoverAndMerge(root string, cfg *config.Config) error {
	paths, err := config.DiscoverEnvFiles(root)
	if err != nil {
		return err
	}
	changed, warnings := cfg.MergeDiscovered(paths)
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
	if !changed {
		return nil
	}
	if err := config.Save(root, cfg); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "updated %s with discovered env files\n", config.ConfigFileName)
	return nil
}
