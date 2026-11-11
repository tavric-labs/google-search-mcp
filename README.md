# google-search-mcp

A **generic**, token-efficient MCP server for searching **anything** on the web via
**Gemini Google Search grounding** (Google AI Studio). No Custom Search Engine
(`cx`) and no `GOOGLE_CSE_ID`.

Public Runtime URL:

```text
https://mcp.mcpruntime.org/google-search/mcp
```

No platform login / OAuth. Clients authenticate with header **`X-Gemini-Api-Key`**
(your Gemini AI Studio key). The gateway is enabled and **audits tool calls** for
usage metrics.

Agents are steered to **craft → search → open 0–2 pages → pack → stop**.

## Install in your client

One-click install badges (same HTTPS deeplink pattern used by Perplexity, AWS Labs,
and other public MCP READMEs — GitHub strips `cursor://` / `vscode:` schemes, so
badges must use `https://cursor.com/...` and `https://vscode.dev/...`):

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-light.svg)](https://cursor.com/en/install-mcp?name=google-search&config=eyJ1cmwiOiJodHRwczovL21jcC5tY3BydW50aW1lLm9yZy9nb29nbGUtc2VhcmNoL21jcCIsImhlYWRlcnMiOnsiWC1HZW1pbmktQXBpLUtleSI6IiR7ZW52OkdFTUlOSV9BUElfS0VZfSJ9fQ%3D%3D)
[![Install in VS Code](https://img.shields.io/badge/VS_Code-Install-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://vscode.dev/redirect/mcp/install?name=google-search&config=%7B%22type%22%3A%22http%22%2C%22url%22%3A%22https%3A%2F%2Fmcp.mcpruntime.org%2Fgoogle-search%2Fmcp%22%2C%22headers%22%3A%7B%22X-Gemini-Api-Key%22%3A%22%24%7Binput%3Agemini-api-key%7D%22%7D%7D&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22gemini-api-key%22%2C%22description%22%3A%22Gemini%20AI%20Studio%20API%20Key%22%2C%22password%22%3Atrue%7D%5D)
[![Install in VS Code Insiders](https://img.shields.io/badge/VS_Code_Insiders-Install-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=google-search&config=%7B%22type%22%3A%22http%22%2C%22url%22%3A%22https%3A%2F%2Fmcp.mcpruntime.org%2Fgoogle-search%2Fmcp%22%2C%22headers%22%3A%7B%22X-Gemini-Api-Key%22%3A%22%24%7Binput%3Agemini-api-key%7D%22%7D%7D&inputs=%5B%7B%22type%22%3A%22promptString%22%2C%22id%22%3A%22gemini-api-key%22%2C%22description%22%3A%22Gemini%20AI%20Studio%20API%20Key%22%2C%22password%22%3Atrue%7D%5D&quality=insiders)
[![Install in Visual Studio](https://img.shields.io/badge/Visual_Studio-Install-C16FDE?style=flat-square&logo=visualstudio&logoColor=white)](https://vs-open.link/mcp-install?%7B%22name%22%3A%22google-search%22%2C%22type%22%3A%22http%22%2C%22url%22%3A%22https%3A%2F%2Fmcp.mcpruntime.org%2Fgoogle-search%2Fmcp%22%2C%22headers%22%3A%7B%22X-Gemini-Api-Key%22%3A%22%24%7Binput%3Agemini-api-key%7D%22%7D%7D)
[![Add to Claude](https://img.shields.io/badge/Add_to-Claude-D97757?style=flat-square&logo=anthropic&logoColor=white)](https://claude.ai/customize/connectors?modal=add-custom-connector&connectorName=Google%20Search&connectorUrl=https%3A%2F%2Fmcp.mcpruntime.org%2Fgoogle-search%2Fmcp)

Grab a key first: [Google AI Studio → API keys](https://aistudio.google.com/apikey).
Export it locally (never commit it):

```bash
export GEMINI_API_KEY='...'
```

Install links never embed the secret. Cursor resolves `${env:GEMINI_API_KEY}`;
VS Code / Visual Studio prompt for the key via `${input:gemini-api-key}`.
Claude.ai connectors are URL-only — if that UI cannot set `X-Gemini-Api-Key`,
use Claude Code / Cursor / VS Code / Codex instead.

### Cursor (`~/.cursor/mcp.json`)

```json
{
  "mcpServers": {
    "google-search": {
      "url": "https://mcp.mcpruntime.org/google-search/mcp",
      "headers": {
        "X-Gemini-Api-Key": "${env:GEMINI_API_KEY}"
      }
    }
  }
}
```

### VS Code (user `mcp.json` or `.vscode/mcp.json`)

```json
{
  "inputs": [
    {
      "id": "gemini-api-key",
      "type": "promptString",
      "password": true,
      "description": "Gemini AI Studio API key"
    }
  ],
  "servers": {
    "google-search": {
      "type": "http",
      "url": "https://mcp.mcpruntime.org/google-search/mcp",
      "headers": {
        "X-Gemini-Api-Key": "${input:gemini-api-key}"
      }
    }
  }
}
```

Or from a terminal:

```bash
code --add-mcp '{"name":"google-search","type":"http","url":"https://mcp.mcpruntime.org/google-search/mcp","headers":{"X-Gemini-Api-Key":"${input:gemini-api-key}"}}'
```

### Claude Code

```bash
claude mcp add --transport http google-search \
  https://mcp.mcpruntime.org/google-search/mcp \
  --header "X-Gemini-Api-Key: ${GEMINI_API_KEY}"
```

### Claude.ai (custom connector)

Use **Add to Claude** above, or **Customize → Connectors → Add custom connector**
with URL `https://mcp.mcpruntime.org/google-search/mcp`. Claude.ai custom
connectors are OAuth-oriented; if the UI cannot attach a custom header, use
Claude Code / Cursor / VS Code with the JSON header instead.

### Codex (`~/.codex/config.toml`)

```toml
[mcp_servers.google-search]
url = "https://mcp.mcpruntime.org/google-search/mcp"

[mcp_servers.google-search.env_http_headers]
X-Gemini-Api-Key = "GEMINI_API_KEY"
```

`GEMINI_API_KEY` is the **name** of the environment variable (not the secret).

```bash
codex mcp add google-search --url https://mcp.mcpruntime.org/google-search/mcp
# then add the env_http_headers block above (CLI has no --header for custom names yet)
```

### OpenAI / ChatGPT

In ChatGPT, add a custom MCP / connector with URL
`https://mcp.mcpruntime.org/google-search/mcp` if your workspace supports remote
MCP with custom headers. If the product only supports OAuth connectors, use
Cursor, VS Code, Claude Code, or Codex with the header config above.

### Windsurf

`~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "google-search": {
      "serverUrl": "https://mcp.mcpruntime.org/google-search/mcp",
      "headers": {
        "X-Gemini-Api-Key": "${env:GEMINI_API_KEY}"
      }
    }
  }
}
```

### Gemini CLI

`~/.gemini/settings.json`:

```json
{
  "mcpServers": {
    "google-search": {
      "httpUrl": "https://mcp.mcpruntime.org/google-search/mcp",
      "headers": {
        "X-Gemini-Api-Key": "${GEMINI_API_KEY}"
      }
    }
  }
}
```

### MCP Runtime console

On [platform.mcpruntime.org](https://platform.mcpruntime.org), open the public
**google-search** server → **Install in your client** for Cursor, Codex, Claude
Code, and VS Code (when signed in). For this header-mode server, still set
`X-Gemini-Api-Key` in the generated client config.

## Get a Gemini API key

1. Open [Google AI Studio → API keys](https://aistudio.google.com/apikey).
2. Sign in with the Google account that will own the quota.
3. Click **Create API key** (create or pick a Google Cloud project when asked).
4. Copy the key once. Store it in a password manager or `export GEMINI_API_KEY=...`.
   Do **not** paste keys into chat, issues, PRs, or git.
5. Optional: check **Usage / rate limits** in AI Studio for live RPM / TPM / RPD.

Local smoke env file (gitignored):

```bash
echo 'GEMINI_API_KEY=...' > .env
chmod 600 .env
```

## How auth works on the public deploy

| Piece | Value |
|---|---|
| Scope | `public` (`mcp-servers-public`) |
| Auth mode | `header` (CRD `spec.auth.mode`) |
| Platform login | Not required |
| Client credential | Header `X-Gemini-Api-Key` (required; no header → `401 missing_credential`) |
| Gateway | `enabled: true` (audits allow/deny tool use) |
| Policy | `allow-list` + delegated allow rules for the five tools |

Without the header the gateway returns `401 missing_credential`. With a valid
Gemini key, search tools call Gemini Search grounding.

Operators may also mount `GEMINI_API_KEY` on the pod as a fallback for local
smoke; public clients still send the header.

## Free quota (Search grounding)

Official Gemini Developer API pricing (check
[pricing](https://ai.google.dev/gemini-api/docs/pricing) — numbers change):

| Plan | Google Search grounding |
|---|---|
| **Free tier** | **500 requests/day (RPD)** shared across Gemini **Flash** and **Flash-Lite**. Not available for **Pro** on free tier. |
| **Paid (Gemini 2.5 Flash / Flash-Lite)** | 1,500 RPD free (shared), then paid per grounded prompt. |
| **Paid (Gemini 3+)** | Separate monthly free search allowance, then paid; default daily caps still apply. |

Also remember:

- Limits are **per Google Cloud project**, not per API key string.
- Daily quotas reset at **midnight Pacific Time**.
- This MCP defaults to `gemini-2.5-flash`, which is the free-tier Search path.
- Each `google_search` call ≈ **1** grounded Gemini request.
- Each `google_multi_search` can use **up to 3** requests (one per query).
- `google_craft_query`, `google_open`, and `google_pack_evidence` do **not** use
  Search grounding (`google_open` only fetches the page URL).

## Tools

| Tool | Calls Gemini Search? | Purpose |
|---|---|---|
| `google_craft_query` | No | Design 1–3 engine-ready queries from any plain-language goal. |
| `google_search` | Yes | Slim hits (`title`, `url`, `snippet`, `hit_id`) from grounding citations. |
| `google_multi_search` | Yes (≤3) | Up to 3 queries in parallel; merge/dedupe. |
| `google_open` | No (HTTP fetch) | Short extract from one prior hit / URL (budgeted). |
| `google_pack_evidence` | No | Citation pack from selected `hit_id`s. |

## Configuration (operators)

| Variable | Purpose |
|---|---|
| `GEMINI_API_KEY` | Optional platform-held key (pod Secret). |
| `GEMINI_MODEL` | Defaults to `gemini-2.5-flash`. |
| `GEMINI_API_BASE` | Defaults to `https://generativelanguage.googleapis.com/v1beta`. |
| `GOOGLE_SEARCH_MAX_RESULTS` | Default hit count (capped at 10). |
| `GOOGLE_SEARCH_MAX_PARALLEL` | Max queries in `google_multi_search`. |
| `GOOGLE_SEARCH_MAX_OPEN_CHARS` | Char budget for `google_open`. |
| `GOOGLE_SEARCH_MAX_PACK_CHARS` | Char budget for evidence packs. |
| `PORT`, `MCP_PATH` | Defaults `8080` and `/mcp`. |

See `.mcp/servers.yaml` and `manifests/google-search.yaml` for the public deploy.

## Develop

```bash
gofmt -l . && go vet ./... && go test -race ./...
```

## Release

Push a `vX.Y.Z` tag on `main`. The Release workflow publishes
`ghcr.io/tavric-labs/google-search-mcp` (multi-arch) and creates the GitHub
release.
