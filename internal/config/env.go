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

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse KEY=VALUE format
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])

					// Remove quotes if present
					if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
						value = value[1 : len(value)-1]
					}

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
