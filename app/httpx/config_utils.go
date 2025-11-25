package httpx

import (
	"os"
	"strings"
)

// ExpandEnvWithDefaults expands environment variables in a string, supporting ${VAR} and ${VAR:-default} syntax
// This allows config files to use environment variable substitution with optional default values
func ExpandEnvWithDefaults(s string) string {
	return os.Expand(s, func(key string) string {
		// Check for ${VAR:-default} syntax
		if idx := strings.Index(key, ":-"); idx != -1 {
			varName := key[:idx]
			defaultValue := key[idx+2:]
			if val := os.Getenv(varName); val != "" {
				return val
			}
			return defaultValue
		}
		// Simple ${VAR} syntax
		return os.Getenv(key)
	})
}

