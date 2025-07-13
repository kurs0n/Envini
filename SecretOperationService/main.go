package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kurs0n/SecretOperationService/internal"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	internal.RunGRPCServer()
}
