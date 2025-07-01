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
		fmt.Println("Your repos: ")
		if auth.IfRefreshIsRequired() {
			auth.RefreshTokens()
		}
		list.ListRepos()
	case "list":
		if len(os.Args) >= 3 && os.Args[2] == "environments" {
			var repoName, owner string
			for i, arg := range os.Args {
				if arg == "--repo" && i+1 < len(os.Args) {
					repoName = os.Args[i+1]
				}
				if arg == "--owner" && i+1 < len(os.Args) {
					owner = os.Args[i+1]
				}
			}
			if repoName == "" || owner == "" {
				fmt.Println("Please specify a repository name with --repo <repo name> and an owner with --owner <owner>")
				return
			}
		} else {
			fmt.Println("Unknown list command. Use: list environments --repo <repo name> --owner <owner>")
		}
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
