# a21e CLI

Minimal CLI for workspace setup and tool configuration. Install via:

```bash
curl -fsSL https://get.a21e.com/install.sh | bash
```

## Commands

- **a21e version** — Show version (set at build time from release tag).
- **a21e init** — Create a CLI key for a tool and print export snippet.
  - **Device login (no key yet):** Run `a21e init` with no `A21E_API_KEY` set. The CLI will print a URL; open it in your browser, sign in, and click “Authorize this device.” The key is then saved to `~/.a21e/credentials` and used automatically on the next run.
  - **Single command in IDEs:** With a key (from device login or env), run `a21e init` from Cursor, VS Code, or JetBrains; the CLI auto-detects the tool so you don’t need `--tool`.
  - Interactive: `a21e init` (uses default workspace; tool is auto-detected when possible, else use `--tool <id>`).
  - Scoped: `a21e init --tool claude_code_cli [--workspace <id>]`.
  - Auto-apply supported configs: `a21e init --tool vscode --workspace <id> --apply`.
  - CI: `a21e init --non-interactive --tool <id> --workspace <id> --yes` (or set `A21E_TOOL_ID`).
  - Key source: `A21E_API_KEY` env, or `~/.a21e/credentials` (written by device login). Optional `A21E_API_URL`, `A21E_TOOL_ID`.
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
