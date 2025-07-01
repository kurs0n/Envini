package main

import (
	"Envini-CLI/auth"
	"Envini-CLI/help"
	"Envini-CLI/list"
	"Envini-CLI/upload"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	switch os.Args[1] {
	case "help":
		help.DisplayHelp()
	case "auth":
		fmt.Println("Authorization process enabled!")
		auth.Authorize()
	case "list":
		fmt.Println("Your repos: ")
		if auth.IfRefreshIsRequired() == true {
			auth.RefreshTokens()
		}
		list.ListRepos()
	case "list_environments":
	case "upload":
		upload.UploadFile(os.Args[2])

	}
}

// https://docs.github.com/en/apps/creating-github-apps/writing-code-for-a-github-app/building-a-cli-with-a-github-app#prerequisites
