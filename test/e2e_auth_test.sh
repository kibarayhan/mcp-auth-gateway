#!/bin/bash
# test/e2e_auth_test.sh — Test auth + policy engine end-to-end
set -e

cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

mkdir -p /private/tmp/mcp-gateway-test
echo "secret data" > /private/tmp/mcp-gateway-test/hello.txt

PASS_COUNT=0
FAIL_COUNT=0

echo "=== Test 1: Authenticated engineer — read_file ALLOWED ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

if echo "$OUTPUT" | grep -q "secret data"; then
  echo "PASS: Engineer can read files"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Engineer should be able to read files"
  echo "$OUTPUT"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 2: Authenticated engineer — write_file DENIED (admin only) ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hack.txt","content":"pwned"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

if echo "$OUTPUT" | grep -q "access denied"; then
  echo "PASS: Engineer blocked from write_file (admin only)"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Engineer should be blocked from write_file"
  echo "$OUTPUT"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 3: No auth token — DENIED ==="
OUTPUT=$(unset MCP_AUTH_TOKEN; printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | perl -e 'alarm 10; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>&1 || true)

if echo "$OUTPUT" | grep -qi "auth"; then
  echo "PASS: No token rejected"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Should reject missing token"
  echo "$OUTPUT"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 4: Admin — write_file ALLOWED ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-admin-xyz789 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/admin-wrote.txt","content":"admin was here"}}}' \
  | MCP_AUTH_TOKEN=sk-admin-xyz789 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

if echo "$OUTPUT" | grep -q '"id":2' && ! echo "$OUTPUT" | grep -q "access denied"; then
  echo "PASS: Admin can write files"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Admin should be able to write files"
  echo "$OUTPUT"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Results: $PASS_COUNT passed, $FAIL_COUNT failed ==="
if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi
