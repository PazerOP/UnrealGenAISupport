package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSpecialTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("how_to_use",
			mcp.WithDescription("Hey LLM, this grabs the how_to_use.md from knowledge_base\u2014it's your cheat sheet for running Unreal with this MCP. Fetch it at the start of a new chat session to get the lowdown on quirks and how shit works."),
		),
		handleHowToUse,
	)

	s.AddTool(
		mcp.NewTool("take_editor_screenshot",
			mcp.WithDescription("Takes a screenshot of the primary monitor using a vendored OS-level library. This is a robust method that requires no installation and bypasses the Unreal API."),
		),
		handleTakeEditorScreenshot,
	)
}

func handleHowToUse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Try multiple candidate paths to find knowledge_base/how_to_use.md
	candidates := []string{}

	// Check environment variable first (set by init_unreal.py)
	if pluginPath := os.Getenv("UNREAL_PLUGIN_PATH"); pluginPath != "" {
		candidates = append(candidates,
			filepath.Join(pluginPath, "Content", "Python", "knowledge_base", "how_to_use.md"),
		)
	}

	// Relative to executable
	candidates = append(candidates,
		filepath.Join(execDir, "knowledge_base", "how_to_use.md"),
		filepath.Join(execDir, "..", "Content", "Python", "knowledge_base", "how_to_use.md"),
		filepath.Join(execDir, "..", "..", "Content", "Python", "knowledge_base", "how_to_use.md"),
	)

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return mcp.NewToolResultText(string(data)), nil
		}
	}

	return mcp.NewToolResultText("Error: how_to_use.md not found in knowledge_base subfolder."), nil
}

func handleTakeEditorScreenshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imgBase64, err := captureScreenshot()
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("OS-level screenshot failed: %v", err)), nil
	}
	return mcp.NewToolResultImage("Editor screenshot", imgBase64, "image/png"), nil
}
