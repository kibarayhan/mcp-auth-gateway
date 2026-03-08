package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/akibar/mcp-auth-gateway/internal/auth"
	"github.com/akibar/mcp-auth-gateway/internal/config"
	"github.com/akibar/mcp-auth-gateway/internal/gateway"
	"github.com/akibar/mcp-auth-gateway/internal/mcp"
	"github.com/akibar/mcp-auth-gateway/internal/transport"
	"github.com/akibar/mcp-auth-gateway/internal/upstream"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-gateway",
	Short: "MCP Auth Gateway — secure proxy for MCP servers",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway",
	RunE:  runStart,
}

func init() {
	startCmd.Flags().StringP("config", "c", "gateway.yaml", "Path to config file")
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	slog.Info("starting gateway", "config", configPath, "servers", len(cfg.Servers))

	gw := gateway.New(cfg)
	ctx := context.Background()

	// Authenticate caller
	authenticator := auth.NewAPIKeyAuth(cfg.Auth.Users)
	token := os.Getenv("MCP_AUTH_TOKEN")
	user, err := authenticator.Authenticate(token)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	gw.User = user
	slog.Info("authenticated", "user", user.Name, "roles", user.Roles, "authenticated", user.Authenticated)

	// Start upstream servers and discover their tools
	for _, serverCfg := range cfg.Servers {
		slog.Info("starting upstream server", "name", serverCfg.Name, "command", serverCfg.Command)

		srv, err := upstream.Start(ctx, serverCfg.Command, serverCfg.Args, nil)
		if err != nil {
			return fmt.Errorf("start server %s: %w", serverCfg.Name, err)
		}
		srv.Name = serverCfg.Name
		gw.SetServer(serverCfg.Name, srv)

		// Send initialize request to discover capabilities
		initReq := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "initialize",
			Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"mcp-gateway","version":"0.1.0"}}`),
		}
		id := json.RawMessage(`1`)
		initReq.ID = &id

		if err := transport.WriteMessage(srv.Stdin, initReq); err != nil {
			slog.Error("failed to send initialize", "server", serverCfg.Name, "error", err)
			continue
		}

		// Read initialize response
		resp, err := transport.ReadMessage(srv.Stdout)
		if err != nil {
			slog.Error("failed to read initialize response", "server", serverCfg.Name, "error", err)
			continue
		}
		slog.Debug("initialize response", "server", serverCfg.Name, "result", string(resp.Result))

		// Send initialized notification
		initNotif := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "notifications/initialized",
		}
		transport.WriteMessage(srv.Stdin, initNotif)

		// Request tools list
		toolsReq := &mcp.JSONRPCMessage{
			JSONRPC: "2.0",
			Method:  "tools/list",
		}
		toolsID := json.RawMessage(`2`)
		toolsReq.ID = &toolsID

		if err := transport.WriteMessage(srv.Stdin, toolsReq); err != nil {
			slog.Error("failed to request tools", "server", serverCfg.Name, "error", err)
			continue
		}

		toolsResp, err := transport.ReadMessage(srv.Stdout)
		if err != nil {
			slog.Error("failed to read tools", "server", serverCfg.Name, "error", err)
			continue
		}

		var toolsResult mcp.ToolListResult
		if err := json.Unmarshal(toolsResp.Result, &toolsResult); err != nil {
			slog.Error("failed to parse tools", "server", serverCfg.Name, "error", err)
			continue
		}

		gw.RegisterTools(serverCfg.Name, toolsResult.Tools)
		slog.Info("discovered tools", "server", serverCfg.Name, "count", len(toolsResult.Tools))
	}

	slog.Info("gateway ready", "total_tools", len(gw.AllTools()))

	// Main proxy loop: read from stdin, route to upstream, write to stdout
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var msg mcp.JSONRPCMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			slog.Error("invalid JSON from client", "error", err)
			continue
		}

		switch msg.Method {
		case "initialize":
			// Respond with our own capabilities
			result := map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":   map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo":     map[string]interface{}{"name": "mcp-gateway", "version": "0.1.0"},
			}
			resultJSON, _ := json.Marshal(result)
			resp := &mcp.JSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  resultJSON,
			}
			transport.WriteMessage(os.Stdout, resp)

		case "notifications/initialized":
			// Ignore

		case "tools/list":
			// Return aggregated tools from all upstream servers
			result := mcp.ToolListResult{Tools: gw.AllTools()}
			resultJSON, _ := json.Marshal(result)
			resp := &mcp.JSONRPCMessage{
				JSONRPC: "2.0",
				ID:      msg.ID,
				Result:  resultJSON,
			}
			transport.WriteMessage(os.Stdout, resp)

		case "tools/call":
			// Route to the correct upstream server
			var params mcp.ToolCallParams
			json.Unmarshal(msg.Params, &params)

			serverName, ok := gw.RouteToolCall(params.Name)
			if !ok {
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32601, Message: fmt.Sprintf("tool %q not found", params.Name)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			srv, err := gw.GetServer(serverName)
			if err != nil {
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32603, Message: err.Error()},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			// Policy checks
			serverCfg, _ := gw.ServerConfigByName(serverName)

			decision := gw.Policy.CheckServerAccess(gw.User, serverCfg)
			if !decision.Allowed {
				slog.Warn("policy denied", "user", gw.User.Name, "server", serverName, "tool", params.Name, "reason", decision.Reason)
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32600, Message: fmt.Sprintf("access denied: %s", decision.Reason)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			decision = gw.Policy.CheckToolAccess(gw.User, serverCfg, params.Name, params.Arguments)
			if !decision.Allowed {
				slog.Warn("policy denied", "user", gw.User.Name, "server", serverName, "tool", params.Name, "reason", decision.Reason)
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32600, Message: fmt.Sprintf("access denied: %s", decision.Reason)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			slog.Info("policy allowed", "user", gw.User.Name, "server", serverName, "tool", params.Name)

			// Forward to upstream
			transport.WriteMessage(srv.Stdin, &msg)

			// Read response from upstream
			resp, err := transport.ReadMessage(srv.Stdout)
			if err != nil {
				if err == io.EOF {
					slog.Error("upstream server closed", "server", serverName)
				}
				errResp := &mcp.JSONRPCMessage{
					JSONRPC: "2.0",
					ID:      msg.ID,
					Error:   &mcp.JSONRPCError{Code: -32603, Message: fmt.Sprintf("upstream error: %v", err)},
				}
				transport.WriteMessage(os.Stdout, errResp)
				continue
			}

			// Forward response back to client
			resp.ID = msg.ID
			transport.WriteMessage(os.Stdout, resp)

		default:
			slog.Debug("unhandled method", "method", msg.Method)
		}
	}

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
