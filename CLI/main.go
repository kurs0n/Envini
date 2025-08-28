package main

import (
	"Envini-CLI/auth"
	"Envini-CLI/help"
	"Envini-CLI/list"
	"Envini-CLI/secrets"
	"Envini-CLI/upload"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

func parseFlags(args []string) map[string]string {
	flags := make(map[string]string)

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			if len(parts) == 2 {
				flags[parts[0]] = parts[1]
			} else {
				flags[parts[0]] = "true"
			}
		} else if strings.HasPrefix(arg, "-") {
			parts := strings.SplitN(arg[1:], "=", 2)
			if len(parts) == 2 {
				flags[parts[0]] = parts[1]
			} else {
				flags[parts[0]] = "true"
			}
		}
	}

	return flags
}

func getNonFlagArgs(args []string) []string {
	var nonFlagArgs []string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}
	return nonFlagArgs
}

func getGitRepoInfo() (string, string, error) {
	// Get remote origin URL
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remote: %v", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse GitHub URL to extract owner and repo
	// Handle both HTTPS and SSH formats
	var owner, repo string

	if strings.Contains(remoteURL, "github.com") {
		// HTTPS format: https://github.com/owner/repo.git
		parts := strings.Split(remoteURL, "/")
		if len(parts) >= 2 {
			repo = strings.TrimSuffix(parts[len(parts)-1], ".git")
			owner = parts[len(parts)-2]
		}
	} else if strings.Contains(remoteURL, "git@github.com") {
		// SSH format: git@github.com:owner/repo.git
		parts := strings.Split(remoteURL, ":")
		if len(parts) >= 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				repo = strings.TrimSuffix(pathParts[len(pathParts)-1], ".git")
				owner = pathParts[len(pathParts)-2]
			}
		}
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("could not parse git remote URL: %s", remoteURL)
	}

	return owner, repo, nil
}

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
	case "repos-with-versions":
		if auth.IfRefreshIsRequired() {
			fmt.Println("Session expired. Please run `auth` again.")
			os.Exit(1)
		}
		list.ListReposWithVersions()
	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("Please specify a file to upload.")
			return
		}

		// Parse flags and non-flag arguments
		flags := parseFlags(os.Args[2:])
		nonFlagArgs := getNonFlagArgs(os.Args[2:])

		// Check if this is explicit repository format (has owner and repo) or git-auto-detect
		if len(nonFlagArgs) >= 3 {
			// Explicit repository format: upload <owner> <repo> <file> [--tag=development]
			if len(nonFlagArgs) < 3 {
				fmt.Println("Usage: envini upload <owner> <repo> <file> [--tag=development]")
				fmt.Println("Example: envini upload kurs0n 8080-emulator .env --tag=production")
				return
			}
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			ownerLogin := nonFlagArgs[0]
			repoName := nonFlagArgs[1]
			filePath := nonFlagArgs[2]
			tag := flags["tag"]
			if tag == "" {
				tag = "development" // Default tag
			}

			secrets.UploadSecret(ownerLogin, repoName, tag, filePath)
		} else {
			// Git-auto-detect format: upload <file> [--tag=development]
			if len(nonFlagArgs) < 1 {
				fmt.Println("Please specify a file to upload.")
				return
			}

			// Try to detect git repository and use it as defaults
			owner, repo, err := getGitRepoInfo()
			if err != nil {
				// Fall back to legacy upload if git detection fails
				fmt.Printf("Warning: Could not detect git repository (%v), using legacy upload\n", err)
				upload.UploadFile(os.Args[2])
				return
			}

			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			filePath := nonFlagArgs[0]
			tag := flags["tag"]
			if tag == "" {
				tag = "development" // Default tag
			}

			fmt.Printf("📁 Detected repository: %s/%s\n", owner, repo)
			fmt.Printf("📄 Uploading: %s\n", filePath)
			fmt.Printf("🏷️  Tag: %s\n", tag)

			secrets.UploadSecret(owner, repo, tag, filePath)
		}
	case "download":
		flags := parseFlags(os.Args[2:])
		nonFlagArgs := getNonFlagArgs(os.Args[2:])

		if len(nonFlagArgs) < 2 {
			// Try to use git repository as defaults
			owner, repo, err := getGitRepoInfo()
			if err != nil {
				fmt.Println("Usage: envini download <owner> <repo> [output-file] [--version=latest]")
				fmt.Println("Example: envini download kurs0n 8080-emulator .env.downloaded --version=1")
				return
			}

			// Use git defaults
			if len(nonFlagArgs) < 1 {
				fmt.Println("Please specify an output file.")
				return
			}

			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			outputPath := nonFlagArgs[0]
			versionStr := flags["version"]
			tag := flags["tag"]
			version := 0 // Default to latest version
			if versionStr != "" && versionStr != "latest" {
				var err error
				version, err = strconv.Atoi(versionStr)
				if err != nil {
					fmt.Printf("Invalid version: %s\n", versionStr)
					return
				}
			}

			fmt.Printf("📁 Detected repository: %s/%s\n", owner, repo)
			fmt.Printf("💾 Downloading to: %s\n", outputPath)
			if version > 0 {
				fmt.Printf("📋 Version: %d\n", version)
			} else if tag != "" {
				fmt.Printf("🏷️  Tag: %s\n", tag)
			} else {
				fmt.Printf("📋 Version: latest (development tag)\n")
			}

			secrets.DownloadSecret(owner, repo, version, tag, outputPath)
		} else {
			// Explicit owner/repo provided
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			ownerLogin := nonFlagArgs[0]
			repoName := nonFlagArgs[1]
			outputPath := ""
			if len(nonFlagArgs) > 2 {
				outputPath = nonFlagArgs[2]
			}

			versionStr := flags["version"]
			tag := flags["tag"]
			version := 0 // Default to latest version
			if versionStr != "" && versionStr != "latest" {
				var err error
				version, err = strconv.Atoi(versionStr)
				if err != nil {
					fmt.Printf("Invalid version: %s\n", versionStr)
					return
				}
			}

			secrets.DownloadSecret(ownerLogin, repoName, version, tag, outputPath)
		}
	case "delete":
		flags := parseFlags(os.Args[2:])
		nonFlagArgs := getNonFlagArgs(os.Args[2:])

		if len(nonFlagArgs) < 2 {
			// Try to use git repository as defaults
			owner, repo, err := getGitRepoInfo()
			if err != nil {
				fmt.Println("Usage: envini delete <owner> <repo> [--version=latest]")
				fmt.Println("Example: envini delete kurs0n 8080-emulator --version=1")
				return
			}

			// Use git defaults
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			versionStr := flags["version"]
			tag := flags["tag"]
			version := 0 // Default to latest version
			if versionStr != "" && versionStr != "latest" {
				var err error
				version, err = strconv.Atoi(versionStr)
				if err != nil {
					fmt.Printf("Invalid version: %s\n", versionStr)
					return
				}
			}

			fmt.Printf("📁 Detected repository: %s/%s\n", owner, repo)
			if version > 0 {
				fmt.Printf("🗑️  Deleting version: %d\n", version)
			} else if tag != "" {
				fmt.Printf("🗑️  Deleting latest from tag: %s\n", tag)
			} else {
				fmt.Printf("🗑️  Deleting latest version (development tag)\n")
			}

			secrets.DeleteSecret(owner, repo, version, tag)
		} else {
			// Explicit owner/repo provided
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			ownerLogin := nonFlagArgs[0]
			repoName := nonFlagArgs[1]

			versionStr := flags["version"]
			tag := flags["tag"]
			version := 0 // Default to latest version
			if versionStr != "" && versionStr != "latest" {
				var err error
				version, err = strconv.Atoi(versionStr)
				if err != nil {
					fmt.Printf("Invalid version: %s\n", versionStr)
					return
				}
			}

			secrets.DeleteSecret(ownerLogin, repoName, version, tag)
		}
	case "versions":
		nonFlagArgs := getNonFlagArgs(os.Args[2:])

		if len(nonFlagArgs) < 2 {
			// Try to use git repository as defaults
			owner, repo, err := getGitRepoInfo()
			if err != nil {
				fmt.Println("Usage: envini versions <owner> <repo>")
				fmt.Println("Example: envini versions kurs0n 8080-emulator")
				return
			}

			// Use git defaults
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			fmt.Printf("📁 Detected repository: %s/%s\n", owner, repo)
			secrets.ListSecretVersions(owner, repo)
		} else {
			// Explicit owner/repo provided
			if auth.IfRefreshIsRequired() {
				fmt.Println("Session expired. Please run `auth` again.")
				os.Exit(1)
			}

			ownerLogin := nonFlagArgs[0]
			repoName := nonFlagArgs[1]
			secrets.ListSecretVersions(ownerLogin, repoName)
		}
	default:
		help.DisplayHelp()
	}
}
