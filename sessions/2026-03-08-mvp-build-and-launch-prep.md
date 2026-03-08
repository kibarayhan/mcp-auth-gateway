# Session: MVP Build and Launch Prep

**Date:** 2026-03-08
**Duration:** Single session — full 4-week plan executed
**Workspace:** mcp-auth-gateway

## Summary

Built the entire MCP Auth Gateway MVP from scratch in one session. Executed all 4 weeks of the design plan: transparent proxy (week 1), auth + policy engine (week 2), audit log + rate limiter (week 3), PII filter + polish (week 4). Then prepared launch materials: MIT license, demo script, HN/Reddit post drafts, and NotebookLM-generated video + podcast. Updated design doc to remove company-specific references and stale data.

## Architecture

```
AI Client (Claude Code) → stdio → MCP Auth Gateway → stdio → Upstream MCP Servers
```

Gateway intercepts every JSON-RPC message. For `tools/call`:
1. Auth check (API key → user + roles)
2. Server-level policy check
3. Tool-level policy check
4. Argument-level blocked pattern check
5. Rate limit check
6. Forward to upstream
7. PII filter on response
8. Audit log entry
9. Forward response to client

## Code Stats

| Package | Files | Tests | Purpose |
|---------|-------|-------|---------|
| auth | 3 | 7 | User types, API key authenticator |
| config | 2 | 9 | YAML loader, validation |
| gateway | 2 | 2 | Core routing, server registry |
| mcp | 2 | 3 | JSON-RPC protocol types |
| transport | 2 | 2 | stdio message reader/writer |
| upstream | 2 | 2 | Child process manager |
| policy | 2 | 6 | 3-level policy engine |
| ratelimit | 2 | 3 | Token-bucket rate limiter |
| audit | 2 | 3 | JSONL audit logger |
| pii | 2 | 7 | Regex PII filter |
| **Total** | **21** | **44** | |

E2E test scripts: 4 (proxy, auth, audit+ratelimit, pii) covering 29 scenarios.

## Decisions

- **stdio transport only for MVP** — SSE/HTTP deferred. Claude Code uses stdio natively.
- **API key auth only for MVP** — OAuth 2.0/OIDC deferred. API keys are simple and testable.
- **In-memory rate limiting** — No Redis. Token-bucket per user+server. Resets on restart.
- **Append-only JSONL audit** — No database. Simple file. Rotation/retention not implemented yet.
- **PII filter on response only** — Doesn't filter request args (user's own input). Only filters upstream responses.
- **No tool name prefixing** — Gateway exposes upstream tool names as-is. No `servername__toolname` namespacing.
- **macOS `/private/tmp` path** — Tests use `/private/tmp` to avoid macOS symlink issues.

## PII Patterns

| Pattern | Placeholder | Added |
|---------|------------|-------|
| Email addresses | [REDACTED_EMAIL] | Week 4 |
| Phone numbers | [REDACTED_PHONE] | Week 4 |
| Credit card numbers | [REDACTED_CC] | Week 4 |
| AWS access keys (AKIA...) | [REDACTED_KEY] | Week 4 |
| AWS secret keys | [REDACTED_SECRET] | Post-MVP |
| AWS ARNs | [REDACTED_ARN] | Post-MVP |
| AWS account IDs (12-digit) | [AWS_ACCOUNT] | Post-MVP |
| GitHub PATs (ghp_...) | [REDACTED_KEY] | Week 4 |
| OpenAI keys (sk-proj-...) | [REDACTED_KEY] | Week 4 |
| JWTs (eyJ...) | [REDACTED_JWT] | Week 4 |

## Launch Materials

| Asset | Path | Status |
|-------|------|--------|
| MIT License | `LICENSE` | Committed |
| README | `README.md` | Committed |
| Demo script | `demo.sh` | Committed |
| HN/Reddit drafts | `docs/launch-posts.md` | Committed |
| Video (29MB) | `docs/mcp-auth-gateway-video.mp4` | Downloaded, not in git |
| Podcast (33MB) | `docs/mcp-auth-gateway-podcast.mp3` | Downloaded, not in git |

## Config Changes by User Post-MVP

- Added AWS serverless MCP server to `gateway.yaml` with `AWS_PROFILE: "dev"` env
- Added AWS-specific PII patterns to `internal/pii/filter.go` (ARNs, account IDs, secret keys, region URLs)
- Removed company-specific references from design doc
- Replaced "3,500+" with generic "growing fast" language

## Errors Encountered

- `timeout` command not available on macOS — switched to `perl -e 'alarm 30; exec @ARGV'`
- macOS `/tmp` symlink to `/private/tmp` caused filesystem MCP server to reject paths — fixed by using `/private/tmp` in configs and tests
- NotebookLM auth expires frequently — need to re-run `notebooklm login` before each session
- `duration_ms` omitted from audit log when 0 (json `omitempty` on int64) — reordered test to check after an allowed call with real duration
