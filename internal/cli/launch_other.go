//go:build !windows

package cli

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

func shouldShowLaunchHelp() bool {
	if len(os.Args) != 1 {
		return false
	}
	return !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd()))
}

func showLaunchHelp() {
	fmt.Fprintln(os.Stderr, launchHelp)
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		time.Sleep(5 * time.Second)
	}
}
