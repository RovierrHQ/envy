package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var tokenFlag, apiURL string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store an API token via device login or --token",
		RunE: func(cmd *cobra.Command, args []string) error {
			if apiURL == "" {
				apiURL = api.DefaultAPIURL
			}
			token := tokenFlag
			if token == "" {
				if v := getenvToken(); v != "" {
					token = v
				}
			}
			if token == "" {
				start, err := api.StartDevice(context.Background(), apiURL)
				if err != nil {
					return err
				}
				fmt.Printf("Open %s\nCode: %s\n", start.VerificationURI, start.UserCode)
				token, err = api.WaitForDevice(context.Background(), apiURL, start)
				if err != nil {
					return err
				}
			}
			if err := keyring.SetAuthToken(token); err != nil {
				return err
			}
			fmt.Println("Logged in; token stored in OS keyring")
			return nil
		},
	}
	cmd.Flags().StringVar(&tokenFlag, "token", "", "paste a PAT from /settings/tokens")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API origin (default "+api.DefaultAPIURL+")")
	return cmd
}

func getenvToken() string {
	return os.Getenv("ENVY_API_TOKEN")
}
