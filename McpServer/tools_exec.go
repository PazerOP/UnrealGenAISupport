package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerExecTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("execute_python_script",
			mcp.WithDescription(`Execute a Python script within Unreal Engine's Python interpreter.

Note:
    This tool sends the script to Unreal Engine, where it is executed via a temporary file using Unreal's internal
    Python execution system (similar to GEngine->Exec). This method is stable but may not handle Blueprint-specific
    APIs as seamlessly as direct Python API calls. For Blueprint manipulation, consider using dedicated tools like
    `+"`add_node_to_blueprint`"+` or ensuring the script uses stable `+"`unreal`"+` module functions. Use this tool for Python
    script execution instead of `+"`execute_unreal_command`"+` with 'py' commands.`),
			mcp.WithString("script",
				mcp.Required(),
				mcp.Description("A string containing the Python code to execute in Unreal Engine"),
			),
		),
		handleExecutePythonScript,
	)

	s.AddTool(
		mcp.NewTool("execute_unreal_command",
			mcp.WithDescription(`Execute an Unreal Engine command-line (CMD) command.

Note:
    This tool executes commands directly in Unreal Engine's command system, similar to the editor's console.
    It is intended for built-in editor commands (e.g., "stat fps", "obj list") and not for running Python scripts.
    Do not use this tool with 'py' commands (e.g., "py script.py"); instead, use `+"`execute_python_script`"+` for Python
    execution, which provides dedicated safety checks and output handling.`),
			mcp.WithString("command",
				mcp.Required(),
				mcp.Description("A string containing the Unreal Engine command to execute (e.g., \"obj list\", \"stat fps\")"),
			),
		),
		handleExecuteUnrealCommand,
	)
}

func handleExecutePythonScript(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	script := req.GetString("script", "")

	if isPotentiallyDestructive(script) {
		return mcp.NewToolResultText(
			"This script appears to involve potentially destructive actions (e.g., deleting or saving files) " +
				"that were not explicitly requested. Please confirm if you want to proceed by saying 'Yes, execute it' " +
				"or modify your request to explicitly allow such actions.",
		), nil
	}

	resp := sendToUnreal(map[string]interface{}{
		"type":   "execute_python",
		"script": script,
	})

	if success, _ := resp["success"].(bool); success {
		output, _ := resp["output"].(string)
		if output == "" {
			output = "No output returned"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Script executed successfully. Output: %s", output)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	output, _ := resp["output"].(string)
	if output != "" {
		errMsg += "\n\nPartial output before error: " + output
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to execute script: %s", errMsg)), nil
}

func handleExecuteUnrealCommand(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command := req.GetString("command", "")

	// Block Python commands
	if strings.HasPrefix(strings.TrimSpace(strings.ToLower(command)), "py ") {
		return mcp.NewToolResultText(
			"Error: Use `execute_python_script` to run Python scripts instead of `execute_unreal_command` with 'py' commands. " +
				"For example, use `execute_python_script(script='your_code_here')` for Python execution.",
		), nil
	}

	// Check for destructive commands
	if isDestructiveCommand(command) {
		return mcp.NewToolResultText(
			"This command appears to involve potentially destructive actions (e.g., deleting or saving). " +
				"Please confirm by saying 'Yes, execute it' or explicitly request such actions.",
		), nil
	}

	resp := sendToUnreal(map[string]interface{}{
		"type":    "execute_unreal_command",
		"command": command,
	})

	if success, _ := resp["success"].(bool); success {
		output, _ := resp["output"].(string)
		if output == "" {
			output = "Command executed with no detailed output returned"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Command '%s' executed successfully. Output: %s", command, output)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to execute command '%s': %s", command, errMsg)), nil
}
