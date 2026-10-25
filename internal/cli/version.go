package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set via GoReleaser ldflags (-X .../internal/cli.Version=...).
var Version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print envy version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
