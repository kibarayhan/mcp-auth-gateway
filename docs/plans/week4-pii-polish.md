# MCP Auth Gateway — Week 4: PII Filter + Polish

> **For Claude:** REQUIRED SUB-SKILL: Use `/executing-plans` to implement this plan task-by-task.

**Goal:** Add regex-based PII filtering that redacts sensitive data from MCP server responses, add config validation, and polish the README for open-source launch.

**Architecture:** The PII filter scans upstream server responses for patterns matching emails, phone numbers, credit card numbers, API keys, and JWTs. When `pii_filter: true` is set on a server, matching content is replaced with `[REDACTED]` before being forwarded to the client. Config validation runs at startup and rejects invalid configurations with clear error messages. The README is rewritten for open-source launch.

**Tech Stack:** Go 1.26, `regexp` stdlib

**Project:** `~/projects/mcp-auth-gateway/`

---

### Task 1: PII Filter

**Files:**
- Create: `internal/pii/filter.go`
- Create: `internal/pii/filter_test.go`

### Task 2: Wire PII Filter into Proxy Loop

**Files:**
- Modify: `internal/gateway/gateway.go`
- Modify: `main.go`

### Task 3: Config Validation

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

### Task 4: Update Example Config

**Files:**
- Modify: `gateway.yaml`

### Task 5: README for Open-Source Launch

**Files:**
- Modify: `README.md`

### Task 6: E2E Integration Test — PII Filtering

**Files:**
- Create: `test/e2e_pii_test.sh`
