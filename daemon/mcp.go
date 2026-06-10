package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// runMCP starts an MCP server over stdio exposing workspace state tools.
func runMCP(state *StateManager) {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "echostate",
		Version: version,
		Title:   "EchoState Workspace Context",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workspace_context",
		Description: "Return the current workspace state including .echo-state.md notes, file tree, and latest git commit.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, StateResponse, error) {
		if err := state.Refresh(); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "failed to refresh workspace state: " + err.Error()},
				},
			}, StateResponse{}, nil
		}
		resp := state.Get()
		return nil, resp, nil
	})

	log.Printf("echostate MCP server starting (stdio)")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("mcp server: %v", err)
	}
}
