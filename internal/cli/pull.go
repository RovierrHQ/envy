package cli

import (
	"fmt"

	"github.com/RovierrHQ/envy/internal/config"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Download ciphertext, decrypt, and write local env file(s)",
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
				err := pullOne(root, cfg, backend, key, st, env, all /*skipClean*/, flagForce)
				if err != nil {
					if all {
						fmt.Fprintf(cmd.ErrOrStderr(), "env=%s: %v\n", env, err)
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
				return fmt.Errorf("%d env(s) failed to pull", failed)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&flagEnv, "env", "", "environment name (default: all)")
	cmd.Flags().BoolVar(&flagForce, "force", false, "overwrite local file even if dirty")
	return cmd
}
