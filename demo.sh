#!/bin/bash
# demo.sh — 3-minute demo script for MCP Auth Gateway
# Run this with: bash demo.sh
# Record with: asciinema rec demo.cast && asciinema play demo.cast

set -e
cd ~/projects/mcp-auth-gateway

echo "╔══════════════════════════════════════════════════════╗"
echo "║         MCP Auth Gateway — Live Demo                ║"
echo "╚══════════════════════════════════════════════════════╝"
echo ""

# --- Scene 1: The Problem ---
echo "▸ PROBLEM: 3,500+ MCP servers. Almost none handle security."
echo "  No auth. No rate limits. No audit logs. No PII filtering."
echo ""
sleep 2

# --- Scene 2: The Solution ---
echo "▸ SOLUTION: A transparent proxy. Single binary. Simple YAML."
echo ""
echo "  Let's look at the config:"
echo ""
cat gateway.yaml
echo ""
sleep 3

# --- Scene 3: Build ---
echo "▸ BUILD: One command."
echo ""
echo "  \$ go build -o mcp-gateway ."
go build -o mcp-gateway .
echo "  Done. 12MB binary. Zero dependencies."
echo ""
sleep 2

# --- Scene 4: Auth ---
echo "▸ AUTH: API keys map to users with roles."
echo ""
echo "  Engineer (sk-eng-abc123) → can read files, blocked from writing"
echo "  Admin (sk-admin-xyz789)  → full access"
echo ""
sleep 2

# --- Scene 5: Live test — Engineer reads file ---
echo "▸ LIVE TEST 1: Engineer reads a file through the gateway"
echo ""
mkdir -p /private/tmp/mcp-gateway-test
echo "Hello from MCP Auth Gateway demo!" > /private/tmp/mcp-gateway-test/hello.txt

OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"demo","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

echo "  Response:"
echo "$OUTPUT" | python3 -c "
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    msg = json.loads(line)
    if msg.get('id') == 2:
        if 'result' in msg:
            print('  ✓ ALLOWED:', json.dumps(msg['result'], indent=4)[:200])
        elif 'error' in msg:
            print('  ✗ DENIED:', msg['error']['message'])
"
echo ""
sleep 2

# --- Scene 6: Live test — Engineer blocked from writing ---
echo "▸ LIVE TEST 2: Engineer tries to write (admin-only tool)"
echo ""

OUTPUT=$(MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"demo","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"write_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hack.txt","content":"nope"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml 2>/dev/null)

echo "  Response:"
echo "$OUTPUT" | python3 -c "
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line: continue
    msg = json.loads(line)
    if msg.get('id') == 2:
        if 'error' in msg:
            print('  ✗ BLOCKED:', msg['error']['message'])
        elif 'result' in msg:
            print('  ✓ Allowed (unexpected)')
"
echo ""
sleep 2

# --- Scene 7: Audit log ---
echo "▸ AUDIT LOG: Every call is logged."
echo ""
rm -f ./logs/audit.jsonl
mkdir -p ./logs

MCP_AUTH_TOKEN=sk-eng-abc123 printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"demo","version":"1.0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"read_file","arguments":{"path":"/private/tmp/mcp-gateway-test/hello.txt"}}}' \
  | MCP_AUTH_TOKEN=sk-eng-abc123 perl -e 'alarm 30; exec @ARGV' ./mcp-gateway start --config gateway.yaml >/dev/null 2>/dev/null

echo "  $ cat logs/audit.jsonl"
cat ./logs/audit.jsonl | python3 -m json.tool
echo ""
sleep 2

# --- Scene 8: Summary ---
echo "╔══════════════════════════════════════════════════════╗"
echo "║  6 security modules. Single binary. Simple YAML.    ║"
echo "║                                                     ║"
echo "║  ✓ Auth (API key / OAuth)                           ║"
echo "║  ✓ Policy Engine (server / tool / argument)         ║"
echo "║  ✓ Rate Limiter (token bucket)                      ║"
echo "║  ✓ PII Filter (emails, phones, cards, keys, JWTs)   ║"
echo "║  ✓ Audit Logger (structured JSONL)                  ║"
echo "║  ✓ Router (tool discovery, forwarding)              ║"
echo "║                                                     ║"
echo "║  github.com/akibar/mcp-auth-gateway                 ║"
echo "║  MIT License — Star it, try it, open issues.        ║"
echo "╚══════════════════════════════════════════════════════╝"
