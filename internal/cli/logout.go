package cli

import (
	"fmt"

	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the API token from the OS keyring",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := keyring.DeleteAuthToken(); err != nil {
				return err
			}
			fmt.Println("Logged out")
			return nil
		},
	}
}
