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
