package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zsarvas/RL-Discord-Matchmaking/api"
)

func main() {
	port := flag.String("port", "8080", "Port to run the API server on")
	flag.Parse()

	// Load environment variables - try multiple possible locations
	var err error
	envPaths := []string{
		"../../dev.env", // When running from src/cmd/api-server/
		"../dev.env",    // When running from src/
		"dev.env",       // When running from src/ directly
		"./dev.env",     // Current directory
	}

	for _, path := range envPaths {
		if err = godotenv.Load(path); err == nil {
			log.Printf("Loaded environment from %s\n", path)
			break
		}
	}

	if err != nil {
		log.Println("Warning: Could not load dev.env file, using environment variables")
	}

	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connStr == "" {
		log.Fatal("POSTGRES_CONNECTION_STRING environment variable is required")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	// Remove quotes if present
	connStr = strings.Trim(connStr, "\"")
	apiKey = strings.Trim(apiKey, "\"")

	// Initialize API
	apiServer := api.NewAPI(connStr, apiKey)

	// Start API server
	apiServer.StartAPI(*port)
}
