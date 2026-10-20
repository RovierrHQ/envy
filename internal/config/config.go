package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/RovierrHQ/envy/internal/fsutil"
)

const (
	ConfigFileName = ".envy.json"
	StateDirName   = ".envy"
	StateFileName  = "state.json"
)

var ErrNotInitialized = errors.New("config: no .envy.json found; run envy init")

type Config struct {
	ProjectID   string            `json:"project_id"`
	APIURL      string            `json:"api_url"`
	Salt        string            `json:"salt"` // base64
	KeyVerifier string            `json:"key_verifier"`
	Files       map[string]string `json:"files"`
}

func DefaultFiles() map[string]string {
	return map[string]string{
		"dev":     ".env",
		"staging": ".env.staging",
		"prod":    ".env.production",
	}
}

func (c *Config) SaltBytes() ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(c.Salt)
	if err != nil {
		return nil, fmt.Errorf("config: decode salt: %w", err)
	}
	return b, nil
}

func (c *Config) EnvFile(env string) (string, error) {
	if c.Files == nil {
		return "", fmt.Errorf("config: no files map")
	}
	f, ok := c.Files[env]
	if !ok || f == "" {
		return "", fmt.Errorf("config: unknown env %q", env)
	}
	return f, nil
}

func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, ConfigFileName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInitialized
		}
		dir = parent
	}
}

func Load(root string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(root, ConfigFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotInitialized
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("config: parse: %w", err)
	}
	if c.Files == nil {
		c.Files = DefaultFiles()
	}
	return &c, nil
}

func Save(root string, c *Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return fsutil.AtomicWriteFile(filepath.Join(root, ConfigFileName), data, 0o644)
}

type EnvState struct {
	Version     int    `json:"version"`
	ContentHash string `json:"content_hash"`
}

type State struct {
	Envs map[string]EnvState `json:"envs"`
}

func statePath(root string) string {
	return filepath.Join(root, StateDirName, StateFileName)
}

func LoadState(root string) (*State, error) {
	data, err := os.ReadFile(statePath(root))
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Envs: map[string]EnvState{}}, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.Envs == nil {
		s.Envs = map[string]EnvState{}
	}
	return &s, nil
}

func SaveState(root string, s *State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return fsutil.AtomicWriteFile(statePath(root), data, 0o600)
}

func (s *State) Get(env string) EnvState {
	if s.Envs == nil {
		return EnvState{}
	}
	return s.Envs[env]
}

func (s *State) Set(env string, st EnvState) {
	if s.Envs == nil {
		s.Envs = map[string]EnvState{}
	}
	s.Envs[env] = st
}
