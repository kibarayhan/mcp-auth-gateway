package transport

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/akibar/mcp-auth-gateway/internal/mcp"
)

func TestReadMessage(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}` + "\n"
	reader := bytes.NewBufferString(input)

	msg, err := ReadMessage(reader)
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}

	if msg.Method != "tools/list" {
		t.Errorf("Method = %q, want %q", msg.Method, "tools/list")
	}
}

func TestWriteMessage(t *testing.T) {
	msg := &mcp.JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "test",
	}

	var buf bytes.Buffer
	err := WriteMessage(&buf, msg)
	if err != nil {
		t.Fatalf("WriteMessage error: %v", err)
	}

	// Verify it's valid JSON followed by newline
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	var parsed mcp.JSONRPCMessage
	if err := json.Unmarshal(lines[0], &parsed); err != nil {
		t.Fatalf("Output not valid JSON: %v", err)
	}

	if parsed.Method != "test" {
		t.Errorf("Method = %q, want %q", parsed.Method, "test")
	}
}
