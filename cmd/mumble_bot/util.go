package main

import "strings"

// keys returns the keys from a string indexed map.
func keys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// parseCommand splits a command string into the main command and its arguments.
func parseCommand(s string) (cmd string, args []string) {
	parts := strings.Split(s, " ")
	return parts[0], parts[1:]
}
