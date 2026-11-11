# google-search-mcp

A token-efficient MCP server for **Google Custom Search** (Programmable Search JSON API).
It is designed so agents search with **short, targeted queries**, see **slim SERPs**, open
**few pages**, and answer from a **small citation pack**.

Inspired by agent search research (MindSearch, ManuSearch, Adaptive-RAG, CRAG, RECOMP, SAGE):
craft → search → selective open → pack → stop.

## Tools

| Tool | Calls Google? | Purpose |
|---|---|---|
| `google_craft_query` | No | Design 1–3 engine-ready queries + filter hints from a plain goal. |
| `google_search` | Yes | One Custom Search; returns title, url, snippet, `hit_id`. |
| `google_multi_search` | Yes | Up to 3 queries in parallel; merge/dedupe. |
| `google_search_papers` | Yes | Papers bias (`filetype:pdf`, arxiv or papers CSE). |
| `google_open` | No (HTTP fetch) | Short extract from one prior hit / URL (budgeted). |
| `google_pack_evidence` | No | Citation pack from selected `hit_id`s. |

Default flow: **craft → search → open 0–2 hits → pack → answer**.

## Configuration

| Variable | Purpose |
|---|---|
| `GOOGLE_API_KEY` | Google API key with Custom Search API enabled. |
| `GOOGLE_CSE_ID` | Programmable Search Engine `cx`. |
| `GOOGLE_CSE_PAPERS_ID` | Optional papers-only `cx`. |
| `GOOGLE_CSE_API_BASE` | Defaults to `https://www.googleapis.com/customsearch/v1`. |
| `GOOGLE_SEARCH_MAX_RESULTS` | Default hit count (capped at 10). |
| `GOOGLE_SEARCH_MAX_PARALLEL` | Max queries in `google_multi_search`. |
| `GOOGLE_SEARCH_MAX_OPEN_CHARS` | Char budget for `google_open`. |
| `GOOGLE_SEARCH_MAX_PACK_CHARS` | Char budget for evidence packs. |
| `PORT`, `MCP_PATH` | Defaults `8080` and `/mcp`. |

## Develop

```bash
gofmt -l . && go vet ./... && go test -race ./...
```

Deploy with MCP Runtime using `.mcp/servers.yaml` after filling CSE ids and creating the API key Secret.
