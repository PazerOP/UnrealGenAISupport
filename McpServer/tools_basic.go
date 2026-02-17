package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBasicTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("handshake_test",
			mcp.WithDescription("Send a handshake message to Unreal Engine"),
			mcp.WithString("message",
				mcp.Required(),
				mcp.Description("Message to send"),
			),
		),
		handleHandshakeTest,
	)

	s.AddTool(
		mcp.NewTool("spawn_object",
			mcp.WithDescription(`Spawn an object in the Unreal Engine level

For basic shapes, use: "Cube", "Sphere", "Cylinder", or "Cone".
For other actors, use class name like "PointLight" or full path.`),
			mcp.WithString("actor_class",
				mcp.Required(),
				mcp.Description("For basic shapes, use: \"Cube\", \"Sphere\", \"Cylinder\", or \"Cone\". For other actors, use class name like \"PointLight\" or full path."),
			),
			mcp.WithArray("location",
				mcp.Description("[X, Y, Z] coordinates"),
				mcp.WithNumberItems(),
			),
			mcp.WithArray("rotation",
				mcp.Description("[Pitch, Yaw, Roll] in degrees"),
				mcp.WithNumberItems(),
			),
			mcp.WithArray("scale",
				mcp.Description("[X, Y, Z] scale factors"),
				mcp.WithNumberItems(),
			),
			mcp.WithString("actor_label",
				mcp.Description("Optional custom name for the actor"),
			),
		),
		handleSpawnObject,
	)

	s.AddTool(
		mcp.NewTool("create_material",
			mcp.WithDescription("Create a new material with the specified color"),
			mcp.WithString("material_name",
				mcp.Required(),
				mcp.Description("Name for the new material"),
			),
			mcp.WithArray("color",
				mcp.Required(),
				mcp.Description("[R, G, B] color values (0-1)"),
				mcp.WithNumberItems(),
			),
		),
		handleCreateMaterial,
	)

	s.AddTool(
		mcp.NewTool("get_all_scene_objects",
			mcp.WithDescription("Retrieve all actors in the current Unreal Engine level."),
		),
		handleGetAllSceneObjects,
	)

	s.AddTool(
		mcp.NewTool("create_project_folder",
			mcp.WithDescription("Create a new folder in the Unreal project content directory."),
			mcp.WithString("folder_path",
				mcp.Required(),
				mcp.Description("Path relative to /Game (e.g., \"FlappyBird/Assets\")"),
			),
		),
		handleCreateProjectFolder,
	)

	s.AddTool(
		mcp.NewTool("get_files_in_folder",
			mcp.WithDescription("List all files in a specified project folder."),
			mcp.WithString("folder_path",
				mcp.Required(),
				mcp.Description("Path relative to /Game (e.g., \"FlappyBird/Assets\")"),
			),
		),
		handleGetFilesInFolder,
	)

	s.AddTool(
		mcp.NewTool("add_input_binding",
			mcp.WithDescription("Add an input action binding to Project Settings."),
			mcp.WithString("action_name",
				mcp.Required(),
				mcp.Description("Name of the action (e.g., \"Flap\")"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Key to bind (e.g., \"Space Bar\")"),
			),
		),
		handleAddInputBinding,
	)
}

func handleHandshakeTest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message := req.GetString("message", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":    "handshake",
		"message": message,
	})

	if success, _ := resp["success"].(bool); success {
		msg, _ := resp["message"].(string)
		return mcp.NewToolResultText(fmt.Sprintf("Handshake successful: %s", msg)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Handshake failed: %s", errMsg)), nil
}

func handleSpawnObject(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actorClass := req.GetString("actor_class", "")
	location := req.GetFloatSlice("location", []float64{0, 0, 0})
	rotation := req.GetFloatSlice("rotation", []float64{0, 0, 0})
	scale := req.GetFloatSlice("scale", []float64{1, 1, 1})
	actorLabel := req.GetString("actor_label", "")

	cmd := map[string]interface{}{
		"type":        "spawn",
		"actor_class": actorClass,
		"location":    location,
		"rotation":    rotation,
		"scale":       scale,
		"actor_label": nil,
	}
	if actorLabel != "" {
		cmd["actor_label"] = actorLabel
	}

	resp := sendToUnreal(cmd)

	if success, _ := resp["success"].(bool); success {
		msg := fmt.Sprintf("Successfully spawned %s", actorClass)
		if actorLabel != "" {
			msg += fmt.Sprintf(" with label '%s'", actorLabel)
		}
		return mcp.NewToolResultText(msg), nil
	}

	errStr, _ := resp["error"].(string)
	if errStr == "" {
		errStr = "Unknown error"
	}
	if strings.Contains(errStr, "not found") {
		errStr += "\nHint: For basic shapes, use 'Cube', 'Sphere', 'Cylinder', or 'Cone'. For other actors, try using '/Script/Engine.PointLight' format."
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to spawn object: %s", errStr)), nil
}

func handleCreateMaterial(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	materialName := req.GetString("material_name", "")
	color := req.GetFloatSlice("color", []float64{})

	resp := sendToUnreal(map[string]interface{}{
		"type":          "create_material",
		"material_name": materialName,
		"color":         color,
	})

	if success, _ := resp["success"].(bool); success {
		matPath, _ := resp["material_path"].(string)
		return mcp.NewToolResultText(fmt.Sprintf("Successfully created material '%s' with path: %s", materialName, matPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to create material: %s", errMsg)), nil
}

func handleGetAllSceneObjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp := sendToUnreal(map[string]interface{}{
		"type": "get_all_scene_objects",
	})

	if success, _ := resp["success"].(bool); success {
		jsonBytes, _ := json.Marshal(resp)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}

func handleCreateProjectFolder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderPath := req.GetString("folder_path", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":        "create_project_folder",
		"folder_path": folderPath,
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

func handleGetFilesInFolder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderPath := req.GetString("folder_path", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":        "get_files_in_folder",
		"folder_path": folderPath,
	})

	if success, _ := resp["success"].(bool); success {
		files := resp["files"]
		jsonBytes, _ := json.Marshal(files)
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}

	errMsg, _ := resp["error"].(string)
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}

func handleAddInputBinding(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actionName := req.GetString("action_name", "")
	key := req.GetString("key", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":        "add_input_binding",
		"action_name": actionName,
		"key":         key,
	})

	msg, _ := resp["message"].(string)
	if msg != "" {
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}
