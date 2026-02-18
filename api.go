package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var validToolIDs = []string{
	"codex_cli", "claude_code_cli", "cursor", "vscode", "jetbrains", "openai_cli_custom",
}

func isValidToolID(id string) bool {
	for _, t := range validToolIDs {
		if t == id {
			return true
		}
	}
	return false
}

type workspaceResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type defaultWorkspaceResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listWorkspacesResp struct {
	Items []workspaceResp `json:"items"`
}

type createCliKeyReq struct {
	ToolID     string `json:"tool_id"`
	Label      string `json:"label,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
	Scope      string `json:"scope,omitempty"`      // user (default), workspace, project
	ProjectID  string `json:"project_id,omitempty"` // required when scope=project
}

type createCliKeyResp struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	Prefix    string `json:"prefix"`
	Label     string `json:"label"`
	ToolID    string `json:"tool_id"`
	CreatedAt string `json:"created_at"`
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func apiRequest(apiKey, baseURL, method, path string, body interface{}) ([]byte, int, error) {
	url := strings.TrimSuffix(baseURL, "/") + path
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return raw, resp.StatusCode, nil
}

func getDefaultWorkspace(apiKey, baseURL string) (*defaultWorkspaceResp, error) {
	raw, code, err := apiRequest(apiKey, baseURL, "GET", "/v1/workspaces/default", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		var ae apiError
		_ = json.Unmarshal(raw, &ae)
		msg := ae.Error
		if msg == "" {
			msg = string(raw)
		}
		return nil, fmt.Errorf("API %d: %s", code, msg)
	}
	var w defaultWorkspaceResp
	if err := json.Unmarshal(raw, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func listWorkspaces(apiKey, baseURL string) ([]workspaceResp, error) {
	raw, code, err := apiRequest(apiKey, baseURL, "GET", "/v1/workspaces", nil)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		var ae apiError
		_ = json.Unmarshal(raw, &ae)
		msg := ae.Error
		if msg == "" {
			msg = string(raw)
		}
		return nil, fmt.Errorf("API %d: %s", code, msg)
	}
	var r listWorkspacesResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return r.Items, nil
}

func createCLIKey(apiKey, baseURL, workspaceID, toolID, label, scope, projectID string) (*createCliKeyResp, error) {
	req := createCliKeyReq{ToolID: toolID, Label: label}
	if scope != "" {
		req.Scope = scope
	}
	if projectID != "" {
		req.ProjectID = projectID
	}
	raw, code, err := apiRequest(apiKey, baseURL, "POST", "/v1/workspaces/"+workspaceID+"/cli-keys", req)
	if err != nil {
		return nil, err
	}
	if code != http.StatusCreated && code != http.StatusOK {
		var ae apiError
		_ = json.Unmarshal(raw, &ae)
		msg := ae.Error
		if msg == "" {
			msg = string(raw)
		}
		return nil, fmt.Errorf("API %d: %s", code, msg)
	}
	var r createCliKeyResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
