package main

import (
	"io/ioutil"
	"os"
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

func findFile(path string, extensions ...string) (p string, err error) {
	for _, e := range extensions {
		file := path + e
		_, err := os.Stat(file)
		if err == nil {
			return file, nil
		}
		if os.IsNotExist(err) {
			continue
		}
		return "", err
	}
	return "", err
}

func listFiles(path string, extensions ...string) (paths []string, err error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	paths = make([]string, 0, len(files))
	for _, info := range files {
		for _, e := range extensions {
			if ext := filepath.Ext(info.Name()); ext == e {
				paths = append(paths, strings.TrimSuffix(info.Name(), ext))
			}
		}
	}
	return
}
