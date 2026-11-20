package cli

import (
	"errors"
	"fmt"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/config"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Encrypt local env file(s) and upload ciphertext",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			cfg, err := config.Load(root)
			if err != nil {
				return err
			}
			all := flagEnv == ""
			if all {
				if err := discoverAndMerge(root, cfg); err != nil {
					return err
				}
			}
			envs, err := resolveEnvs(cfg, flagEnv)
			if err != nil {
				return err
			}
			key, err := projectKey(cfg)
			if err != nil {
				return err
			}
			backend, err := resolveBackend(cfg)
			if err != nil {
				return err
			}
			st, err := config.LoadState(root)
			if err != nil {
				return err
			}
			var failed int
			for _, env := range envs {
				err := pushOne(root, cfg, backend, key, st, env, all /*skipClean*/, all /*softMissing*/)
				if err != nil {
					if all {
						var ce *api.ConflictError
						if errors.As(err, &ce) {
							fmt.Fprintf(cmd.ErrOrStderr(), "env=%s: %v\n", env, ce)
						} else {
							fmt.Fprintf(cmd.ErrOrStderr(), "env=%s: %v\n", env, err)
						}
						failed++
						continue
					}
					return err
				}
			}
			if err := config.SaveState(root, st); err != nil {
				return err
			}
			if failed > 0 {
				return fmt.Errorf("%d env(s) failed to push", failed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagEnv, "env", "", "environment name (default: all)")
	return cmd
}
