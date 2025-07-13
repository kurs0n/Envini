package list

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const BackendGateURL = "http://localhost:3000"

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

	req, err := http.NewRequest("GET", BackendGateURL+"/repos/list", nil)
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
