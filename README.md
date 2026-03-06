# a21e CLI

Minimal CLI for workspace setup and tool configuration. Install via:

```bash
curl -fsSL https://get.a21e.com/install.sh | bash
```

## Commands

- **a21e version** — Show version (set at build time from release tag).
- **a21e init** — Browser auth if needed, then create a tool-specific CLI key automatically.
  - **Two-command setup (recommended):**
    1. `curl -fsSL https://get.a21e.com/install.sh | bash`
    2. `a21e init --tool <tool_id> --workspace <workspace_id> --apply --yes`
  - **Device login (no key yet):** If no key exists, `a21e init` prints a browser URL. After you authorize, setup continues in the same command and creates the tool key automatically.
  - **Key hygiene:** The temporary bootstrap key used during browser auth is revoked automatically after the tool key is created.
  - **No manual key export required:** Generated keys are saved to `~/.a21e/credentials` and reused automatically.
  - **Single command in IDEs:** Run `a21e init` from Cursor, VS Code, or JetBrains; the CLI auto-detects the tool so you don’t need `--tool`.
  - Interactive: `a21e init` (uses default workspace; tool is auto-detected when possible, else use `--tool <id>`).
  - Scoped: `a21e init --tool claude_code_cli [--workspace <id>]`.
  - Auto-apply supported configs: `a21e init --tool vscode --workspace <id> --apply --yes`.
  - CI: `a21e init --non-interactive --tool <id> --workspace <id> --yes` (or set `A21E_TOOL_ID`).
  - Key source: `~/.a21e/credentials` (default) or `A21E_API_KEY` (optional override). Optional `A21E_API_URL`, `A21E_TOOL_ID`.
  - Supported tool IDs: `codex_cli`, `claude_code_cli`, `cursor`, `vscode`, `jetbrains`, `openai_cli_custom`.
  - **Note:** Cursor’s integrated terminal often sets `TERM_PROGRAM=vscode`, so tool may be detected as `vscode`. To create a Cursor key and apply Cursor settings, use `a21e init --tool cursor` or `A21E_TOOL_ID=cursor a21e init`.

## Auto-apply support

- `vscode`: patches VS Code user settings (`a21e.apiUrl`, `a21e.apiKey`, `a21e.defaultModel`).
- `cursor`: patches Cursor user settings (`a21e.apiUrl`, `a21e.apiKey`, `a21e.defaultModel`).
- `openai_cli_custom`: updates shell profile with `OPENAI_API_BASE`, `OPENAI_BASE_URL`, `OPENAI_API_KEY`.
- `codex_cli`, `claude_code_cli`, `jetbrains`: manual configuration remains required.

## Building

From repo root or `packages/cli`:

```bash
cd packages/cli
go build -ldflags "-X main.version=dev" -o a21e .
```

## Releases

Binaries are built and attached to GitHub Releases by the [Release CLI](../../.github/workflows/release-cli.yml) workflow when you **publish a Release** (e.g. from tag `v0.1.0`). Assets: `a21e-darwin-arm64.tar.gz`, `a21e-darwin-x86_64.tar.gz`, `a21e-linux-arm64.tar.gz`, `a21e-linux-x86_64.tar.gz` and their `.sha256` files. The install script uses `https://github.com/refactor-to/a21e-cli/releases/latest/download/<asset>` (default repo: refactor-to/a21e-cli).
