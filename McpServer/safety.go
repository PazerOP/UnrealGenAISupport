package main

import (
	"regexp"
	"strings"
)

// destructivePatterns matches the exact same regex patterns as the Python implementation.
var destructivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)unreal\.EditorAssetLibrary\.delete_asset`),
	regexp.MustCompile(`(?i)unreal\.EditorLevelLibrary\.destroy_actor`),
	regexp.MustCompile(`(?i)unreal\.save_package`),
	regexp.MustCompile(`(?i)os\.remove`),
	regexp.MustCompile(`(?i)shutil\.rmtree`),
	regexp.MustCompile(`(?i)file\.write`),
	regexp.MustCompile(`(?i)unreal\.EditorAssetLibrary\.save_asset`),
}

// isPotentiallyDestructive checks a Python script for destructive patterns.
func isPotentiallyDestructive(script string) bool {
	for _, pattern := range destructivePatterns {
		if pattern.MatchString(script) {
			return true
		}
	}
	return false
}

// destructiveCommandKeywords are keywords that indicate destructive Unreal commands.
var destructiveCommandKeywords = []string{"delete", "save", "quit", "exit", "restart"}

// isDestructiveCommand checks Unreal commands for destructive keywords.
func isDestructiveCommand(command string) bool {
	lower := strings.ToLower(command)
	for _, keyword := range destructiveCommandKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
