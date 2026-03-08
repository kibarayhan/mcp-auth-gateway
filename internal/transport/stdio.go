package transport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"github.com/akibar/mcp-auth-gateway/internal/mcp"
)

// ReadMessage reads a single JSON-RPC message from a newline-delimited stream.
func ReadMessage(r io.Reader) (*mcp.JSONRPCMessage, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read: %w", err)
		}
		return nil, io.EOF
	}

	var msg mcp.JSONRPCMessage
	if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
		return nil, fmt.Errorf("parse JSON-RPC: %w", err)
	}

	return &msg, nil
}

// WriteMessage writes a single JSON-RPC message as a newline-delimited JSON line.
func WriteMessage(w io.Writer, msg *mcp.JSONRPCMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
