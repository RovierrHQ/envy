package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPBackend struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewHTTPBackend(baseURL, token string) *HTTPBackend {
	return &HTTPBackend{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type createProjectResp struct {
	ProjectID string `json:"project_id"`
}

type secretJSON struct {
	Ciphertext string `json:"ciphertext"`
	Version    int    `json:"version"`
}

func (h *HTTPBackend) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, h.BaseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if h.Token != "" {
		req.Header.Set("Authorization", "Bearer "+h.Token)
	}
	return h.HTTPClient.Do(req)
}

func (h *HTTPBackend) CreateProject(ctx context.Context) (string, error) {
	resp, err := h.do(ctx, http.MethodPost, "/v1/projects", map[string]any{})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("api: create project: %s: %s", resp.Status, bytes.TrimSpace(b))
	}
	var out createProjectResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ProjectID == "" {
		return "", fmt.Errorf("api: empty project_id")
	}
	return out.ProjectID, nil
}

func (h *HTTPBackend) PutSecret(ctx context.Context, projectID, env string, ciphertext []byte, baseVersion int) (int, error) {
	path := fmt.Sprintf("/v1/projects/%s/secrets/%s", projectID, env)
	body := secretJSON{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Version:    baseVersion,
	}
	resp, err := h.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return 0, ErrUnauthorized
	case http.StatusNotFound:
		return 0, ErrNotFound
	case http.StatusConflict:
		var remote secretJSON
		_ = json.NewDecoder(resp.Body).Decode(&remote)
		return 0, &ConflictError{RemoteVersion: remote.Version}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("api: put secret: %s: %s", resp.Status, bytes.TrimSpace(b))
	}
	var out secretJSON
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, err
	}
	return out.Version, nil
}

func (h *HTTPBackend) GetSecret(ctx context.Context, projectID, env string) (*SecretBundle, error) {
	path := fmt.Sprintf("/v1/projects/%s/secrets/%s", projectID, env)
	resp, err := h.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusNotFound:
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api: get secret: %s: %s", resp.Status, bytes.TrimSpace(b))
	}
	var out secretJSON
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	ct, err := base64.StdEncoding.DecodeString(out.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("api: decode ciphertext: %w", err)
	}
	return &SecretBundle{Ciphertext: ct, Version: out.Version}, nil
}
