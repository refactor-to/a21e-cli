// a21e CLI — workspace setup, init, and API access.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}
	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Println("a21e", version)
	case "init":
		runInit(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `a21e — Agent Performance Layer CLI

Usage:
  a21e version          Show version
  a21e init            Interactive setup (or use --tool and --workspace)

Init:
  a21e init                              Auto-detect tool in Cursor/VS Code/JetBrains terminal, or prompt
  a21e init --tool <tool_id>              Create user-scoped CLI key (works with any workspace)
  a21e init --tool <tool_id> --workspace <id>   Create key in workspace (user-scoped by default)
  a21e init --tool <tool_id> --workspace <id> --workspace-scoped   Key bound to that workspace only
  a21e init --tool <tool_id> --workspace <id> --project <id>   Key bound to that project only
  a21e init --tool <tool_id> --workspace <id> --apply   Auto-apply supported tool settings
  a21e init --non-interactive --tool <id> --workspace <id> --yes   CI mode

Environment:
  A21E_API_KEY   Your API key (get one at https://a21e.com/api-key)
  A21E_API_URL   API base URL (default https://api.a21e.com)
  A21E_TOOL_ID   Override auto-detected tool (e.g. cursor, vscode, jetbrains)

Supported tool_id: codex_cli, claude_code_cli, cursor, vscode, jetbrains, openai_cli_custom
`)
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	tool := fs.String("tool", "", "Tool ID to configure (e.g. claude_code_cli)")
	workspaceID := fs.String("workspace", "", "Workspace ID (omit to use default)")
	workspaceScoped := fs.Bool("workspace-scoped", false, "Bind key to this workspace only")
	projectID := fs.String("project", "", "Bind key to this project (must be in the given workspace)")
	apply := fs.Bool("apply", false, "Auto-apply configuration where supported")
	nonInteractive := fs.Bool("non-interactive", false, "CI/non-interactive mode")
	yes := fs.Bool("yes", false, "Skip confirmations")
	_ = yes
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	apiKey := getAPIKey()
	baseURL := getAPIBaseURL()

	// --- No API key: offer device flow (browser sign-in) or exit ---
	if apiKey == "" {
		if *nonInteractive {
			fmt.Fprintf(os.Stderr, "a21e init: A21E_API_KEY is required in non-interactive mode (or run without --non-interactive to use device login)\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "No API key found. Authorize this device in your browser to get a key.\n")
		key, err := startDeviceFlow(baseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "a21e init: %v\n", err)
			os.Exit(1)
		}
		if err := writeCredentialsFile(key); err != nil {
			fmt.Fprintf(os.Stderr, "a21e init: could not save key to file: %v\n", err)
			fmt.Fprintf(os.Stderr, "Save the key below and set A21E_API_KEY in your environment.\n")
		}
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "You're all set. Your key has been saved to ~/.a21e/credentials.")
		fmt.Fprintln(os.Stderr, "To use it in this shell or add to your profile:")
		fmt.Fprintf(os.Stderr, "  export A21E_API_KEY=%s\n", key)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "To create a tool-specific key (e.g. for Cursor), run: a21e init --tool cursor")
		return
	}

	// --- Resolve workspace ---
	var wid string
	if *workspaceID != "" {
		wid = *workspaceID
	} else {
		ws, err := getDefaultWorkspace(apiKey, baseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "a21e init: %v\n", err)
			os.Exit(1)
		}
		wid = ws.ID
		if *workspaceID == "" && !*nonInteractive && *tool == "" {
			fmt.Printf("Using workspace: %s (%s)\n", ws.Name, wid)
		}
	}

	// --- Tool: explicit flag, then auto-detect from environment, else prompt ---
	if *tool == "" {
		*tool = detectToolFromEnvironment()
		if *tool != "" && !*nonInteractive {
			fmt.Fprintf(os.Stderr, "Detected tool: %s\n", *tool)
		}
	}
	if *tool == "" {
		if *nonInteractive {
			fmt.Fprintf(os.Stderr, "a21e init: --tool is required in non-interactive mode (or set A21E_TOOL_ID)\n")
			os.Exit(1)
		}
		fmt.Println("To create a CLI key for a tool, run:")
		fmt.Printf("  a21e init --tool <tool_id> [--workspace %s]\n", wid)
		fmt.Println("Supported tool_id: codex_cli, claude_code_cli, cursor, vscode, jetbrains, openai_cli_custom")
		fmt.Println("Or run 'a21e init' from inside Cursor, VS Code, or JetBrains terminal to auto-detect.")
		fmt.Println("Or complete setup in the dashboard: https://a21e.com")
		return
	}

	if !isValidToolID(*tool) {
		fmt.Fprintf(os.Stderr, "a21e init: invalid tool_id %q. Supported: %s\n", *tool, strings.Join(validToolIDs, ", "))
		os.Exit(1)
	}

	if *projectID != "" && *workspaceID == "" {
		fmt.Fprintf(os.Stderr, "a21e init: --project requires --workspace\n")
		os.Exit(1)
	}

	// --- Create CLI key ---
	label := suggestLabel(*tool)
	scope := "user"
	if *projectID != "" {
		scope = "project"
	} else if *workspaceScoped {
		scope = "workspace"
	}
	resp, err := createCLIKey(apiKey, baseURL, wid, *tool, label, scope, *projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "a21e init: %v\n", err)
		os.Exit(1)
	}

	// --- Show key once with warning (RFC: "Show once with warning") ---
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Save this key now. You will not be able to view it again.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Println(resp.Key)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Add to your environment (e.g. in ~/.zshrc or ~/.bashrc):")
	fmt.Fprintf(os.Stderr, "  export A21E_API_KEY=%s\n", resp.Key)
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintln(os.Stderr, "Tool configuration values:")
	fmt.Fprintf(os.Stderr, "  Base URL: %s\n", openAIBaseURL(baseURL))
	fmt.Fprintf(os.Stderr, "  API key:  %s\n", resp.Key)
	fmt.Fprintln(os.Stderr, "  Model:    a21e-auto")
	fmt.Fprintln(os.Stderr, "")

	if *apply {
		summary, err := applyToolConfiguration(*tool, resp.Key, baseURL)
		if err == nil {
			fmt.Fprintln(os.Stderr, "Auto-configuration applied:")
			fmt.Fprintf(os.Stderr, "  %s\n", summary.Details)
			fmt.Fprintf(os.Stderr, "  Updated: %s\n", summary.UpdatedPath)
			if summary.BackupPath != "" {
				fmt.Fprintf(os.Stderr, "  Backup:  %s\n", summary.BackupPath)
			}
			fmt.Fprintln(os.Stderr, "")
		} else if errors.Is(err, errAutoConfigUnsupported) {
			fmt.Fprintln(os.Stderr, "Auto-configuration is not supported for this tool yet.")
			fmt.Fprintln(os.Stderr, "Configure your tool manually with the values above.")
			fmt.Fprintln(os.Stderr, "")
		} else {
			fmt.Fprintf(os.Stderr, "Auto-configuration failed: %v\n", err)
			fmt.Fprintln(os.Stderr, "Configure your tool manually with the values above.")
			fmt.Fprintln(os.Stderr, "")
		}
	}

	if !*nonInteractive && isTerminal() {
		fmt.Fprint(os.Stderr, "Press Enter to continue... ")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func suggestLabel(toolID string) string {
	labels := map[string]string{
		"codex_cli":         "Codex CLI API key",
		"claude_code_cli":   "Claude Code API key",
		"cursor":            "Cursor API key",
		"vscode":            "VS Code API key",
		"jetbrains":         "JetBrains API key",
		"openai_cli_custom": "OpenAI-compatible CLI API key",
	}
	if l, ok := labels[toolID]; ok {
		return l
	}
	return "CLI API key"
}

func isTerminal() bool {
	f, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (f.Mode() & os.ModeCharDevice) != 0
}
