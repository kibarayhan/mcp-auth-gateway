# MCP Auth Gateway

A transparent reverse proxy that adds enterprise security to any MCP server. Single binary, zero dependencies, simple YAML config.

```
AI Client (Claude Code, Agent SDK)
    |
    v
MCP Auth Gateway
    ├── Auth (API key / OAuth 2.0)
    ├── Policy Engine (server / tool / argument rules)
    ├── Rate Limiter (token bucket, per-user)
    ├── PII Filter (emails, phones, cards, keys, JWTs)
    ├── Audit Logger (structured JSONL)
    └── Router (tool discovery, forwarding)
    |
    v
Upstream MCP Servers (unmodified)
```

## Quick Start

```bash
go build -o mcp-gateway .
./mcp-gateway start --config gateway.yaml
```

## Configuration

```yaml
auth:
  provider: apikey
  users:
    - key: "sk-eng-abc123"
      name: "akibar"
      roles: ["engineer"]
    - key: "sk-admin-xyz789"
      name: "admin-bot"
      roles: ["admin", "engineer"]

audit:
  path: "./logs/audit.jsonl"

servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    policies:
      allowed_roles: ["engineer", "admin"]
      rate_limit: "100/hour"
      pii_filter: true
      tools:
        write_file:
          allowed_roles: ["admin"]
      blocked_args:
        - pattern: "(?i)DROP\\s+TABLE"
```

## Features

**Auth** — API key authentication. Each key maps to a user with roles. Set `MCP_AUTH_TOKEN` env var. No auth config = open access.

**Policy Engine** — Three levels of access control:
- Server-level: `allowed_roles: ["admin"]` blocks non-admins from the entire server
- Tool-level: restrict specific tools to specific roles
- Argument-level: `blocked_args` rejects calls matching regex patterns (e.g., `DROP TABLE`)

**Rate Limiter** — Token-bucket algorithm. `rate_limit: "100/hour"` per user per server. Supports `N/second`, `N/minute`, `N/hour`.

**PII Filter** — Scans upstream responses and redacts: emails, phone numbers, credit card numbers, API keys (AWS, GitHub, OpenAI), and JWTs. Enable per server with `pii_filter: true`.

**Audit Logger** — Every tool call produces a structured JSONL entry:
```json
{"ts":"2026-03-08T00:15:03Z","user":"akibar","roles":["engineer"],"server":"filesystem","tool":"read_file","decision":"ALLOWED","duration_ms":42}
```

Decisions: `ALLOWED`, `DENIED`, `RATE_LIMITED`.

## Claude Code Integration

Add to `.mcp.json` in your project:

```json
{
  "mcpServers": {
    "mcp-gateway": {
      "command": "/path/to/mcp-gateway",
      "args": ["start", "--config", "/path/to/gateway.yaml"],
      "env": {
        "MCP_AUTH_TOKEN": "sk-eng-abc123"
      }
    }
  }
}
```

Claude Code connects to the gateway as a stdio MCP server. The gateway spawns and manages upstream servers, discovers their tools, and exposes them all through a single endpoint.

## Running Tests

```bash
go test ./... -v                              # Unit tests (34 tests)
bash test/e2e_claude_code_test.sh             # Transparent proxy e2e
bash test/e2e_auth_test.sh                    # Auth + policy e2e
bash test/e2e_audit_ratelimit_test.sh         # Audit + rate limit e2e
bash test/e2e_pii_test.sh                     # PII filter e2e
```

## License

MIT
