#!/bin/bash
# test/e2e_pii_test.sh — Test PII filtering end-to-end
set -e

cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

mkdir -p /private/tmp/mcp-gateway-test
mkdir -p ./logs

PASS_COUNT=0
FAIL_COUNT=0

# Create a file with PII content
cat > /private/tmp/mcp-gateway-test/pii-data.txt << 'PIIEOF'
Customer: John Doe
Email: john.doe@example.com
Phone: +1-555-123-4567
Card: 4111-1111-1111-1111
API Key: sk-proj-abcdef123456789012
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U
PIIEOF

# Create config with PII filter enabled
cat > /tmp/test-pii.yaml << 'EOF'
auth:
  provider: apikey
  users:
    - key: "sk-test"
      name: "tester"
      roles: ["engineer"]
servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/private/tmp/mcp-gateway-test"]
    policies:
      allowed_roles: ["engineer"]
      pii_filter: true
EOF

echo "=== Test 1: PII filter redacts email ==="
OUTPUT=$(MCP_AUTH_TOKEN=sk-test printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/pii-data.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-test perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config /tmp/test-pii.yaml 2>/dev/null)

if echo "$OUTPUT" | grep -q "john.doe@example.com"; then
  echo "FAIL: Email should be redacted"
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  echo "PASS: Email is redacted"
  PASS_COUNT=$((PASS_COUNT + 1))
fi

echo ""
echo "=== Test 2: PII filter redacts phone number ==="
if echo "$OUTPUT" | grep -q "555-123-4567"; then
  echo "FAIL: Phone should be redacted"
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  echo "PASS: Phone is redacted"
  PASS_COUNT=$((PASS_COUNT + 1))
fi

echo ""
echo "=== Test 3: PII filter redacts API key ==="
if echo "$OUTPUT" | grep -q "sk-proj-abcdef"; then
  echo "FAIL: API key should be redacted"
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  echo "PASS: API key is redacted"
  PASS_COUNT=$((PASS_COUNT + 1))
fi

echo ""
echo "=== Test 4: PII filter redacts JWT ==="
if echo "$OUTPUT" | grep -q "eyJhbGci"; then
  echo "FAIL: JWT should be redacted"
  FAIL_COUNT=$((FAIL_COUNT + 1))
else
  echo "PASS: JWT is redacted"
  PASS_COUNT=$((PASS_COUNT + 1))
fi

echo ""
echo "=== Test 5: Non-PII content preserved ==="
if echo "$OUTPUT" | grep -q "Customer"; then
  echo "PASS: Non-PII content preserved"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: Non-PII content should be preserved"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Test 6: PII filter disabled — no redaction ==="
cat > /tmp/test-nopii.yaml << 'EOF'
auth:
  provider: apikey
  users:
    - key: "sk-test"
      name: "tester"
      roles: ["engineer"]
servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/private/tmp/mcp-gateway-test"]
    policies:
      allowed_roles: ["engineer"]
      pii_filter: false
EOF

OUTPUT_NOPII=$(MCP_AUTH_TOKEN=sk-test printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/pii-data.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-test perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config /tmp/test-nopii.yaml 2>/dev/null)

if echo "$OUTPUT_NOPII" | grep -q "john.doe@example.com"; then
  echo "PASS: PII preserved when filter disabled"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  echo "FAIL: PII should NOT be redacted when filter disabled"
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Results: $PASS_COUNT passed, $FAIL_COUNT failed ==="
if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi
