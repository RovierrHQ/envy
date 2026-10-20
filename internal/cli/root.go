package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagBackend string
	flagEnv     string
	flagForce   bool
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "envy",
		Short:         "End-to-end encrypted .env sync for developer teams",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&flagBackend, "backend", "", "storage backend: mock|http (default: mock if api_url empty, else http)")
	root.AddCommand(newInitCmd())
	root.AddCommand(newLoginCmd())
	root.AddCommand(newUnlockCmd())
	root.AddCommand(newPushCmd())
	root.AddCommand(newPullCmd())
	return root
}

func Execute() {
	if err := NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
