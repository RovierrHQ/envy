package cli

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/crypto"
	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		apiURL    string
		projectID string
		register  bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .envy.json in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			cfgPath := filepath.Join(cwd, config.ConfigFileName)
			if _, err := os.Stat(cfgPath); err == nil {
				return fmt.Errorf("%s already exists", config.ConfigFileName)
			}

			pass, err := promptSecret("Project passphrase: ")
			if err != nil {
				return err
			}
			if pass == "" {
				return fmt.Errorf("passphrase required")
			}
			pass2, err := promptSecret("Confirm passphrase: ")
			if err != nil {
				return err
			}
			if pass != pass2 {
				return fmt.Errorf("passphrases do not match")
			}

			salt, err := crypto.GenerateSalt()
			if err != nil {
				return err
			}
			key := crypto.DeriveKey(pass, salt)

			cfg := &config.Config{
				ProjectID:   projectID,
				APIURL:      apiURL,
				Salt:        base64.StdEncoding.EncodeToString(salt),
				KeyVerifier: crypto.Verifier(key),
				Files:       config.DefaultFiles(),
			}

			backend, err := resolveBackend(cfg)
			if err != nil {
				return err
			}

			if cfg.ProjectID == "" {
				if register || flagBackend == "mock" || cfg.APIURL == "" {
					id, err := backend.CreateProject(context.Background())
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: could not register project (%v); generating local id\n", err)
						cfg.ProjectID = base64.RawURLEncoding.EncodeToString(salt)[:16]
					} else {
						cfg.ProjectID = id
					}
				} else {
					return fmt.Errorf("project_id required (pass --project-id or --register)")
				}
			}

			if err := config.Save(cwd, cfg); err != nil {
				return err
			}
			if err := keyring.SetProjectKey(cfg.ProjectID, key); err != nil {
				return fmt.Errorf("save key to keyring: %w", err)
			}
			fmt.Printf("Initialized %s (project_id=%s)\n", config.ConfigFileName, cfg.ProjectID)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "remote API base URL")
	cmd.Flags().StringVar(&projectID, "project-id", "", "existing project id")
	cmd.Flags().BoolVar(&register, "register", false, "POST /v1/projects to allocate a project_id")
	return cmd
}
