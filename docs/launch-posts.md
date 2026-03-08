# Launch Posts — MCP Auth Gateway

## Hacker News

**Title:** Show HN: MCP Auth Gateway – Add auth, rate limits, and PII filtering to any MCP server

**URL:** https://github.com/akibar/mcp-auth-gateway

**Comment:**

I built a transparent reverse proxy that sits in front of MCP servers and adds enterprise security. Single Go binary, simple YAML config, works with Claude Code out of the box.

The MCP ecosystem is growing fast — thousands of servers and counting. Almost none handle authentication, authorization, rate limiting, or audit logging. Every company adopting MCP for AI agents faces the same gap.

MCP Auth Gateway solves this with six modules:

- **Auth** — API key authentication (OAuth 2.0 coming). Each key maps to a user with roles.
- **Policy engine** — Three levels: server-level ("only admins access the database"), tool-level ("engineers read Slack, admins post"), argument-level ("block DROP TABLE").
- **Rate limiter** — Token bucket. "100/hour" per user per server. Prevents runaway AI agent loops.
- **PII filter** — Regex scans responses for emails, phone numbers, credit cards, API keys, JWTs. Redacts before the AI sees them.
- **Audit logger** — Structured JSONL. Every tool call: who, what, when, allowed/denied, duration.
- **Router** — Discovers tools from upstream servers, aggregates them, routes calls.

The AI client thinks it's talking to MCP servers directly. The MCP servers don't know the proxy exists. All config lives in one YAML file.

Built in Go, no external dependencies. 12MB binary. MIT licensed.

Happy to answer questions about the architecture or MCP protocol details.

---

## Reddit r/ClaudeAI

**Title:** I built an auth gateway for MCP servers — role-based access, rate limiting, PII filtering, audit logs

**Body:**

If you're using MCP servers with Claude Code or the Agent SDK, you've probably noticed they have zero security built in. Any tool call goes through, no questions asked.

I built **MCP Auth Gateway** — a transparent proxy that adds:

- API key auth with role-based access control
- Per-tool policies (e.g., engineers can read Slack but only admins can post)
- Argument blocking (e.g., reject any SQL with "DROP TABLE")
- Rate limiting (token bucket, configurable per user)
- PII filtering (redacts emails, phone numbers, credit cards, API keys from responses)
- Full audit trail (every tool call logged to JSONL)

It's a single Go binary with YAML config. You point Claude Code at the gateway instead of directly at MCP servers. The gateway discovers tools, enforces policies, and forwards everything transparently.

Example: an engineer using Claude Code can read files but gets blocked from write_file (admin only). All of this is configurable per server, per tool.

Open source, MIT licensed: https://github.com/akibar/mcp-auth-gateway

Would love feedback from anyone running MCP servers in production.

---

## Reddit r/MCP

**Title:** Open-source auth gateway for MCP servers — adds security to any existing server without modification

**Body:**

Built a transparent proxy for MCP servers that adds auth, policies, rate limiting, PII filtering, and audit logging. The upstream servers don't need any changes.

Config is one YAML file:

```yaml
servers:
  - name: database
    command: "npx @anthropic-ai/mcp-server-postgres"
    policies:
      allowed_roles: ["admin"]
      rate_limit: "20/hour"
      pii_filter: true
      blocked_args:
        - pattern: "DROP TABLE"
```

Works with Claude Code as a stdio MCP server. Single Go binary, MIT license.

https://github.com/akibar/mcp-auth-gateway
