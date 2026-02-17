package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerBlueprintTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("create_blueprint",
			mcp.WithDescription("Create a new Blueprint class"),
			mcp.WithString("blueprint_name",
				mcp.Required(),
				mcp.Description("Name for the new Blueprint"),
			),
			mcp.WithString("parent_class",
				mcp.Description("Parent class name or path (e.g., \"Actor\", \"/Script/Engine.Actor\")"),
			),
			mcp.WithString("save_path",
				mcp.Description("Path to save the Blueprint asset"),
			),
		),
		handleCreateBlueprint,
	)

	s.AddTool(
		mcp.NewTool("add_component_to_blueprint",
			mcp.WithDescription("Add a component to a Blueprint"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("component_class",
				mcp.Required(),
				mcp.Description("Component class to add (e.g., \"StaticMeshComponent\", \"PointLightComponent\")"),
			),
			mcp.WithString("component_name",
				mcp.Description("Name for the new component (optional)"),
			),
		),
		handleAddComponentToBlueprint,
	)

	s.AddTool(
		mcp.NewTool("add_variable_to_blueprint",
			mcp.WithDescription("Add a variable to a Blueprint"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("variable_name",
				mcp.Required(),
				mcp.Description("Name for the new variable"),
			),
			mcp.WithString("variable_type",
				mcp.Required(),
				mcp.Description("Type of the variable (e.g., \"float\", \"vector\", \"boolean\")"),
			),
			mcp.WithString("default_value",
				mcp.Description("Default value for the variable (optional)"),
			),
			mcp.WithString("category",
				mcp.Description("Category for organizing variables in the Blueprint editor"),
			),
		),
		handleAddVariableToBlueprint,
	)

	s.AddTool(
		mcp.NewTool("add_function_to_blueprint",
			mcp.WithDescription("Add a function to a Blueprint"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_name",
				mcp.Required(),
				mcp.Description("Name for the new function"),
			),
			mcp.WithArray("inputs",
				mcp.Description("List of input parameters [{\"name\": \"param1\", \"type\": \"float\"}, ...]"),
			),
			mcp.WithArray("outputs",
				mcp.Description("List of output parameters [{\"name\": \"return\", \"type\": \"boolean\"}, ...]"),
			),
		),
		handleAddFunctionToBlueprint,
	)

	s.AddTool(
		mcp.NewTool("add_node_to_blueprint",
			mcp.WithDescription(`Add a node to a Blueprint graph

Common supported node types:
    - Basic nodes: "ReturnNode", "FunctionEntry", "Branch", "Sequence"
    - Math operations: "Multiply", "Add", "Subtract", "Divide"
    - Utilities: "PrintString", "Delay", "GetActorLocation", "SetActorLocation"
    - For other functions, try using the exact function name from Blueprints
If the requested node type isn't found, the system will search for alternatives and return suggestions.

IMPORTANT: Space nodes at least 400 units apart horizontally and 300 units vertically.`),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_id",
				mcp.Required(),
				mcp.Description("ID of the function to add the node to"),
			),
			mcp.WithString("node_type",
				mcp.Required(),
				mcp.Description("Type of node to add"),
			),
			mcp.WithArray("node_position",
				mcp.Description("Position of the node in the graph [X, Y]"),
				mcp.WithNumberItems(),
			),
			mcp.WithObject("node_properties",
				mcp.Description("Properties to set on the node (optional)"),
			),
		),
		handleAddNodeToBlueprint,
	)

	s.AddTool(
		mcp.NewTool("get_node_suggestions",
			mcp.WithDescription("Get suggestions for a node type in Unreal Blueprints"),
			mcp.WithString("node_type",
				mcp.Required(),
				mcp.Description("The partial or full node type to get suggestions for"),
			),
		),
		handleGetNodeSuggestions,
	)

	s.AddTool(
		mcp.NewTool("delete_node_from_blueprint",
			mcp.WithDescription("Delete a node from a Blueprint graph"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_id",
				mcp.Required(),
				mcp.Description("ID of the function containing the node"),
			),
			mcp.WithString("node_id",
				mcp.Required(),
				mcp.Description("ID of the node to delete"),
			),
		),
		handleDeleteNodeFromBlueprint,
	)

	s.AddTool(
		mcp.NewTool("get_all_nodes_in_graph",
			mcp.WithDescription("Get all nodes in a Blueprint graph with their positions and types"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_id",
				mcp.Required(),
				mcp.Description("ID of the function to get nodes from"),
			),
		),
		handleGetAllNodesInGraph,
	)

	s.AddTool(
		mcp.NewTool("connect_blueprint_nodes",
			mcp.WithDescription("Connect two nodes in a Blueprint graph"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_id",
				mcp.Required(),
				mcp.Description("ID of the function containing the nodes"),
			),
			mcp.WithString("source_node_id",
				mcp.Required(),
				mcp.Description("ID of the source node"),
			),
			mcp.WithString("source_pin",
				mcp.Required(),
				mcp.Description("Name of the source pin"),
			),
			mcp.WithString("target_node_id",
				mcp.Required(),
				mcp.Description("ID of the target node"),
			),
			mcp.WithString("target_pin",
				mcp.Required(),
				mcp.Description("Name of the target pin"),
			),
		),
		handleConnectBlueprintNodes,
	)

	s.AddTool(
		mcp.NewTool("compile_blueprint",
			mcp.WithDescription("Compile a Blueprint"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
		),
		handleCompileBlueprint,
	)

	s.AddTool(
		mcp.NewTool("spawn_blueprint_actor",
			mcp.WithDescription("Spawn a Blueprint actor in the level"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
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
		handleSpawnBlueprintActor,
	)

	s.AddTool(
		mcp.NewTool("add_component_with_events",
			mcp.WithDescription("Add a component to a Blueprint with overlap events if applicable."),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint (e.g., \"/Game/FlappyBird/BP_FlappyBird\")"),
			),
			mcp.WithString("component_name",
				mcp.Required(),
				mcp.Description("Name of the new component (e.g., \"TriggerBox\")"),
			),
			mcp.WithString("component_class",
				mcp.Required(),
				mcp.Description("Class of the component (e.g., \"BoxComponent\")"),
			),
		),
		handleAddComponentWithEvents,
	)

	s.AddTool(
		mcp.NewTool("connect_blueprint_nodes_bulk",
			mcp.WithDescription("Connect multiple pairs of nodes in a Blueprint graph"),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("function_id",
				mcp.Required(),
				mcp.Description("ID of the function containing the nodes"),
			),
			mcp.WithArray("connections",
				mcp.Required(),
				mcp.Description("Array of connection definitions, each containing: source_node_id, source_pin, target_node_id, target_pin"),
			),
		),
		handleConnectBlueprintNodesBulk,
	)

	s.AddTool(
		mcp.NewTool("get_blueprint_node_guid",
			mcp.WithDescription("Retrieve the GUID of a pre-existing node in a Blueprint graph."),
			mcp.WithString("blueprint_path",
				mcp.Required(),
				mcp.Description("Path to the Blueprint asset"),
			),
			mcp.WithString("graph_type",
				mcp.Description("Type of graph to query (\"EventGraph\" or \"FunctionGraph\", default: \"EventGraph\")"),
			),
			mcp.WithString("node_name",
				mcp.Description("Name of the node to find (e.g., \"BeginPlay\" for EventGraph)"),
			),
			mcp.WithString("function_id",
				mcp.Description("ID of the function (used with graph_type=\"FunctionGraph\")"),
			),
		),
		handleGetBlueprintNodeGUID,
	)
}

func handleCreateBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintName := req.GetString("blueprint_name", "")
	parentClass := req.GetString("parent_class", "Actor")
	savePath := req.GetString("save_path", "/Game/Blueprints")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "create_blueprint",
		"blueprint_name": blueprintName,
		"parent_class":   parentClass,
		"save_path":      savePath,
	})

	if success, _ := resp["success"].(bool); success {
		bpPath, _ := resp["blueprint_path"].(string)
		if bpPath == "" {
			bpPath = savePath + "/" + blueprintName
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully created Blueprint '%s' with path: %s", blueprintName, bpPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to create Blueprint: %s", errMsg)), nil
}

func handleAddComponentToBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	componentClass := req.GetString("component_class", "")
	componentName := req.GetString("component_name", "")

	cmd := map[string]interface{}{
		"type":            "add_component",
		"blueprint_path":  blueprintPath,
		"component_class": componentClass,
		"component_name":  nil,
	}
	if componentName != "" {
		cmd["component_name"] = componentName
	}

	resp := sendToUnreal(cmd)

	if success, _ := resp["success"].(bool); success {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully added %s to Blueprint at %s", componentClass, blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to add component: %s", errMsg)), nil
}

func handleAddVariableToBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	variableName := req.GetString("variable_name", "")
	variableType := req.GetString("variable_type", "")
	defaultValue := req.GetString("default_value", "")
	category := req.GetString("category", "Default")

	cmd := map[string]interface{}{
		"type":           "add_variable",
		"blueprint_path": blueprintPath,
		"variable_name":  variableName,
		"variable_type":  variableType,
		"default_value":  nil,
		"category":       category,
	}
	if defaultValue != "" {
		cmd["default_value"] = defaultValue
	}

	resp := sendToUnreal(cmd)

	if success, _ := resp["success"].(bool); success {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully added %s variable '%s' to Blueprint at %s", variableType, variableName, blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to add variable: %s", errMsg)), nil
}

func handleAddFunctionToBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionName := req.GetString("function_name", "")

	// Extract inputs and outputs from raw arguments
	args := req.GetArguments()
	inputs := []interface{}{}
	outputs := []interface{}{}
	if rawInputs, ok := args["inputs"].([]interface{}); ok {
		inputs = rawInputs
	}
	if rawOutputs, ok := args["outputs"].([]interface{}); ok {
		outputs = rawOutputs
	}

	resp := sendToUnreal(map[string]interface{}{
		"type":           "add_function",
		"blueprint_path": blueprintPath,
		"function_name":  functionName,
		"inputs":         inputs,
		"outputs":        outputs,
	})

	if success, _ := resp["success"].(bool); success {
		funcID, _ := resp["function_id"].(string)
		if funcID == "" {
			funcID = "unknown"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully added function '%s' to Blueprint at %s with ID: %s", functionName, blueprintPath, funcID)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to add function: %s", errMsg)), nil
}

func handleAddNodeToBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionID := req.GetString("function_id", "")
	nodeType := req.GetString("node_type", "")
	nodePosition := req.GetFloatSlice("node_position", []float64{0, 0})

	// Extract node_properties from raw arguments
	args := req.GetArguments()
	nodeProperties := map[string]interface{}{}
	if rawProps, ok := args["node_properties"].(map[string]interface{}); ok {
		nodeProperties = rawProps
	}

	resp := sendToUnreal(map[string]interface{}{
		"type":            "add_node",
		"blueprint_path":  blueprintPath,
		"function_id":     functionID,
		"node_type":       nodeType,
		"node_position":   nodePosition,
		"node_properties": nodeProperties,
	})

	if success, _ := resp["success"].(bool); success {
		nodeID, _ := resp["node_id"].(string)
		if nodeID == "" {
			nodeID = "unknown"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Successfully added %s node to function %s in Blueprint at %s with ID: %s", nodeType, functionID, blueprintPath, nodeID)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to add node: %s", errMsg)), nil
}

func handleGetNodeSuggestions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeType := req.GetString("node_type", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":      "get_node_suggestions",
		"node_type": nodeType,
	})

	if success, _ := resp["success"].(bool); success {
		suggestions, _ := resp["suggestions"].([]interface{})
		if len(suggestions) > 0 {
			strs := make([]string, len(suggestions))
			for i, s := range suggestions {
				strs[i] = fmt.Sprintf("%v", s)
			}
			return mcp.NewToolResultText(fmt.Sprintf("Suggestions for '%s': %s", nodeType, joinStrings(strs, ", "))), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("No suggestions found for '%s'", nodeType)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to get suggestions for '%s': %s", nodeType, errMsg)), nil
}

func handleDeleteNodeFromBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionID := req.GetString("function_id", "")
	nodeID := req.GetString("node_id", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "delete_node",
		"blueprint_path": blueprintPath,
		"function_id":    functionID,
		"node_id":        nodeID,
	})

	if success, _ := resp["success"].(bool); success {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted node %s from function %s in Blueprint at %s", nodeID, functionID, blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to delete node: %s", errMsg)), nil
}

func handleGetAllNodesInGraph(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionID := req.GetString("function_id", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "get_all_nodes",
		"blueprint_path": blueprintPath,
		"function_id":    functionID,
	})

	if success, _ := resp["success"].(bool); success {
		nodes, _ := resp["nodes"].(string)
		if nodes == "" {
			nodes = "[]"
		}
		return mcp.NewToolResultText(nodes), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to get nodes: %s", errMsg)), nil
}

func handleConnectBlueprintNodes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionID := req.GetString("function_id", "")
	sourceNodeID := req.GetString("source_node_id", "")
	sourcePin := req.GetString("source_pin", "")
	targetNodeID := req.GetString("target_node_id", "")
	targetPin := req.GetString("target_pin", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "connect_nodes",
		"blueprint_path": blueprintPath,
		"function_id":    functionID,
		"source_node_id": sourceNodeID,
		"source_pin":     sourcePin,
		"target_node_id": targetNodeID,
		"target_pin":     targetPin,
	})

	if success, _ := resp["success"].(bool); success {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully connected %s.%s to %s.%s in Blueprint at %s",
			sourceNodeID, sourcePin, targetNodeID, targetPin, blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}

	// Include available pins information if provided
	if srcPins, ok := resp["source_available_pins"]; ok {
		srcJSON, _ := json.MarshalIndent(srcPins, "", "  ")
		errMsg += fmt.Sprintf("\nAvailable pins on source (%s): %s", sourceNodeID, string(srcJSON))
	}
	if tgtPins, ok := resp["target_available_pins"]; ok {
		tgtJSON, _ := json.MarshalIndent(tgtPins, "", "  ")
		errMsg += fmt.Sprintf("\nAvailable pins on target (%s): %s", targetNodeID, string(tgtJSON))
	}

	return mcp.NewToolResultText(fmt.Sprintf("Failed to connect nodes: %s", errMsg)), nil
}

func handleCompileBlueprint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "compile_blueprint",
		"blueprint_path": blueprintPath,
	})

	if success, _ := resp["success"].(bool); success {
		return mcp.NewToolResultText(fmt.Sprintf("Successfully compiled Blueprint at %s", blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to compile Blueprint: %s", errMsg)), nil
}

func handleSpawnBlueprintActor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	location := req.GetFloatSlice("location", []float64{0, 0, 0})
	rotation := req.GetFloatSlice("rotation", []float64{0, 0, 0})
	scale := req.GetFloatSlice("scale", []float64{1, 1, 1})
	actorLabel := req.GetString("actor_label", "")

	cmd := map[string]interface{}{
		"type":           "spawn_blueprint",
		"blueprint_path": blueprintPath,
		"location":       location,
		"rotation":       rotation,
		"scale":          scale,
		"actor_label":    nil,
	}
	if actorLabel != "" {
		cmd["actor_label"] = actorLabel
	}

	resp := sendToUnreal(cmd)

	if success, _ := resp["success"].(bool); success {
		msg := fmt.Sprintf("Successfully spawned Blueprint %s", blueprintPath)
		if actorLabel != "" {
			msg += fmt.Sprintf(" with label '%s'", actorLabel)
		}
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to spawn Blueprint: %s", errMsg)), nil
}

func handleAddComponentWithEvents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	componentName := req.GetString("component_name", "")
	componentClass := req.GetString("component_class", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":            "add_component_with_events",
		"blueprint_path":  blueprintPath,
		"component_name":  componentName,
		"component_class": componentClass,
	})

	if success, _ := resp["success"].(bool); success {
		msg, _ := resp["message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("Added component %s", componentName)
		}
		// Parse events if present
		if eventsStr, ok := resp["events"].(string); ok && eventsStr != "" {
			var events map[string]interface{}
			if err := json.Unmarshal([]byte(eventsStr), &events); err == nil {
				beginGUID := events["begin_guid"]
				endGUID := events["end_guid"]
				if beginGUID != nil || endGUID != nil {
					msg += fmt.Sprintf("\nOverlap Events - Begin GUID: %v, End GUID: %v", beginGUID, endGUID)
				}
			}
		}
		return mcp.NewToolResultText(msg), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed: %s", errMsg)), nil
}

func handleConnectBlueprintNodesBulk(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	functionID := req.GetString("function_id", "")

	// Extract connections from raw arguments
	args := req.GetArguments()
	connections := []interface{}{}
	if rawConns, ok := args["connections"].([]interface{}); ok {
		connections = rawConns
	}

	resp := sendToUnreal(map[string]interface{}{
		"type":           "connect_nodes_bulk",
		"blueprint_path": blueprintPath,
		"function_id":    functionID,
		"connections":    connections,
	})

	if success, _ := resp["success"].(bool); success {
		successful := resp["successful_connections"]
		total := resp["total_connections"]
		return mcp.NewToolResultText(fmt.Sprintf("Successfully connected %v/%v node pairs in Blueprint at %s", successful, total, blueprintPath)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}

	// Extract detailed error info
	var failedConnections []string
	if results, ok := resp["results"].([]interface{}); ok {
		for _, r := range results {
			result, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			if succ, _ := result["success"].(bool); !succ {
				idx := result["connection_index"]
				src := result["source_node"]
				tgt := result["target_node"]
				err := result["error"]
				failedConnections = append(failedConnections,
					fmt.Sprintf("Connection %v: %v to %v - %v", idx, src, tgt, err))
			}
		}
	}

	if len(failedConnections) > 0 {
		errMsg += "\n- " + joinStrings(failedConnections, "\n- ")
	}

	return mcp.NewToolResultText(fmt.Sprintf("Failed to connect nodes: %s", errMsg)), nil
}

func handleGetBlueprintNodeGUID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	blueprintPath := req.GetString("blueprint_path", "")
	graphType := req.GetString("graph_type", "EventGraph")
	nodeName := req.GetString("node_name", "")
	functionID := req.GetString("function_id", "")

	resp := sendToUnreal(map[string]interface{}{
		"type":           "get_node_guid",
		"blueprint_path": blueprintPath,
		"graph_type":     graphType,
		"node_name":      nodeName,
		"function_id":    functionID,
	})

	if success, _ := resp["success"].(bool); success {
		guid, _ := resp["node_guid"].(string)
		nameOrEntry := nodeName
		if nameOrEntry == "" {
			nameOrEntry = "FunctionEntry"
		}
		return mcp.NewToolResultText(fmt.Sprintf("Node GUID for %s in %s of %s: %s", nameOrEntry, graphType, blueprintPath, guid)), nil
	}

	errMsg, _ := resp["error"].(string)
	if errMsg == "" {
		errMsg = "Unknown error"
	}
	return mcp.NewToolResultText(fmt.Sprintf("Failed to get node GUID: %s", errMsg)), nil
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
