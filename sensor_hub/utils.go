package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// This function reads a properties file and returns a map of key-value pairs
// It expects the properties file to be in the format: key=value
// It will log a fatal error if it cannot read the file or parse it correctly.
func read_properties_file(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	defer file.Close()

	props := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	return props, nil
}
