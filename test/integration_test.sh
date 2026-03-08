#!/bin/bash
# test/integration_test.sh — Test gateway with a real MCP server
set -e

echo "Building gateway..."
cd ~/projects/mcp-auth-gateway
go build -o mcp-gateway .

echo "Testing: send initialize + tools/list through gateway..."

# Create a minimal test config using the filesystem MCP server
cat > /tmp/test-gateway.yaml << 'EOF'
gateway:
  listen: "localhost:3100"
servers:
  - name: filesystem
    command: "npx"
    args: ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
EOF

# Send initialize and tools/list, capture output
# Use perl-based timeout for macOS compatibility
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}
{"jsonrpc":"2.0","method":"notifications/initialized"}
{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config /tmp/test-gateway.yaml 2>/tmp/gateway-test.log

echo ""
echo "Gateway logs:"
cat /tmp/gateway-test.log
echo ""
echo "DONE"
