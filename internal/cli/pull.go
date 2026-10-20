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
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Download ciphertext, decrypt, and write local env file",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			rel, err := cfg.EnvFile(flagEnv)
			if err != nil {
				return err
			}
			path := filepath.Join(root, rel)
			key, err := keyring.GetProjectKey(cfg.ProjectID)
			if err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					return fmt.Errorf("project key not in keyring; run envy unlock")
				}
				return err
			}
			backend, err := resolveBackend(cfg)
			if err != nil {
				return err
			}
			bundle, err := backend.GetSecret(context.Background(), cfg.ProjectID, flagEnv)
			if err != nil {
				if errors.Is(err, api.ErrNotFound) {
					return fmt.Errorf("no remote secret for env %q", flagEnv)
				}
				return err
			}
			plaintext, err := crypto.Open(bundle.Ciphertext, key)
			if err != nil {
				return err
			}

			st, err := config.LoadState(root)
			if err != nil {
				return err
			}
			localState := st.Get(flagEnv)
			if existing, err := os.ReadFile(path); err == nil {
				localHash := fsutil.ContentHash(existing)
				if localState.ContentHash != "" && localHash != localState.ContentHash && !flagForce {
					ok, err := confirm(fmt.Sprintf("%s has local changes since last pull; overwrite?", rel))
					if err != nil {
						return err
					}
					if !ok {
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
			st.Set(flagEnv, config.EnvState{
				Version:     bundle.Version,
				ContentHash: fsutil.ContentHash(plaintext),
			})
			if err := config.SaveState(root, st); err != nil {
				return err
			}
			fmt.Printf("Pulled env=%s version=%d → %s\n", flagEnv, bundle.Version, rel)
			return nil
		},
	}
	cmd.Flags().StringVar(&flagEnv, "env", "dev", "environment name")
	cmd.Flags().BoolVar(&flagForce, "force", false, "overwrite local file even if dirty")
	return cmd
}
