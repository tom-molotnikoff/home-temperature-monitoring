package cmd

import (
	"fmt"
	"strings"
)

// parseKVPairs parses a slice of "key=value" strings into a map. Used by the
// sensors and (formerly) other config-bearing commands. Errors when a pair is
// malformed.
func parseKVPairs(pairs []string) (map[string]string, error) {
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config pair: %q (expected key=value)", p)
		}
		out[parts[0]] = parts[1]
	}
	return out, nil
}
