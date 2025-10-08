package list

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getBackendURL() string {
	if url := os.Getenv("BACKEND_URL"); url != "" {
		return url
	}
	return "http://localhost:3000" // default fallback
}

type BackendGateRepo struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	FullName       string `json:"fullName"`
	HtmlUrl        string `json:"htmlUrl"`
	Description    string `json:"description"`
	Private        bool   `json:"private"`
	OwnerLogin     string `json:"ownerLogin"`
	OwnerAvatarUrl string `json:"ownerAvatarUrl"`
}

type BackendGateListReposResponse struct {
	Repos            []BackendGateRepo `json:"repos,omitempty"`
	Error            string            `json:"error,omitempty"`
	ErrorDescription string            `json:"errorDescription,omitempty"`
}

type BackendGateRepoWithVersions struct {
	Id          int64           `json:"id"`
	OwnerLogin  string          `json:"ownerLogin"`
	RepoName    string          `json:"repoName"`
	RepoId      int64           `json:"repoId"`
	FullName    string          `json:"fullName"`
	HtmlUrl     string          `json:"htmlUrl"`
	Description string          `json:"description"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
	Versions    []SecretVersion `json:"versions"`
}

type SecretVersion struct {
	Version     int    `json:"version"`
	Tag         string `json:"tag"`
	Checksum    string `json:"checksum"`
	UploadedBy  string `json:"uploadedBy"`
	CreatedAt   string `json:"createdAt"`
	IsEncrypted bool   `json:"isEncrypted"`
}

type BackendGateListReposWithVersionsResponse struct {
	Repositories     []BackendGateRepoWithVersions `json:"repositories,omitempty"`
	Error            string                        `json:"error,omitempty"`
	ErrorDescription string                        `json:"errorDescription,omitempty"`
}

type StoredAuthData struct {
	Jwt string `json:"jwt"`
}

func retrieveJwt() string {
	bytes, err := os.ReadFile("./temp/auth.json")
	if err != nil {
		fmt.Println("No auth file found. Please authenticate first using the auth command.")
		os.Exit(1)
	}

	var authData StoredAuthData
	if err := json.Unmarshal(bytes, &authData); err != nil {
		fmt.Println("Error parsing auth file:", err)
		os.Exit(1)
	}

	return authData.Jwt
}

func ListRepos() {
	jwt := retrieveJwt()

	req, err := http.NewRequest("GET", getBackendURL()+"/repos/list", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", "Bearer "+jwt)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	var reposResponse BackendGateListReposResponse
	if err := json.Unmarshal(body, &reposResponse); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if reposResponse.Error != "" {
		fmt.Printf("Error: %s", reposResponse.Error)
		if reposResponse.ErrorDescription != "" {
			fmt.Printf(" - %s", reposResponse.ErrorDescription)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println("Your repositories:")
	for i, repo := range reposResponse.Repos {
		fmt.Printf("%d. %s (%s)\n", i+1, repo.Name, repo.OwnerLogin)
	}
}

func ListReposWithVersions() {
	jwt := retrieveJwt()

	req, err := http.NewRequest("GET", getBackendURL()+"/repos/list-with-versions", nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", "Bearer "+jwt)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	var reposResponse BackendGateListReposWithVersionsResponse
	if err := json.Unmarshal(body, &reposResponse); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if reposResponse.Error != "" {
		fmt.Printf("Error: %s", reposResponse.Error)
		if reposResponse.ErrorDescription != "" {
			fmt.Printf(" - %s", reposResponse.ErrorDescription)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println("Your repositories with versions:")
	for i, repo := range reposResponse.Repositories {
		fmt.Printf("\n%d. %s (%s)\n", i+1, repo.RepoName, repo.OwnerLogin)
		fmt.Printf("   Description: %s\n", repo.Description)
		fmt.Printf("   URL: %s\n", repo.HtmlUrl)
		fmt.Printf("   Created: %s\n", repo.CreatedAt)
		fmt.Printf("   Updated: %s\n", repo.UpdatedAt)

		if len(repo.Versions) > 0 {
			fmt.Printf("   Versions:\n")
			for _, version := range repo.Versions {
				fmt.Printf("     v%d (%s) - %s\n", version.Version, version.Tag, version.CreatedAt)
				fmt.Printf("       Checksum: %s\n", version.Checksum)
			}
		} else {
			fmt.Printf("   No versions uploaded yet\n")
		}
	}
}
