package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "UnrealHandshake"
	serverVersion = "1.0.0"
	unrealPort    = 9877
)

// execDir is the directory containing this executable, used for locating knowledge_base.
var execDir string

func main() {
	// Determine executable directory for locating knowledge_base
	ex, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine executable path: %v\n", err)
		os.Exit(1)
	}
	execDir = filepath.Dir(ex)

	// Write PID file
	pidPath, err := writePIDFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write PID file: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "MCP Server started with PID file at: %s\n", pidPath)
	}

	// Handle signals for graceful cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		if pidPath != "" {
			cleanupPIDFile(pidPath)
		}
		os.Exit(0)
	}()
	defer func() {
		if pidPath != "" {
			cleanupPIDFile(pidPath)
		}
	}()

	// Create MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Register all 29 tools
	registerAllTools(s)

	// Start stdio transport
	fmt.Fprintf(os.Stderr, "Server starting...\n")
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server crashed with error: %v\n", err)
		os.Exit(1)
	}
}

func registerAllTools(s *server.MCPServer) {
	registerSpecialTools(s)
	registerBasicTools(s)
	registerExecTools(s)
	registerBlueprintTools(s)
	registerActorTools(s)
	registerUITools(s)
	registerGameModeTools(s)
}

func writePIDFile() (string, error) {
	pid := os.Getpid()
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	pidDir := filepath.Join(home, ".unrealgenai")
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return "", err
	}
	pidPath := filepath.Join(pidDir, "mcp_server.pid")
	content := fmt.Sprintf("%d\n%d", pid, unrealPort)
	return pidPath, os.WriteFile(pidPath, []byte(content), 0644)
}

func cleanupPIDFile(path string) {
	os.Remove(path)
}
