# google-search-mcp

A **generic**, token-efficient MCP server for searching **anything** on the web via
**Google Custom Search** (Programmable Search JSON API): news, products, docs,
how-tos, people, places, code, or research — same tools for all of it.

Agents are steered to use **short targeted queries**, read **slim SERPs**, open
**few pages**, and answer from a **small citation pack** (craft → search →
selective open → pack → stop).

## Auth (header)

Users only need a **Google API key** with the Custom Search API enabled. Send it
as:

```http
X-Google-Api-Key: <your-google-api-key>
```

The platform owns the Programmable Search Engine `cx` (`GOOGLE_CSE_ID` on the
Deployment). Clients never configure a CSE id.

On MCP Runtime, `.mcp/servers.yaml` sets `auth.mode: header` with that header
name so connect configs and the optional credential gateway sign-in form collect
it. Grab a key from Google Cloud → APIs & Services → Credentials, enable
**Custom Search API**, then paste the key into the header / form.

## Tools

| Tool | Calls Google? | Purpose |
|---|---|---|
| `google_craft_query` | No | Design 1–3 engine-ready queries from any plain-language goal. |
| `google_search` | Yes | One Custom Search; slim hits (`title`, `url`, `snippet`, `hit_id`). |
| `google_multi_search` | Yes | Up to 3 queries in parallel; merge/dedupe. |
| `google_open` | No (HTTP fetch) | Short extract from one prior hit / URL (budgeted). |
| `google_pack_evidence` | No | Citation pack from selected `hit_id`s. |

Default flow: **craft → search → open 0–2 hits → pack → answer**.

## Configuration (operators)

| Variable | Purpose |
|---|---|
| `GOOGLE_CSE_ID` | Required. Platform Programmable Search Engine `cx`. |
| `GOOGLE_CSE_API_BASE` | Defaults to `https://www.googleapis.com/customsearch/v1`. |
| `GOOGLE_SEARCH_MAX_RESULTS` | Default hit count (capped at 10). |
| `GOOGLE_SEARCH_MAX_PARALLEL` | Max queries in `google_multi_search`. |
| `GOOGLE_SEARCH_MAX_OPEN_CHARS` | Char budget for `google_open`. |
| `GOOGLE_SEARCH_MAX_PACK_CHARS` | Char budget for evidence packs. |
| `PORT`, `MCP_PATH` | Defaults `8080` and `/mcp`. |

`GET <MCP_PATH>/auth/check` returns 204 when `X-Google-Api-Key` is present (used
by the credential gateway).

## Develop

```bash
gofmt -l . && go vet ./... && go test -race ./...
```

## Release

Push a `vX.Y.Z` tag on `main`. The Release workflow publishes
`ghcr.io/tavric-labs/google-search-mcp` (multi-arch) and creates the GitHub
release. Deploy with MCP Runtime using `.mcp/servers.yaml` after setting
`GOOGLE_CSE_ID` and mirroring or retagging the image into your registry.
