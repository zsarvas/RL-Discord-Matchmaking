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

	// Load environment variables
	err := godotenv.Load("dev.env")
	if err != nil {
		log.Println("Warning: Could not load dev.env file, using environment variables")
	}

	connStr := os.Getenv("POSTGRES_CONNECTION_STRING")
	if connStr == "" {
		log.Fatal("POSTGRES_CONNECTION_STRING environment variable is required")
	}

	// Remove quotes if present
	connStr = strings.Trim(connStr, "\"")

	// Initialize API
	apiServer := api.NewAPI(connStr)

	// Start API server
	apiServer.StartAPI(*port)
}

