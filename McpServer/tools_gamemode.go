package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerGameModeTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("create_game_mode",
			mcp.WithDescription("Create a game mode Blueprint, set its default pawn, and assign it as the current scene's default game mode."),
			mcp.WithString("game_mode_path",
				mcp.Required(),
				mcp.Description("Path for new game mode (e.g., \"/Game/MyGameMode\")"),
			),
			mcp.WithString("pawn_blueprint_path",
				mcp.Required(),
				mcp.Description("Path to pawn Blueprint (e.g., \"/Game/Blueprints/BP_Player\")"),
			),
			mcp.WithString("base_class",
				mcp.Description("Base class for game mode (default: \"GameModeBase\")"),
			),
		),
		handleCreateGameMode,
	)
}

func handleCreateGameMode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameModePath := req.GetString("game_mode_path", "")
	pawnBlueprintPath := req.GetString("pawn_blueprint_path", "")
	baseClass := req.GetString("base_class", "GameModeBase")

	resp := sendToUnreal(map[string]interface{}{
		"type":                 "create_game_mode",
		"game_mode_path":       gameModePath,
		"pawn_blueprint_path":  pawnBlueprintPath,
		"base_class":           baseClass,
	})

	msg, _ := resp["message"].(string)
	if msg != "" {
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}
