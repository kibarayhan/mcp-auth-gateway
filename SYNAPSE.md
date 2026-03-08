# MCP Auth Gateway — Synapse

## Goal
Build and launch an open-source transparent MCP proxy that adds enterprise security (auth, policies, rate limiting, PII filtering, audit logging) to any MCP server.

## Current Progress
MVP complete. All 4 weeks of the design plan implemented and tested. Launch materials ready.

- **Week 1:** Transparent proxy — stdio MCP forwarding, tool discovery, routing (9 commits)
- **Week 2:** Auth + policy engine — API key auth, 3-level policy (server/tool/argument) (8 commits)
- **Week 3:** Audit log + rate limiter — JSONL audit trail, token-bucket rate limiting (7 commits)
- **Week 4:** PII filter + polish — regex PII redaction (emails, phones, cards, keys, JWTs, AWS ARNs), config validation, README, launch materials (12 commits)

Total: 36 commits, 40+ unit tests across 10 packages, 29 e2e scenarios across 4 test scripts.

See: `sessions/2026-03-08-mvp-build-and-launch-prep.md`

## What Worked
- TDD approach: test first → verify fail → implement → verify pass → commit
- Bite-sized tasks (2-5 min each) executed in batches
- Config types defined early (week 1) then filled in later weeks
- NotebookLM for generating launch video and podcast from design doc + README

## What Didn't Work
- NotebookLM video/audio generation takes 10-15 minutes — can't poll reliably from CLI
- macOS `/tmp` → `/private/tmp` symlink caused initial e2e test failures

## Key Files
- `main.go` — Cobra CLI + full proxy loop with auth, policy, rate limit, PII, audit
- `internal/auth/` — User types, API key authenticator
- `internal/policy/` — 3-level policy engine (server, tool, argument)
- `internal/ratelimit/` — Token-bucket rate limiter with rate string parsing
- `internal/pii/` — Regex PII filter (emails, phones, cards, keys, JWTs, AWS)
- `internal/audit/` — Structured JSONL audit logger
- `internal/gateway/` — Gateway core with routing, server registry
- `internal/config/` — YAML config loader with validation
- `internal/mcp/` — JSON-RPC protocol types
- `internal/transport/` — stdio message reader/writer
- `internal/upstream/` — Child process manager
- `gateway.yaml` — Example config with auth, policies, audit, rate limits, PII
- `.mcp.json` — Claude Code integration config
- `docs/launch-posts.md` — HN, r/ClaudeAI, r/MCP post drafts
- `docs/mcp-auth-gateway-video.mp4` — NotebookLM-generated video (29MB)
- `docs/mcp-auth-gateway-podcast.mp3` — NotebookLM-generated podcast (33MB)
- `demo.sh` — Runnable terminal demo script
- `LICENSE` — MIT

## Next Steps
1. Create GitHub repo: `gh repo create akibar/mcp-auth-gateway --public --source . --push`
2. Upload video/podcast to YouTube or GitHub releases (too large for git)
3. Post launch posts from `docs/launch-posts.md` to HN, r/ClaudeAI, r/MCP
4. Future: OAuth 2.0/OIDC auth provider, SSE transport, web dashboard
