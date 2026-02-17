package main

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerUITools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("add_widget_to_user_widget",
			mcp.WithDescription(`Adds a new widget (like TextBlock, Button, Image, CanvasPanel, VerticalBox) to a User Widget Blueprint.

Widget types are case-sensitive: "TextBlock", "Button", "Image", "CanvasPanel", "VerticalBox", "HorizontalBox", "SizeBox", "Border".`),
			mcp.WithString("user_widget_path",
				mcp.Required(),
				mcp.Description("Path to the User Widget Blueprint (e.g., \"/Game/UI/WBP_MainMenu\")"),
			),
			mcp.WithString("widget_type",
				mcp.Required(),
				mcp.Description("Class name of the widget to add (e.g., \"TextBlock\", \"Button\"). Case-sensitive."),
			),
			mcp.WithString("widget_name",
				mcp.Required(),
				mcp.Description("A unique desired name for the new widget variable (e.g., \"TitleText\", \"StartButton\")"),
			),
			mcp.WithString("parent_widget_name",
				mcp.Description("Optional. Name of an existing Panel widget to attach this new widget to. If empty, attempts to attach to the root or first available CanvasPanel."),
			),
		),
		handleAddWidgetToUserWidget,
	)

	s.AddTool(
		mcp.NewTool("edit_widget_property",
			mcp.WithDescription(`Edits a property of a specific widget within a User Widget Blueprint.

For layout properties controlled by the parent panel, prefix with "Slot." (e.g., "Slot.Position", "Slot.Size", "Slot.Anchors").

Value format examples:
    - Text: '"Hello World!"' (Note: String literal requires inner quotes)
    - Float: '150.0'
    - Boolean: 'true' or 'false'
    - LinearColor: '(R=1.0,G=0.0,B=0.0,A=1.0)'
    - Vector2D: '(X=200.0,Y=50.0)'
    - Anchors: '(Minimum=(X=0.5,Y=0.0),Maximum=(X=0.5,Y=0.0))'
    - Font: "(FontObject=Font'/Engine/EngineFonts/Roboto.Roboto',Size=24)"
    - Texture: "Texture2D'/Game/Textures/MyIcon.MyIcon'"
    - Enum: 'ScaleToFit'`),
			mcp.WithString("user_widget_path",
				mcp.Required(),
				mcp.Description("Path to the User Widget Blueprint"),
			),
			mcp.WithString("widget_name",
				mcp.Required(),
				mcp.Description("The name of the widget whose property you want to change"),
			),
			mcp.WithString("property_name",
				mcp.Required(),
				mcp.Description("The name of the property to edit. Case-sensitive."),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("The new value for the property, formatted as Unreal expects for ImportText"),
			),
		),
		handleEditWidgetProperty,
	)
}

func handleAddWidgetToUserWidget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userWidgetPath := req.GetString("user_widget_path", "")
	widgetType := req.GetString("widget_type", "")
	widgetName := req.GetString("widget_name", "")
	parentWidgetName := req.GetString("parent_widget_name", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":               "add_widget_to_user_widget",
		"user_widget_path":   userWidgetPath,
		"widget_type":        widgetType,
		"widget_name":        widgetName,
		"parent_widget_name": parentWidgetName,
	})

	if success, _ := resp["success"].(bool); success {
		actualName, _ := resp["widget_name"].(string)
		if actualName == "" {
			actualName = widgetName
		}
		msg, _ := resp["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Successfully added widget '%s' of type '%s' to '%s'.", actualName, widgetType, userWidgetPath)
		}
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown C++ error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to add widget: %s", errMsg)), nil
}

func handleEditWidgetProperty(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userWidgetPath := req.GetString("user_widget_path", "")
	widgetName := req.GetString("widget_name", "")
	propertyName := req.GetString("property_name", "")
	value := req.GetString("value", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":             "edit_widget_property",
		"user_widget_path": userWidgetPath,
		"widget_name":      widgetName,
		"property_name":    propertyName,
		"value":            value,
	})

	if success, _ := resp["success"].(bool); success {
		msg, _ := resp["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Successfully set property '%s' on widget '%s'.", propertyName, widgetName)
		}
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown C++ error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to edit widget property: %s", errMsg)), nil
}
