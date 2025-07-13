package main

import "github.com/kurs0n/AuthService/internal"
import "github.com/joho/godotenv"
import "log"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	internal.RunGRPCServer()
} 