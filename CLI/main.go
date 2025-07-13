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
	if len(os.Args) < 2 {
		help.DisplayHelp()
		return
	}

	switch os.Args[1] {
	case "help":
		help.DisplayHelp()
	case "auth":
		fmt.Println("Authorization process enabled!")
		auth.Authorize()
	case "repos":
		if auth.IfRefreshIsRequired() {
			fmt.Println("Session expired. Please run `auth` again.")
			os.Exit(1)
		}
		list.ListRepos()
	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("Please specify a file to upload.")
			return
		}
		upload.UploadFile(os.Args[2])
	default:
		help.DisplayHelp()
	}
}
