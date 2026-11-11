# google-search-mcp

A **generic**, token-efficient MCP server for searching **anything** on the web via
**Gemini Google Search grounding** (Google AI Studio). No Custom Search Engine
(`cx`) and no `GOOGLE_CSE_ID`. Public Runtime deploys can run with no client login
(platform-held key + gateway audit); callers may also send their own key.

Agents are steered to use **short targeted queries**, read **slim SERPs**, open
**few pages**, and answer from a **small citation pack** (craft → search →
selective open → pack → stop).

## Get a Gemini API key

1. Open [Google AI Studio → API keys](https://aistudio.google.com/apikey).
2. Sign in with the Google account that will own the quota.
3. Click **Create API key** (create or pick a Google Cloud project when asked).
4. Copy the key once. Store it in a password manager or a local gitignored `.env`.
   Do **not** paste keys into chat, issues, PRs, or git.
5. Optional: in AI Studio open **Usage / rate limits** for that project to see
   your live RPM / TPM / RPD caps.

### Use the key with this MCP

- **Public Runtime (default in `.mcp/servers.yaml`):** `auth.mode: none`,
  `policy.mode: observe`, `gateway.enabled: true`. The pod Secret supplies
  `GEMINI_API_KEY`. Clients need no login; gateway audit records tool use.
- **Bring-your-own key:** send a header (overrides the platform key when both
  are present):

```http
X-Gemini-Api-Key: <gemini-ai-studio-api-key>
```

Local smoke (never commit the value):

```bash
# gitignored
echo 'GEMINI_API_KEY=...' > .env
chmod 600 .env
```

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
- Model RPM / TPM / RPD caps still apply on top of the Search RPD limit — see
  [rate limits](https://ai.google.dev/gemini-api/docs/rate-limits) in AI Studio.

Practical free-tier budget with the default model: on the order of **~500
`google_search` calls/day**, or fewer if you use `google_multi_search`.

## Tools

| Tool | Calls Gemini Search? | Purpose |
|---|---|---|
| `google_craft_query` | No | Design 1–3 engine-ready queries from any plain-language goal. |
| `google_search` | Yes | Slim hits (`title`, `url`, `snippet`, `hit_id`) from grounding citations. |
| `google_multi_search` | Yes (≤3) | Up to 3 queries in parallel; merge/dedupe. |
| `google_open` | No (HTTP fetch) | Short extract from one prior hit / URL (budgeted). |
| `google_pack_evidence` | No | Citation pack from selected `hit_id`s. |

Default flow: **craft → search → open 0–2 hits → pack → answer**.

## Configuration (operators)

| Variable | Purpose |
|---|---|
| `GEMINI_MODEL` | Defaults to `gemini-2.5-flash` (free-tier Search path). |
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
