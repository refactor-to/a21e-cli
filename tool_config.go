package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var errAutoConfigUnsupported = errors.New("auto configuration is not supported for this tool")

type applySummary struct {
	UpdatedPath string
	BackupPath  string
	Details     string
}

func openAIBaseURL(apiBaseURL string) string {
	trimmed := strings.TrimSpace(strings.TrimRight(apiBaseURL, "/"))
	if trimmed == "" {
		return "https://api.a21e.com/v1"
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed
	}
	return trimmed + "/v1"
}

func applyToolConfiguration(toolID, toolKey, apiBaseURL string) (*applySummary, error) {
	switch toolID {
	case "vscode":
		path, backup, err := upsertEditorSettings("Code", toolKey, apiBaseURL)
		if err != nil {
			return nil, err
		}
		return &applySummary{
			UpdatedPath: path,
			BackupPath:  backup,
			Details:     "Updated VS Code user settings for the a21e extension.",
		}, nil
	case "cursor":
		path, backup, err := upsertEditorSettings("Cursor", toolKey, apiBaseURL)
		if err != nil {
			return nil, err
		}
		return &applySummary{
			UpdatedPath: path,
			BackupPath:  backup,
			Details:     "Updated Cursor user settings for the a21e extension.",
		}, nil
	case "openai_cli_custom":
		path, backup, err := upsertShellEnvBlock(toolID, toolKey, apiBaseURL)
		if err != nil {
			return nil, err
		}
		return &applySummary{
			UpdatedPath: path,
			BackupPath:  backup,
			Details:     "Updated shell profile with OPENAI-compatible a21e environment variables.",
		}, nil
	case "codex_cli", "claude_code_cli", "jetbrains":
		return nil, errAutoConfigUnsupported
	default:
		return nil, errAutoConfigUnsupported
	}
}

func upsertEditorSettings(appName, toolKey, apiBaseURL string) (string, string, error) {
	settingsPath, err := resolveEditorSettingsPath(appName)
	if err != nil {
		return "", "", err
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return "", "", fmt.Errorf("could not prepare editor settings directory: %w", err)
	}

	existing, err := os.ReadFile(settingsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("could not read editor settings file: %w", err)
	}

	updated, changed, err := mergeA21ESettings(existing, toolKey, apiBaseURL)
	if err != nil {
		return "", "", err
	}
	if !changed {
		return settingsPath, "", nil
	}

	backup, err := writeFileWithBackup(settingsPath, existing, updated, 0o600)
	if err != nil {
		return "", "", err
	}
	return settingsPath, backup, nil
}

func mergeA21ESettings(existing []byte, toolKey, apiBaseURL string) ([]byte, bool, error) {
	settings := map[string]any{}
	if len(strings.TrimSpace(string(existing))) > 0 {
		if err := json.Unmarshal(existing, &settings); err != nil {
			return nil, false, fmt.Errorf(
				"settings JSON is invalid. Back up and fix it, then rerun a21e init --apply: %w",
				err,
			)
		}
	}

	changed := false
	changed = setSetting(settings, "a21e.apiUrl", strings.TrimSuffix(openAIBaseURL(apiBaseURL), "/v1")) || changed
	changed = setSetting(settings, "a21e.apiKey", toolKey) || changed
	changed = setSetting(settings, "a21e.defaultModel", "a21e-auto") || changed

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil, false, fmt.Errorf("could not serialize editor settings: %w", err)
	}
	out = append(out, '\n')
	return out, changed, nil
}

func setSetting(target map[string]any, key, value string) bool {
	current, ok := target[key]
	if ok {
		if currentString, ok := current.(string); ok && currentString == value {
			return false
		}
	}
	target[key] = value
	return true
}

func resolveEditorSettingsPath(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName, "User", "settings.json"), nil
	case "linux":
		return filepath.Join(home, ".config", appName, "User", "settings.json"), nil
	default:
		return "", fmt.Errorf("automatic settings patching is not supported on %s", runtime.GOOS)
	}
}

func upsertShellEnvBlock(toolID, toolKey, apiBaseURL string) (string, string, error) {
	rcPath, err := resolveShellRCPath()
	if err != nil {
		return "", "", err
	}

	existing, err := os.ReadFile(rcPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("could not read shell profile: %w", err)
	}

	blockStart := fmt.Sprintf("# >>> a21e %s >>>", toolID)
	blockEnd := fmt.Sprintf("# <<< a21e %s <<<", toolID)
	openAIURL := openAIBaseURL(apiBaseURL)
	block := strings.Join([]string{
		blockStart,
		fmt.Sprintf("export OPENAI_API_BASE=%q", openAIURL),
		fmt.Sprintf("export OPENAI_BASE_URL=%q", openAIURL),
		fmt.Sprintf("export OPENAI_API_KEY=%q", toolKey),
		"export A21E_MODEL=\"a21e-auto\"",
		blockEnd,
	}, "\n")

	updated, changed, err := upsertManagedBlock(string(existing), blockStart, blockEnd, block)
	if err != nil {
		return "", "", err
	}
	if !changed {
		return rcPath, "", nil
	}

	backup, err := writeFileWithBackup(rcPath, existing, []byte(updated), 0o600)
	if err != nil {
		return "", "", err
	}
	return rcPath, backup, nil
}

func resolveShellRCPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve home directory: %w", err)
	}

	switch filepath.Base(os.Getenv("SHELL")) {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".bash_profile"), nil
		}
		return filepath.Join(home, ".bashrc"), nil
	default:
		return filepath.Join(home, ".profile"), nil
	}
}

func upsertManagedBlock(content, startMarker, endMarker, block string) (string, bool, error) {
	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)
	if start >= 0 && end < 0 {
		return "", false, errors.New("found start marker without end marker in existing profile")
	}
	if start < 0 && end >= 0 {
		return "", false, errors.New("found end marker without start marker in existing profile")
	}

	if start >= 0 && end >= 0 {
		end += len(endMarker)
		replaced := content[:start] + block + content[end:]
		return replaced, replaced != content, nil
	}

	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return block + "\n", true, nil
	}
	return trimmed + "\n\n" + block + "\n", true, nil
}

func writeFileWithBackup(path string, oldBytes, newBytes []byte, perm os.FileMode) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("could not create directory for %s: %w", path, err)
	}

	backupPath := ""
	if len(oldBytes) > 0 {
		backupPath = fmt.Sprintf("%s.bak-%s", path, time.Now().UTC().Format("20060102T150405Z"))
		if err := os.WriteFile(backupPath, oldBytes, 0o600); err != nil {
			return "", fmt.Errorf("could not create backup %s: %w", backupPath, err)
		}
	}

	if err := os.WriteFile(path, newBytes, perm); err != nil {
		return "", fmt.Errorf("could not write %s: %w", path, err)
	}
	return backupPath, nil
}
