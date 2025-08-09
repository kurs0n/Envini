package main

import (
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
