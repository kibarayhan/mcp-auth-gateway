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
