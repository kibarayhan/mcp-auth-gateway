# MCP Auth Gateway — Week 3: Audit Log + Rate Limiter

> **For Claude:** REQUIRED SUB-SKILL: Use `/executing-plans` to implement this plan task-by-task.

**Goal:** Add structured JSONL audit logging for every tool call and a token-bucket rate limiter that throttles excessive calls per user.

**Architecture:** The audit logger writes one JSON line per tool call to an append-only file. Each entry includes: timestamp, user, role, server, tool, args, decision (ALLOWED/DENIED/RATE_LIMITED), and duration. The rate limiter uses `golang.org/x/time/rate` with configurable limits parsed from the YAML config (e.g. `100/hour`, `10/minute`). Rate limit checks run after auth and policy checks but before forwarding to upstream. Both modules are wired into the existing proxy loop in `main.go`.

**Tech Stack:** Go 1.26, `golang.org/x/time/rate`, `log/slog` stdlib, `time` stdlib

**Project:** `~/projects/mcp-auth-gateway/`

---

### Task 1: Audit Logger

**Files:**
- Create: `internal/audit/logger.go`
- Create: `internal/audit/logger_test.go`

**Step 1: Write the test**

```go
// internal/audit/logger_test.go
package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogEntry(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	logger, err := NewLogger(tmp)
	if err != nil {
		t.Fatalf("NewLogger error: %v", err)
	}
	defer logger.Close()

	logger.Log(Entry{
		User:     "akibar",
		Roles:    []string{"engineer"},
		Server:   "filesystem",
		Tool:     "read_file",
		Args:     map[string]string{"path": "/tmp/test.txt"},
		Decision: "ALLOWED",
		Duration: 42 * time.Millisecond,
	})

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if entry.User != "akibar" {
		t.Errorf("User = %q, want %q", entry.User, "akibar")
	}
	if entry.Tool != "read_file" {
		t.Errorf("Tool = %q, want %q", entry.Tool, "read_file")
	}
	if entry.Decision != "ALLOWED" {
		t.Errorf("Decision = %q, want %q", entry.Decision, "ALLOWED")
	}
	if entry.Timestamp.IsZero() {
		t.Error("Timestamp should be set automatically")
	}
}

func TestLogMultipleEntries(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "audit.jsonl")
	logger, err := NewLogger(tmp)
	if err != nil {
		t.Fatalf("NewLogger error: %v", err)
	}
	defer logger.Close()

	logger.Log(Entry{User: "user1", Tool: "tool1", Decision: "ALLOWED"})
	logger.Log(Entry{User: "user2", Tool: "tool2", Decision: "DENIED"})

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("lines = %d, want 2", lines)
	}
}

func TestNilLogger(t *testing.T) {
	logger := &Logger{}
	// Should not panic on nil/empty logger
	logger.Log(Entry{User: "test", Decision: "ALLOWED"})
}
```

**Step 2: Run test — verify it fails**

```bash
cd ~/projects/mcp-auth-gateway
go test ./internal/audit/ -v
```

**Step 3: Write implementation**

```go
// internal/audit/logger.go
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp  time.Time         `json:"ts"`
	User       string            `json:"user"`
	Roles      []string          `json:"roles,omitempty"`
	Server     string            `json:"server,omitempty"`
	Tool       string            `json:"tool"`
	Args       map[string]string `json:"args,omitempty"`
	Decision   string            `json:"decision"`
	DurationMs int64             `json:"duration_ms,omitempty"`
	Duration   time.Duration     `json:"-"`
	Reason     string            `json:"reason,omitempty"`
}

// Logger writes audit entries as newline-delimited JSON.
type Logger struct {
	file *os.File
	mu   sync.Mutex
}

// NewLogger creates an audit logger that appends to the given file path.
func NewLogger(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit log: %w", err)
	}
	return &Logger{file: f}, nil
}

// Log writes an audit entry. Sets timestamp and duration_ms automatically.
func (l *Logger) Log(entry Entry) {
	if l.file == nil {
		return
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if entry.Duration > 0 {
		entry.DurationMs = entry.Duration.Milliseconds()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	l.file.Write(data)
	l.file.Write([]byte("\n"))
}

// Close closes the underlying file.
func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}
```

**Step 4: Run tests**

```bash
go test ./internal/audit/ -v
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/audit/
git commit -m "feat: structured JSONL audit logger with tests"
```

---

### Task 2: Rate Limiter

**Files:**
- Create: `internal/ratelimit/limiter.go`
- Create: `internal/ratelimit/limiter_test.go`

**Step 1: Write the test**

```go
// internal/ratelimit/limiter_test.go
package ratelimit

import "testing"

func TestParseRate(t *testing.T) {
	tests := []struct {
		input   string
		rate    float64
		wantErr bool
	}{
		{"100/hour", 100.0 / 3600, false},
		{"10/minute", 10.0 / 60, false},
		{"5/second", 5.0, false},
		{"", 0, true},
		{"invalid", 0, true},
		{"abc/hour", 0, true},
	}

	for _, tt := range tests {
		rate, err := ParseRate(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseRate(%q) should error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRate(%q) error: %v", tt.input, err)
			continue
		}
		if abs(rate-tt.rate) > 0.0001 {
			t.Errorf("ParseRate(%q) = %f, want %f", tt.input, rate, tt.rate)
		}
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func TestLimiter_Allow(t *testing.T) {
	lim := New()
	lim.Configure("akibar", "filesystem", 2.0) // 2 per second

	// First two should be allowed
	if !lim.Allow("akibar", "filesystem") {
		t.Error("first call should be allowed")
	}
	if !lim.Allow("akibar", "filesystem") {
		t.Error("second call should be allowed")
	}
}

func TestLimiter_NoConfig(t *testing.T) {
	lim := New()

	// No rate limit configured — always allow
	if !lim.Allow("anyone", "any-server") {
		t.Error("should allow when no rate limit configured")
	}
}
```

**Step 2: Run test — verify it fails**

```bash
go test ./internal/ratelimit/ -v
```

**Step 3: Write implementation**

```go
// internal/ratelimit/limiter.go
package ratelimit

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter manages per-user per-server rate limits.
type Limiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

// New creates a new rate limiter.
func New() *Limiter {
	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// Configure sets up a rate limit for a user+server combination.
func (l *Limiter) Configure(user, server string, ratePerSecond float64) {
	key := user + ":" + server
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limiters[key] = rate.NewLimiter(rate.Limit(ratePerSecond), int(ratePerSecond)+1)
}

// Allow checks if a request is within the rate limit.
// Returns true if allowed, false if rate limited.
// If no limiter is configured for this user+server, always allows.
func (l *Limiter) Allow(user, server string) bool {
	key := user + ":" + server
	l.mu.Lock()
	lim, ok := l.limiters[key]
	l.mu.Unlock()

	if !ok {
		return true
	}
	return lim.Allow()
}

// ParseRate parses a rate string like "100/hour", "10/minute", "5/second"
// into requests per second.
func ParseRate(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty rate string")
	}

	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rate format %q (expected N/unit)", s)
	}

	count, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate count %q: %w", parts[0], err)
	}

	switch strings.ToLower(parts[1]) {
	case "second", "sec", "s":
		return count, nil
	case "minute", "min", "m":
		return count / 60, nil
	case "hour", "hr", "h":
		return count / 3600, nil
	default:
		return 0, fmt.Errorf("unknown rate unit %q (use second/minute/hour)", parts[1])
	}
}
```

**Step 4: Add dependency and run tests**

```bash
cd ~/projects/mcp-auth-gateway
go get golang.org/x/time/rate
go test ./internal/ratelimit/ -v
```

Expected: 3 tests PASS.

**Step 5: Commit**

```bash
git add internal/ratelimit/ go.mod go.sum
git commit -m "feat: token-bucket rate limiter with configurable per-user limits"
```

---

### Task 3: Wire Audit Logger into Proxy Loop

**Files:**
- Modify: `internal/gateway/gateway.go`
- Modify: `main.go`

**Step 1: Add Audit field to Gateway**

In `internal/gateway/gateway.go`:
- Add `Audit *audit.Logger` field to Gateway struct
- Import `"github.com/akibar/mcp-auth-gateway/internal/audit"`

**Step 2: Update main.go**

- Create audit logger from config after gateway creation
- Set `gw.Audit = auditLogger`
- In the `tools/call` handler, log every decision (allowed, denied, rate limited) with duration tracking

**Step 3: Build and run all tests**

```bash
go build -o mcp-gateway .
go test ./... -v
```

**Step 4: Commit**

```bash
git add main.go internal/gateway/gateway.go
git commit -m "feat: wire audit logger into proxy loop"
```

---

### Task 4: Wire Rate Limiter into Proxy Loop

**Files:**
- Modify: `internal/gateway/gateway.go`
- Modify: `main.go`

**Step 1: Add RateLimiter field to Gateway**

In `internal/gateway/gateway.go`:
- Add `RateLimiter *ratelimit.Limiter` field
- Import `"github.com/akibar/mcp-auth-gateway/internal/ratelimit"`
- Initialize in `New()`

**Step 2: Update main.go**

- After starting servers, configure rate limiters from server policies
- In `tools/call` handler, check rate limit after policy checks but before forwarding
- If rate limited, return JSON-RPC error and audit log with decision "RATE_LIMITED"

**Step 3: Build and run all tests**

```bash
go build -o mcp-gateway .
go test ./... -v
```

**Step 4: Commit**

```bash
git add main.go internal/gateway/gateway.go
git commit -m "feat: wire rate limiter into proxy loop"
```

---

### Task 5: Update Config and Example

**Files:**
- Modify: `gateway.yaml`

**Step 1: Add audit config and rate limits**

```yaml
audit:
  destination: file
  path: "./logs/audit.jsonl"
  retention_days: 90
```

Add `rate_limit: "100/hour"` to the filesystem server policies.

**Step 2: Commit**

```bash
git add gateway.yaml
git commit -m "feat: example config with audit logging and rate limits"
```

---

### Task 6: E2E Integration Test — Audit + Rate Limiting

**Files:**
- Create: `test/e2e_audit_ratelimit_test.sh`

**Step 1: Create test script**

Tests:
1. Make a tool call → verify audit log file contains the entry
2. Make rapid calls with a low rate limit → verify rate limiting kicks in

**Step 2: Run and commit**

```bash
chmod +x test/e2e_audit_ratelimit_test.sh
bash test/e2e_audit_ratelimit_test.sh
git add test/e2e_audit_ratelimit_test.sh
git commit -m "test: e2e audit logging and rate limiting"
```

---

## Summary

| Task | What it builds | Tests |
|------|---------------|-------|
| 1 | Structured JSONL audit logger | 3 unit tests |
| 2 | Token-bucket rate limiter | 3 unit tests |
| 3 | Wire audit into proxy loop | Build + all tests |
| 4 | Wire rate limiter into proxy loop | Build + all tests |
| 5 | Config with audit + rate limits | Config update |
| 6 | E2E integration test | Audit file + rate limit scenarios |

After completing all 6 tasks: every tool call is logged to a JSONL file, excessive calls are throttled, and it's all tested end-to-end.
