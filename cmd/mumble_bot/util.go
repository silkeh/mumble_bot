package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
)

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

func listFiles(path, extension string) (paths []string, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	paths = make([]string, 0, len(files))
	for _, info := range files {
		if ext := filepath.Ext(info.Name()); ext == extension {
			paths = append(paths, strings.TrimSuffix(info.Name(), ext))
		}
	}
	return
}
