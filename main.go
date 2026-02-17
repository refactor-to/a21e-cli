// a21e CLI — workspace setup, init, and API access.
package main

import (
	"bufio"
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
  a21e init                              Prompt for API key if needed, use default workspace
  a21e init --tool <tool_id>              Create CLI key for tool (default workspace)
  a21e init --tool <tool_id> --workspace <id>   Scoped to workspace
  a21e init --non-interactive --tool <id> --workspace <id> --yes   CI mode

Environment:
  A21E_API_KEY   Your API key (get one at https://a21e.com/api-key)
  A21E_API_URL   API base URL (default https://api.a21e.com)

Supported tool_id: codex_cli, claude_code_cli, cursor, vscode, jetbrains, openai_cli_custom
`)
}

func runInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	tool := fs.String("tool", "", "Tool ID to configure (e.g. claude_code_cli)")
	workspaceID := fs.String("workspace", "", "Workspace ID (omit to use default)")
	nonInteractive := fs.Bool("non-interactive", false, "CI/non-interactive mode")
	yes := fs.Bool("yes", false, "Skip confirmations")
	_ = yes
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	apiKey := getAPIKey()
	baseURL := getAPIBaseURL()

	// --- Resolve workspace ---
	var wid string
	if *workspaceID != "" {
		wid = *workspaceID
	} else {
		if apiKey == "" {
			if *nonInteractive {
				fmt.Fprintf(os.Stderr, "a21e init: A21E_API_KEY is required in non-interactive mode\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "No A21E_API_KEY set. Get an API key at https://a21e.com/api-key then:\n")
			fmt.Fprintf(os.Stderr, "  export A21E_API_KEY=your_key\n")
			fmt.Fprintf(os.Stderr, "  a21e init --tool <tool_id>\n")
			os.Exit(1)
		}
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

	// --- Tool required for key creation ---
	if *tool == "" {
		if *nonInteractive {
			fmt.Fprintf(os.Stderr, "a21e init: --tool is required in non-interactive mode\n")
			os.Exit(1)
		}
		fmt.Println("To create a CLI key for a tool, run:")
		fmt.Printf("  a21e init --tool <tool_id> [--workspace %s]\n", wid)
		fmt.Println("Supported tool_id: codex_cli, claude_code_cli, cursor, vscode, jetbrains, openai_cli_custom")
		fmt.Println("Or complete setup in the dashboard: https://a21e.com")
		return
	}

	if !isValidToolID(*tool) {
		fmt.Fprintf(os.Stderr, "a21e init: invalid tool_id %q. Supported: %s\n", *tool, strings.Join(validToolIDs, ", "))
		os.Exit(1)
	}

	if apiKey == "" {
		fmt.Fprintf(os.Stderr, "a21e init: A21E_API_KEY is required. Get one at https://a21e.com/api-key\n")
		os.Exit(1)
	}

	// --- Create CLI key ---
	label := suggestLabel(*tool)
	resp, err := createCLIKey(apiKey, baseURL, wid, *tool, label)
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
