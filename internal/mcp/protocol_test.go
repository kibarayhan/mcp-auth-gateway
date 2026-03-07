package mcp

import (
	"encoding/json"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"search","arguments":{"query":"test"}}}`

	var msg JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if msg.Method != "tools/call" {
		t.Errorf("Method = %q, want %q", msg.Method, "tools/call")
	}

	if msg.ID == nil {
		t.Error("ID should not be nil for a request")
	}
}

func TestParseNotification(t *testing.T) {
	raw := `{"jsonrpc":"2.0","method":"notifications/initialized"}`

	var msg JSONRPCMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if msg.ID != nil {
		t.Error("ID should be nil for a notification")
	}
}

func TestToolCallParams(t *testing.T) {
	raw := `{"name":"search_code","arguments":{"query":"auth token","repo":"infra-main"}}`

	var params ToolCallParams
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if params.Name != "search_code" {
		t.Errorf("Name = %q, want %q", params.Name, "search_code")
	}

	if params.Arguments["query"] != "auth token" {
		t.Errorf("query arg = %q, want %q", params.Arguments["query"], "auth token")
	}
}
