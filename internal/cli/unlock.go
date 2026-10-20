package cli

import (
	"fmt"

	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/crypto"
	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newUnlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Derive the project key from passphrase and store it in the OS keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			salt, err := cfg.SaltBytes()
			if err != nil {
				return err
			}
			pass, err := promptSecret("Project passphrase: ")
			if err != nil {
				return err
			}
			key := crypto.DeriveKey(pass, salt)
			if !crypto.CheckVerifier(key, cfg.KeyVerifier) {
				return crypto.ErrVerifierMismatch
			}
			if err := keyring.SetProjectKey(cfg.ProjectID, key); err != nil {
				return err
			}
			fmt.Println("Unlocked; project key stored in OS keyring")
			return nil
		},
	}
}
