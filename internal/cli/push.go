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

func newPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Encrypt local env file and upload ciphertext",
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
			plaintext, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", rel, err)
			}
			key, err := keyring.GetProjectKey(cfg.ProjectID)
			if err != nil {
				if errors.Is(err, keyring.ErrNotFound) {
					return fmt.Errorf("project key not in keyring; run envy unlock")
				}
				return err
			}
			blob, err := crypto.Seal(plaintext, key)
			if err != nil {
				return err
			}
			st, err := config.LoadState(root)
			if err != nil {
				return err
			}
			base := st.Get(flagEnv).Version
			backend, err := resolveBackend(cfg)
			if err != nil {
				return err
			}
			newVer, err := backend.PutSecret(context.Background(), cfg.ProjectID, flagEnv, blob, base)
			if err != nil {
				var ce *api.ConflictError
				if errors.As(err, &ce) {
					return ce
				}
				return err
			}
			st.Set(flagEnv, config.EnvState{
				Version:     newVer,
				ContentHash: fsutil.ContentHash(plaintext),
			})
			if err := config.SaveState(root, st); err != nil {
				return err
			}
			fmt.Printf("Pushed %s → env=%s version=%d\n", rel, flagEnv, newVer)
			return nil
		},
	}
	cmd.Flags().StringVar(&flagEnv, "env", "dev", "environment name")
	return cmd
}
