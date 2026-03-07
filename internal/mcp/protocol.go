package mcp

import "encoding/json"

// JSONRPCMessage represents a JSON-RPC 2.0 message (request, response, or notification).
type JSONRPCMessage struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ToolCallParams is the params for a "tools/call" request.
type ToolCallParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

// ToolListResult is the result for a "tools/list" response.
type ToolListResult struct {
	Tools []ToolInfo `json:"tools"`
}

type ToolInfo struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema,omitempty"`
}

// IsRequest returns true if this message has an ID (it's a request, not a notification).
func (m *JSONRPCMessage) IsRequest() bool {
	return m.ID != nil
}

// IsResponse returns true if this message has a result or error.
func (m *JSONRPCMessage) IsResponse() bool {
	return m.Result != nil || m.Error != nil
}
