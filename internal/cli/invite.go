package cli

import (
	"context"
	"fmt"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/keyring"
	"github.com/spf13/cobra"
)

func newInviteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invite",
		Short: "Create or accept a project invite",
	}
	cmd.AddCommand(newInviteCreateCmd(), newInviteAcceptCmd())
	return cmd
}

func newInviteCreateCmd() *cobra.Command {
	var projectID, role, apiURL string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a single-use invite (owner only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if role != "read" && role != "write" {
				return fmt.Errorf("--role must be read or write")
			}
			token, base, id, err := inviteContext(apiURL, projectID)
			if err != nil {
				return err
			}
			code, err := api.CreateInvite(context.Background(), base, token, id, role)
			if err != nil {
				return err
			}
			fmt.Println(code)
			return nil
		},
	}
	cmd.Flags().StringVar(&projectID, "project", "", "project id (default: .envy.json)")
	cmd.Flags().StringVar(&role, "role", "", "read or write")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API origin")
	_ = cmd.MarkFlagRequired("role")
	return cmd
}

func newInviteAcceptCmd() *cobra.Command {
	var apiURL string
	cmd := &cobra.Command{
		Use:   "accept <code>",
		Short: "Join a project with an invite code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, base, _, err := inviteContext(apiURL, "skip")
			if err != nil {
				return err
			}
			id, err := api.AcceptInvite(context.Background(), base, token, args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Joined project %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API origin")
	return cmd
}

func inviteContext(apiURL, projectID string) (token, base, id string, err error) {
	token, err = keyring.GetAuthToken()
	if err != nil {
		return "", "", "", fmt.Errorf("not logged in: run envy login (%w)", err)
	}
	base = apiURL
	if base == "" {
		if root, rerr := resolveRoot(); rerr == nil {
			if cfg, cerr := config.Load(root); cerr == nil && cfg.APIURL != "" {
				base = cfg.APIURL
			}
		}
	}
	if base == "" {
		base = api.DefaultAPIURL
	}
	id = projectID
	if id == "" || id == "skip" {
		if id == "skip" {
			return token, base, "", nil
		}
		root, rerr := resolveRoot()
		if rerr != nil {
			return "", "", "", fmt.Errorf("pass --project")
		}
		cfg, cerr := config.Load(root)
		if cerr != nil || cfg.ProjectID == "" {
			return "", "", "", fmt.Errorf("pass --project")
		}
		id = cfg.ProjectID
	}
	return token, base, id, nil
}
