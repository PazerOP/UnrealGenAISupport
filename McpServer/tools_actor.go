package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerActorTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("edit_component_property",
			mcp.WithDescription(`Edit a property of a component in a Blueprint or scene actor.

Capabilities:
    - Set component properties in Blueprints (e.g., StaticMesh, bSimulatePhysics).
    - Modify scene actor components (e.g., position, rotation, scale, material).
    - Supports scalar types (float, int, bool), objects (e.g., materials), and vectors/rotators.
    - Examples:
        - Set a mesh: edit_component_property("/Game/BP", "BirdMesh", "StaticMesh", "'/Engine/BasicShapes/Sphere.Sphere'")
        - Move an actor: edit_component_property("", "RootComponent", "RelativeLocation", "100,200,300", true, "Cube_1")
        - Enable physics: edit_component_property("/Game/BP", "BirdMesh", "bSimulatePhysics", "true")`),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint (e.g., \"/Game/FlappyBird/BP_FlappyBird\") or \"\" for scene actors"),
			),
			mcp.WithString("component_name",
				mcp.Required(),
				mcp.Description("Name of the component (e.g., \"BirdMesh\", \"RootComponent\")"),
			),
			mcp.WithString("property_name",
				mcp.Required(),
				mcp.Description("Name of the property to edit (e.g., \"StaticMesh\", \"RelativeLocation\")"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("New value as a string (e.g., \"'/Engine/BasicShapes/Sphere.Sphere'\", \"100,200,300\")"),
			),
			mcp.WithBoolean("is_scene_actor",
				mcp.Description("If true, edit a component on a scene actor (default: false)"),
			),
			mcp.WithString("actor_name",
				mcp.Description("Name of the actor in the scene (required if is_scene_actor is true)"),
			),
		),
		handleEditComponentProperty,
	)
}

func handleEditComponentProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	componentName := req.GetString("component_name", "")
	propertyName := req.GetString("property_name", "")
	value := req.GetString("value", "")
	isSceneActor := req.GetBool("is_scene_actor", false)
	actorName := req.GetString("actor_name", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "edit_component_property",
		"blueprint_path": blueprintPath,
		"component_name": componentName,
		"property_name":  propertyName,
		"value":          value,
		"is_scene_actor": isSceneActor,
		"actor_name":     actorName,
	})

	if success, _ := resp["success"].(bool); success {
		msg, _ := resp["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Set %s of %s to %s", propertyName, componentName, value)
		}
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	if suggestions, ok := resp["suggestions"]; ok {
		errMsg += fmt.Sprintf("\nSuggestions: %v", suggestions)
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}
