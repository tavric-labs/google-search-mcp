# google-search-mcp

A **generic**, token-efficient MCP server for searching **anything** on the web via
**Gemini Google Search grounding** (Google AI Studio). No Custom Search Engine
(`cx`) and no `GOOGLE_CSE_ID` — users only provide a Gemini API key.

Agents are steered to use **short targeted queries**, read **slim SERPs**, open
**few pages**, and answer from a **small citation pack** (craft → search →
selective open → pack → stop).

## Auth (header)

Create a key at [Google AI Studio](https://aistudio.google.com/apikey), then send:

```http
X-Gemini-Api-Key: <gemini-ai-studio-api-key>
```

Search grounding must be available on the key’s plan (not all free-tier quotas
include it). On MCP Runtime, `.mcp/servers.yaml` sets `auth.mode: header` with
that header name so connect configs / the credential gateway collect it.

## Tools

| Tool | Calls Google? | Purpose |
|---|---|---|
| `google_craft_query` | No | Design 1–3 engine-ready queries from any plain-language goal. |
| `google_search` | Yes (Gemini + Search grounding) | Slim hits (`title`, `url`, `snippet`, `hit_id`) from grounding citations. |
| `google_multi_search` | Yes | Up to 3 queries in parallel; merge/dedupe. |
| `google_open` | No (HTTP fetch) | Short extract from one prior hit / URL (budgeted). |
| `google_pack_evidence` | No | Citation pack from selected `hit_id`s. |

Default flow: **craft → search → open 0–2 hits → pack → answer**.

## Configuration (operators)

| Variable | Purpose |
|---|---|
| `GEMINI_MODEL` | Defaults to `gemini-2.5-flash`. |
| `GEMINI_API_BASE` | Defaults to `https://generativelanguage.googleapis.com/v1beta`. |
| `GOOGLE_SEARCH_MAX_RESULTS` | Default hit count (capped at 10). |
| `GOOGLE_SEARCH_MAX_PARALLEL` | Max queries in `google_multi_search`. |
| `GOOGLE_SEARCH_MAX_OPEN_CHARS` | Char budget for `google_open`. |
| `GOOGLE_SEARCH_MAX_PACK_CHARS` | Char budget for evidence packs. |
| `PORT`, `MCP_PATH` | Defaults `8080` and `/mcp`. |

`GET <MCP_PATH>/auth/check` returns 204 when `X-Gemini-Api-Key` is present.

## Develop

```bash
gofmt -l . && go vet ./... && go test -race ./...
```

## Release

Push a `vX.Y.Z` tag on `main`. The Release workflow publishes
`ghcr.io/tavric-labs/google-search-mcp` (multi-arch) and creates the GitHub
release.
