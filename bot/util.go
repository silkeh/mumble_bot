package bot

import (
	"fmt"
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

// listFiles lists all files in a directory with a given extension.
func listFiles(path string, extension string) (paths []string, err error) {
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

// intJoin joins a list of integers by a separator.
func intJoin(elems []int, sep string) string {
	strs := make([]string, len(elems))
	for i, v := range elems {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strs, sep)
}
