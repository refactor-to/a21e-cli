// detect.go — Detect host tool/IDE from environment so "a21e init" can work without --tool.
//
// When run inside an IDE's integrated terminal, many IDEs set environment variables
// that identify the host. We use them to auto-select the tool_id so a single
// "a21e init" command works everywhere.
//
// Detection order:
//   - A21E_TOOL_ID: explicit override (CI or user)
//   - TERM_PROGRAM=cursor → cursor (Cursor may set this in future; currently Cursor often sets vscode)
//   - TERM_PROGRAM=vscode → vscode (VS Code and sometimes Cursor)
//   - TERMINAL_EMULATOR containing "JetBrains" → jetbrains (IntelliJ, PyCharm, etc.)
//
// codex_cli, claude_code_cli, openai_cli_custom have no standard terminal env;
// use --tool or A21E_TOOL_ID for those.

package main

import (
	"os"
	"strings"
)

func detectToolFromEnvironment() string {
	if v := os.Getenv("A21E_TOOL_ID"); v != "" {
		v = strings.TrimSpace(strings.ToLower(v))
		if isValidToolID(v) {
			return v
		}
	}
	switch os.Getenv("TERM_PROGRAM") {
	case "cursor":
		return "cursor"
	case "vscode":
		return "vscode"
	}
	if strings.Contains(os.Getenv("TERMINAL_EMULATOR"), "JetBrains") {
		return "jetbrains"
	}
	return ""
}
