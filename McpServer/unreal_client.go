package main

import (
	"encoding/json"
	"fmt"
	"net"
)

const bufferSize = 8192

// sendToUnreal connects to the Unreal Engine socket server, sends a JSON command,
// and returns the parsed JSON response. On any error, returns a map with
// {"success": false, "error": "..."} rather than propagating errors.
func sendToUnreal(command map[string]interface{}) map[string]interface{} {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", unrealPort))
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Connection error: %v", err),
		}
	}
	defer conn.Close()

	// Encode and send
	jsonBytes, err := json.Marshal(command)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("JSON marshal error: %v", err),
		}
	}

	if _, err := conn.Write(jsonBytes); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Socket write error: %v", err),
		}
	}

	// Receive: accumulate until valid JSON
	var accumulated []byte
	buf := make([]byte, bufferSize)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			accumulated = append(accumulated, buf[:n]...)
			// Try to parse accumulated data as JSON
			var result map[string]interface{}
			if jsonErr := json.Unmarshal(accumulated, &result); jsonErr == nil {
				return result
			}
			// If unmarshal failed, keep reading (incomplete JSON)
		}
		if err != nil {
			break
		}
	}

	// Final attempt to parse whatever we have
	if len(accumulated) > 0 {
		var result map[string]interface{}
		if err := json.Unmarshal(accumulated, &result); err == nil {
			return result
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   "No response received from Unreal Engine",
	}
}
