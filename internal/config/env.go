package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func LoadEnvFile() {
	// Simple .env file loader
	if _, err := os.Stat(".env"); err == nil {
		file, err := os.Open(".env")
		if err != nil {
			log.Printf("Warning: Could not open .env file: %v", err)
			return
		}
		defer file.Close()

		log.Println("Loading .env file...")

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Skip empty lines
			if line == "" {
				continue
			}

			// Skip lines that start with # (full-line comments)
			if strings.HasPrefix(line, "#") {
				continue
			}

			// Remove inline comments (everything after # that's not inside quotes)
			// Simple approach: find first # that's not inside quotes
			commentIdx := -1
			inQuotes := false
			for i, char := range line {
				if char == '"' {
					inQuotes = !inQuotes
				} else if char == '#' && !inQuotes {
					commentIdx = i
					break
				}
			}
			if commentIdx >= 0 {
				line = strings.TrimSpace(line[:commentIdx])
			}

			// Parse KEY=VALUE format
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					// Remove quotes if present (both single and double quotes)
					if len(value) >= 2 {
						if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
							(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
							value = value[1 : len(value)-1]
						}
					}

					// Trim any remaining whitespace
					value = strings.TrimSpace(value)

					os.Setenv(key, value)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading .env file: %v", err)
		}
	} else {
		log.Println("No .env file found, using system environment variables")
	}
}
