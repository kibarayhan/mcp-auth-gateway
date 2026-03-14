# MCP Auth Gateway — Synapse

## Goal
Build and launch an open-source transparent MCP proxy that adds enterprise security (auth, policies, rate limiting, PII filtering, audit logging) to any MCP server.

## Current Progress
MVP complete and live on GitHub: https://github.com/kibarayhan/mcp-auth-gateway

- All 4 weeks of the design plan implemented and tested (36 commits, 40+ unit tests, 29 e2e scenarios)
- Parallel upstream init with mutex-safe gateway added (2026-03-15)
- Repo pushed to GitHub (private → public)
- Branch protection not yet configured

## What Worked
- TDD approach: test first → verify fail → implement → verify pass → commit
- Bite-sized tasks (2-5 min each) executed in batches
- Config types defined early (week 1) then filled in later weeks
- NotebookLM for generating launch video and podcast from design doc + README

## What Didn't Work
- NotebookLM video/audio generation takes 10-15 minutes — can't poll reliably from CLI
- macOS `/tmp` → `/private/tmp` symlink caused initial e2e test failures

## Key Files
- `main.go` — Cobra CLI + full proxy loop with parallel upstream init
- `internal/auth/` — User types, API key authenticator
- `internal/policy/` — 3-level policy engine (server, tool, argument)
- `internal/ratelimit/` — Token-bucket rate limiter
- `internal/pii/` — Regex PII filter (emails, phones, cards, keys, JWTs, AWS)
- `internal/audit/` — Structured JSONL audit logger
- `internal/gateway/` — Gateway core with mutex-safe routing and Ready channel
- `docs/launch-posts.md` — HN, r/ClaudeAI, r/MCP post drafts (ready to publish)
- `docs/mcp-auth-gateway-video.mp4` — NotebookLM-generated video (29MB, not in git)
- `docs/mcp-auth-gateway-podcast.mp3` — NotebookLM-generated podcast (33MB, not in git)
- `sessions/2026-03-08-mvp-build-and-launch-prep.md` — MVP build session record
- `sessions/2026-03-15-github-launch-prep.md` — This session

## Next Steps
1. Configure branch protection on main (require PRs, 1 approval, no force push)
2. Upload video/podcast to YouTube or GitHub Releases
3. Post to HN, r/ClaudeAI, r/MCP — drafts in `docs/launch-posts.md`
4. Fix `gateway.yaml` — revert hardcoded `/Users/akibar/...` audit path back to `./logs/audit.jsonl`
5. Future: OAuth 2.0/OIDC auth provider, SSE transport, web dashboard
