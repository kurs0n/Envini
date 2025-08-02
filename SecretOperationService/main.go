package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/kurs0n/SecretOperationService/internal"
)

func main() {
	// Load environment variables
	if err := loadEnv(); err != nil {
		log.Fatal("Failed to load environment variables:", err)
	}

	// Initialize database
	if err := internal.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Test the new function
	if err := testListAllRepositoriesWithVersions(); err != nil {
		log.Printf("Test failed: %v", err)
	}

	// Start gRPC server
	internal.RunGRPCServer()
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	return nil
}

func testListAllRepositoriesWithVersions() error {
	repos, err := internal.ListAllRepositoriesWithVersions()
	if err != nil {
		return fmt.Errorf("failed to list repositories: %v", err)
	}

	// Print as JSON for easy reading
	jsonData, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	fmt.Println("=== All Repositories with Versions ===")
	fmt.Println(string(jsonData))
	fmt.Println("=====================================")

	return nil
}
