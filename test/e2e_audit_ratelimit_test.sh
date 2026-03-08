#!/bin/bash
# test/e2e_audit_ratelimit_test.sh — Test audit logging and rate limiting end-to-end
set -e

cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

mkdir -p /private/tmp/mcp-gateway-test
mkdir -p ./logs
rm -f ./logs/audit.jsonl
echo "test data" > /private/tmp/mcp-gateway-test/hello.txt

PASS_COUNT=0
FAIL_COUNT=0

echo "=== Test 1: Audit log — allowed call is logged ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

if [ -f ./logs/audit.jsonl ] && grep -q '"decision":"ALLOWED"' ./logs/audit.jsonl && grep -q '"tool":"read_file"' ./logs/audit.jsonl; then
  echo "PASS: Allowed call logged to audit file"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Audit log missing or incomplete"
  cat ./logs/audit.jsonl 2>/dev/null || echo "(no audit file)"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 2: Audit log entry has duration_ms ==="
if grep -q '"duration_ms"' ./logs/audit.jsonl; then
  echo "PASS: Audit entry contains duration_ms"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: No duration_ms in audit entries"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 3: Audit log — denied call is logged ==="
rm -f ./logs/audit.jsonl
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hack.txt","content":"nope"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

if [ -f ./logs/audit.jsonl ] && grep -q '"decision":"DENIED"' ./logs/audit.jsonl && grep -q '"tool":"write_file"' ./logs/audit.jsonl; then
  echo "PASS: Denied call logged to audit file"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Denied call not in audit log"
  cat ./logs/audit.jsonl 2>/dev/null || echo "(no audit file)"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 4: Rate limiting — rapid calls get throttled ==="
rm -f ./logs/audit.jsonl

# Create a config with very low rate limit (1/second, burst 2)
cat > /tmp/test-ratelimit.yaml << 'EOF'
gateway:
  listen: "localhost:3100"
auth:
  provider: apikey
  users:
    - key: "sk-test"
      name: "tester"
      roles: ["engineer"]
audit:
  path: "./logs/audit.jsonl"
servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/private/tmp/mcp-gateway-test"]
    policies:
      allowed_roles: ["engineer"]
      rate_limit: "1/second"
EOF

# Send 5 rapid calls — first ~2 should pass (burst), rest should be rate limited
OUTPUT=$(MCP_AUTH_TOKEN=sk-test printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"/private/tmp/mcp-gateway-test"}}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"/private/tmp/mcp-gateway-test"}}}' \
  '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"/private/tmp/mcp-gateway-test"}}}' \
  '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"/private/tmp/mcp-gateway-test"}}}' \
  '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"list_directory","arguments":{"path":"/private/tmp/mcp-gateway-test"}}}' \
  | MCP_AUTH_TOKEN=sk-test perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config /tmp/test-ratelimit.yaml 2>/dev/null)

if echo "$OUTPUT" | grep -q "rate limit exceeded"; then
  echo "PASS: Rate limiting triggered on rapid calls"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Rate limiting should have triggered"
  echo "$OUTPUT"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 5: Audit log shows RATE_LIMITED entries ==="
if grep -q '"decision":"RATE_LIMITED"' ./logs/audit.jsonl; then
  echo "PASS: Rate-limited calls appear in audit log"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: No RATE_LIMITED entries in audit log"
  cat ./logs/audit.jsonl 2>/dev/null
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Results: $PASS_COUNT passed, $FAIL_COUNT failed ==="
if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi
