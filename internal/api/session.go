package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultAPIURL = "https://envy-api.rovierr.workers.dev"

type DeviceStart struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func StartDevice(ctx context.Context, baseURL string) (*DeviceStart, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/device/code", strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api: device code: %s: %s", resp.Status, bytes.TrimSpace(b))
	}
	var out DeviceStart
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func PollDevice(ctx context.Context, baseURL, deviceCode string) (string, error) {
	body, _ := json.Marshal(map[string]string{"device_code": deviceCode})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/device/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if resp.StatusCode == http.StatusOK && payload.AccessToken != "" {
		return payload.AccessToken, nil
	}
	if payload.Error == "authorization_pending" {
		return "", nil
	}
	if payload.Error == "" {
		payload.Error = resp.Status
	}
	return "", fmt.Errorf("api: device token: %s", payload.Error)
}

func WaitForDevice(ctx context.Context, baseURL string, start *DeviceStart) (string, error) {
	interval := time.Duration(start.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)
	for {
		token, err := PollDevice(ctx, baseURL, start.DeviceCode)
		if err != nil {
			return "", err
		}
		if token != "" {
			return token, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("api: device login expired")
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

func CreateInvite(ctx context.Context, baseURL, token, projectID, role string) (string, error) {
	body, _ := json.Marshal(map[string]string{"role": role})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/projects/"+projectID+"/invites", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload struct {
		Code  string `json:"code"`
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if resp.StatusCode >= 300 {
		if payload.Error == "" {
			payload.Error = resp.Status
		}
		return "", fmt.Errorf("api: invite: %s", payload.Error)
	}
	return payload.Code, nil
}

func AcceptInvite(ctx context.Context, baseURL, token, code string) (string, error) {
	body, _ := json.Marshal(map[string]string{"code": code})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/invites/accept", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var payload struct {
		ProjectID string `json:"project_id"`
		Error     string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if resp.StatusCode >= 300 {
		if payload.Error == "" {
			payload.Error = resp.Status
		}
		return "", fmt.Errorf("api: accept: %s", payload.Error)
	}
	return payload.ProjectID, nil
}
