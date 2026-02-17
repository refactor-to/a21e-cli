# a21e CLI

`a21e` is the command-line client for workspace setup and tool-specific key provisioning.

## Install

```bash
curl -fsSL https://get.a21e.com/install.sh | bash
```

Verify:

```bash
a21e version
```

## Quick Start

1. Create a platform API key at <https://a21e.com/api-key>.
2. Export it:
   ```bash
   export A21E_API_KEY="a21e_your_platform_key"
   ```
3. Generate a tool key:
   ```bash
   a21e init --tool codex_cli
   ```

## Commands

```bash
a21e version
a21e init
a21e init --tool <tool_id>
a21e init --tool <tool_id> --workspace <workspace_id>
a21e init --non-interactive --tool <tool_id> --workspace <workspace_id> --yes
```

## Supported tool_id values

- `codex_cli`
- `claude_code_cli`
- `cursor`
- `vscode`
- `jetbrains`
- `openai_cli_custom`

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `A21E_API_KEY` | Yes (for `init`) | — | Your a21e platform API key |
| `A21E_API_URL` | No | `https://api.a21e.com` | Override API base URL |

## Tool Setup

### Shared OpenAI-compatible values

- Base URL: `https://api.a21e.com/v1`
- API key: key returned by `a21e init`
- Model: `a21e-auto`

### Codex CLI (`codex_cli`)

```bash
a21e init --tool codex_cli
```

### Claude Code (`claude_code_cli`)

```bash
a21e init --tool claude_code_cli
```

### Cursor (`cursor`)

```bash
a21e init --tool cursor
```

### VS Code (`vscode`)

```bash
a21e init --tool vscode
```

### JetBrains (`jetbrains`)

```bash
a21e init --tool jetbrains
```

### OpenAI-compatible CLI (`openai_cli_custom`)

```bash
a21e init --tool openai_cli_custom
export OPENAI_API_BASE="https://api.a21e.com/v1"
export OPENAI_API_KEY="a21e_your_generated_tool_key"

openai api chat.completions.create \
  -m a21e-auto \
  -g user "Hello from the CLI"
```

## Workspace-Scoped Setup

```bash
a21e init --tool claude_code_cli --workspace <workspace_id>
```

## CI / Non-Interactive Setup

```bash
export A21E_API_KEY="a21e_your_platform_key"
a21e init --non-interactive --tool openai_cli_custom --workspace <workspace_id> --yes
```

## Build from Source

```bash
cd packages/cli
go build -ldflags "-X main.version=dev" -o a21e .
./a21e version
```

## Release Process

1. Create and push a tag:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
2. Publish release:
   ```bash
   gh release create v0.1.0 --title "v0.1.0" --notes "Initial CLI release"
   ```
3. The release workflow uploads:
   - `a21e-darwin-arm64.tar.gz`
   - `a21e-darwin-x86_64.tar.gz`
   - `a21e-linux-arm64.tar.gz`
   - `a21e-linux-x86_64.tar.gz`
   - matching `.sha256` files
4. Verify asset availability:
   ```bash
   curl -I https://github.com/refactor-to/a21e/releases/latest/download/a21e-darwin-arm64.tar.gz
   ```

## Troubleshooting

- `a21e init: A21E_API_KEY is required`
  - Export `A21E_API_KEY` first.
- `invalid tool_id`
  - Use one of the supported values listed above.
- Installer download `404`
  - No published release, missing asset, or private-only asset URL.
- API `401`/`403`
  - Wrong key, wrong base URL, or revoked key.

## Security Notes

- Generated tool keys are shown once.
- Treat keys as secrets.
- Rotate immediately if exposed.

## License

See the repository `LICENSE`.
