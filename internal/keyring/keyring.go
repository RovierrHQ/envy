package keyring

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceProjectKey = "envy-project-key"
	serviceAuthToken  = "envy-auth-token"
	authAccount       = "default"
)

var ErrNotFound = errors.New("keyring: secret not found")

func SetProjectKey(projectID string, key []byte) error {
	if projectID == "" {
		return fmt.Errorf("keyring: empty project_id")
	}
	return keyring.Set(serviceProjectKey, projectID, base64.StdEncoding.EncodeToString(key))
}

func GetProjectKey(projectID string) ([]byte, error) {
	s, err := keyring.Get(serviceProjectKey, projectID)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("keyring: decode project key: %w", err)
	}
	return b, nil
}

func DeleteProjectKey(projectID string) error {
	err := keyring.Delete(serviceProjectKey, projectID)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func SetAuthToken(token string) error {
	return keyring.Set(serviceAuthToken, authAccount, token)
}

func GetAuthToken() (string, error) {
	s, err := keyring.Get(serviceAuthToken, authAccount)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return s, nil
}

func DeleteAuthToken() error {
	err := keyring.Delete(serviceAuthToken, authAccount)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
