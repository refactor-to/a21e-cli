// device.go — CLI device login flow: open browser, poll for key.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const devicePollInterval = 2 * time.Second
const devicePollTimeout = 5 * time.Minute

type deviceStartResp struct {
	DeviceCode     string `json:"device_code"`
	UserCode       string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn      int    `json:"expires_in"`
}

type devicePollResp struct {
	Status  string `json:"status"`
	APIKey  string `json:"api_key"`
}

// startDeviceFlow runs the device authorization flow: POST device, show URL, poll until authorized.
// Returns the API key or an error.
func startDeviceFlow(baseURL string) (string, error) {
	url := strings.TrimSuffix(baseURL, "/") + "/v1/cli/device"
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var ae apiError
		_ = json.Unmarshal(raw, &ae)
		return "", fmt.Errorf("API %d: %s", resp.StatusCode, ae.Error)
	}

	var start deviceStartResp
	if err := json.Unmarshal(raw, &start); err != nil {
		return "", fmt.Errorf("invalid response: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\nOpen this URL to sign in and authorize this device:\n\n  %s\n\n", start.VerificationURI)
	openBrowser(start.VerificationURI)

	// Poll until authorized or timeout
	deadline := time.Now().Add(devicePollTimeout)
	pollURL := strings.TrimSuffix(baseURL, "/") + "/v1/cli/device?device_code=" + start.DeviceCode

	for time.Now().Before(deadline) {
		time.Sleep(devicePollInterval)

		req, err := http.NewRequest("GET", pollURL, nil)
		if err != nil {
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var poll devicePollResp
		if err := json.Unmarshal(raw, &poll); err != nil {
			continue
		}

		if poll.Status == "authorized" && poll.APIKey != "" {
			return poll.APIKey, nil
		}
		if poll.Status == "consumed" || poll.Status == "expired" {
			return "", fmt.Errorf("device code expired or already used")
		}
	}

	return "", fmt.Errorf("timed out waiting for authorization")
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		_ = exec.Command("open", url).Start()
	case "linux":
		_ = exec.Command("xdg-open", url).Start()
	case "windows":
		_ = exec.Command("cmd", "/c", "start", url).Start()
	}
}

// writeCredentialsFile writes the API key to ~/.a21e/credentials for reuse.
func writeCredentialsFile(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".a21e")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "credentials")
	content := "A21E_API_KEY=" + key + "\n"
	return os.WriteFile(path, []byte(content), 0600)
}
