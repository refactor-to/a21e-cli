# a21e CLI

The a21e CLI connects your coding tools to the [a21e](https://a21e.com) Agent Performance Layer. It handles authentication, creates tool-specific API keys, and auto-configures supported editors — so you can start using a21e prompts in two commands.

## Quick start

```bash
# 1. Install (macOS or Linux)
curl -fsSL https://get.a21e.com/install.sh | bash

# 2. Set up your tool (opens browser to authenticate)
a21e init
```

The CLI auto-detects your editor when run from its integrated terminal, authenticates via your browser, creates an API key, and saves it to `~/.a21e/credentials`. Verify with `a21e version`.

## Usage

### Interactive setup (recommended)

Run from your editor's integrated terminal:

```bash
a21e init
```

The CLI will:
1. Detect your editor (Cursor, VS Code, JetBrains)
2. Open your browser to sign in (if no API key exists yet)
3. Create a tool-specific API key
4. Save credentials to `~/.a21e/credentials`

### Specify a tool explicitly

```bash
a21e init --tool cursor
a21e init --tool claude_code_cli --workspace <workspace_id>
```

### Auto-apply editor settings

For supported tools, `--apply` patches your editor config automatically:

```bash
a21e init --tool vscode --apply --yes
a21e init --tool cursor --apply --yes
```

### CI / non-interactive mode

```bash
export A21E_API_KEY=a21e_...
a21e init --non-interactive --tool codex_cli --workspace <workspace_id> --yes
```

### Key scoping

By default, keys are user-scoped (work across all workspaces). You can restrict a key to a single workspace:

```bash
a21e init --tool cursor --workspace <id> --workspace-scoped
```

## Supported tools

| Tool ID | Editor | Auto-detect | Auto-apply |
|---------|--------|:-----------:|:----------:|
| `cursor` | Cursor | Yes | Yes — patches Cursor user settings |
| `vscode` | VS Code | Yes | Yes — patches VS Code user settings |
| `jetbrains` | IntelliJ, PyCharm, etc. | Yes | No — configure manually |
| `claude_code_cli` | Claude Code | No | No — configure manually |
| `codex_cli` | Codex CLI | No | No — configure manually |
| `openai_cli_custom` | OpenAI-compatible CLIs | No | Yes — sets shell env vars |

**Auto-detect** means the CLI identifies the tool when run from its integrated terminal.
**Auto-apply** means `--apply` can write the configuration for you.

> **Cursor users:** Cursor's terminal sometimes reports itself as `vscode`. If this happens, specify the tool explicitly: `a21e init --tool cursor`.

## Configuration

### Credentials file

The CLI stores your API key at:

```
~/.a21e/credentials
```

Format:

```
A21E_API_KEY=a21e_...
```

This file is created automatically during `a21e init`. You never need to edit it manually.

### Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `A21E_API_KEY` | API key (overrides credentials file) | Read from `~/.a21e/credentials` |
| `A21E_API_URL` | API base URL | `https://api.a21e.com` |
| `A21E_TOOL_ID` | Override auto-detected tool ID | Auto-detected from terminal |

### What auto-apply configures

| Tool | What gets patched | Settings |
|------|-------------------|----------|
| VS Code | `~/Library/Application Support/Code/User/settings.json` | `a21e.apiUrl`, `a21e.apiKey`, `a21e.defaultModel` |
| Cursor | `~/Library/Application Support/Cursor/User/settings.json` | `a21e.apiUrl`, `a21e.apiKey`, `a21e.defaultModel` |
| OpenAI CLI | Shell profile (`.zshrc`, `.bashrc`, etc.) | `OPENAI_API_BASE`, `OPENAI_BASE_URL`, `OPENAI_API_KEY` |

On Linux, editor settings are at `~/.config/{Code,Cursor}/User/settings.json`.

A backup of your original file is created before any changes (e.g., `settings.json.bak-20260305T120000Z`).

## Updating

Re-run the install script:

```bash
curl -fsSL https://get.a21e.com/install.sh | bash
```

Your credentials are preserved — only the binary is replaced.

## Uninstalling

```bash
# Remove the binary
rm "$(which a21e)"

# Remove credentials and config
rm -rf ~/.a21e
```

## Troubleshooting

**"No API key found" when running `a21e init`:**
This is expected on first run. The CLI will open your browser to authenticate. Complete the sign-in flow and the key is saved automatically.

**Tool detected as `vscode` when using Cursor:**
Cursor's integrated terminal sets `TERM_PROGRAM=vscode`. Use `a21e init --tool cursor` or set `A21E_TOOL_ID=cursor` in your environment.

**"Permission denied" during install:**
The install script places the binary in `/usr/local/bin`. If that fails, run with `sudo` or install to a user directory:
```bash
curl -fsSL https://get.a21e.com/install.sh | INSTALL_DIR=~/.local/bin bash
```

**Auto-apply failed with "settings JSON is invalid":**
Your editor's `settings.json` has a syntax error. Fix it manually (or restore from the `.bak-*` backup the CLI created), then re-run `a21e init --apply`.

**Key not working after init:**
Check that `~/.a21e/credentials` contains a valid key starting with `a21e_`. If you've set `A21E_API_KEY` in your shell profile, it takes precedence over the credentials file.

## License

MIT
