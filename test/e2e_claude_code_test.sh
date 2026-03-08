#!/bin/bash
# test/e2e_claude_code_test.sh — Simulate Claude Code connecting to the gateway
set -e

cd ~/projects/mcp-auth-gateway

echo "=== Building gateway ==="
go build -o mcp-gateway .

echo "=== Creating test file ==="
mkdir -p /private/tmp/mcp-gateway-test
echo "Hello from MCP Gateway test!" > /private/tmp/mcp-gateway-test/hello.txt

echo "=== Starting gateway and sending Claude Code protocol sequence ==="
echo ""

# Simulate the exact sequence Claude Code sends:
# 1. initialize
# 2. notifications/initialized
# 3. tools/list
# 4. tools/call (read_file on our test file)
OUTPUT=$(printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude-code","version":"1.0.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/tmp/gateway-e2e.log)

echo "=== Gateway stdout (responses to Claude Code) ==="
echo "$OUTPUT" | python3 -m json.tool --no-ensure-ascii 2>/dev/null || echo "$OUTPUT"

echo ""
echo "=== Gateway stderr (logs) ==="
cat /tmp/gateway-e2e.log

echo ""
echo "=== RESULT ==="
if echo "$OUTPUT" | grep -q "Hello from MCP Gateway test"; then
  echo "SUCCESS: Gateway proxied tools/call and returned file contents!"
else
  echo "CHECKING: tools/list response..."
  if echo "$OUTPUT" | grep -q "read_file"; then
    echo "PARTIAL SUCCESS: Tool discovery works, checking tools/call..."
    echo "$OUTPUT" | python3 -c "
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    msg = json.loads(line)
    if msg.get('id') == 3:
        print('tools/call response:', json.dumps(msg, indent=2))
" 2>/dev/null || echo "Could not parse tools/call response"
  else
    echo "FAIL: Tool discovery did not work"
  fi
fi
