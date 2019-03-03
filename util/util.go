package util

import "os"

// SetStringFromEnv sets a string from the environment
func SetStringFromEnv(target *string, env string) {
	if str := os.Getenv(env); str != "" {
		*target = str
	}
}
