package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/RovierrHQ/envy/internal/api"
	"github.com/RovierrHQ/envy/internal/config"
	"github.com/RovierrHQ/envy/internal/keyring"
	"golang.org/x/term"
)

func resolveRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return config.FindRoot(cwd)
}

func resolveBackend(cfg *config.Config) (api.StorageBackend, error) {
	kind := flagBackend
	if kind == "" {
		if cfg != nil && cfg.APIURL != "" {
			kind = "http"
		} else {
			kind = "mock"
		}
	}
	token := ""
	if kind == "http" {
		var err error
		token, err = keyring.GetAuthToken()
		if err != nil {
			return nil, fmt.Errorf("not logged in: run envy login (%w)", err)
		}
	}
	apiURL := ""
	if cfg != nil {
		apiURL = cfg.APIURL
	}
	return api.NewBackend(kind, apiURL, token)
}

func promptLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptSecret(prompt string) (string, error) {
	if v := os.Getenv("ENVY_PASSPHRASE"); v != "" && (prompt == "Project passphrase: " || prompt == "Confirm passphrase: ") {
		return v, nil
	}
	if v := os.Getenv("ENVY_API_TOKEN"); v != "" && prompt == "API token: " {
		return v, nil
	}
	fmt.Fprint(os.Stderr, prompt)
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		line, err := promptLine("")
		return line, err
	}
	b, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func confirm(prompt string) (bool, error) {
	ans, err := promptLine(prompt + " [y/N]: ")
	if err != nil {
		return false, err
	}
	ans = strings.ToLower(strings.TrimSpace(ans))
	return ans == "y" || ans == "yes", nil
}
