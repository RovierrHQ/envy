package cli

import (
	"fmt"

	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Save an API bearer token in the OS credential store",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := promptSecret("API token: ")
			if err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("token required")
			}
			if err := keyring.SetAuthToken(token); err != nil {
				return err
			}
			fmt.Println("Logged in; token stored in OS keyring")
			return nil
		},
	}
}
