package secrets

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func getBackendURL() string {
	if url := os.Getenv("BACKEND_URL"); url != "" {
		return url
	}
	return "http://localhost:3000" // default fallback
}

type StoredAuthData struct {
	Jwt string `json:"jwt"`
}

type UploadSecretRequest struct {
	RepoID   int64  `json:"repoId"`
	Tag      string `json:"tag"`
	Content  string `json:"content"`
	Filename string `json:"filename"`
}

type UploadSecretResponse struct {
	Success          bool   `json:"success,omitempty"`
	SecretID         int64  `json:"secretId,omitempty"`
	Version          int    `json:"version,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type DeleteSecretRequest struct {
	RepoID   int64 `json:"repoId"`
	Version  int   `json:"version"`
	SecretID int64 `json:"secretId"`
}

type DeleteSecretResponse struct {
	Success          bool   `json:"success,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type DownloadSecretRequest struct {
	RepoID  int64 `json:"repoId"`
	Version int   `json:"version"`
}

type DownloadSecretResponse struct {
	Content          string `json:"content,omitempty"`
	Filename         string `json:"filename,omitempty"`
	Version          int    `json:"version,omitempty"`
	Tag              string `json:"tag,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type ListSecretVersionsRequest struct {
	RepoID int64 `json:"repoId"`
}

type SecretVersionInfo struct {
	Version     int    `json:"version"`
	Tag         string `json:"tag"`
	Checksum    string `json:"checksum"`
	UploadedBy  string `json:"uploadedBy"`
	CreatedAt   string `json:"createdAt"`
	IsEncrypted bool   `json:"isEncrypted"`
}

type ListSecretVersionsResponse struct {
	Versions         []SecretVersionInfo `json:"versions,omitempty"`
	Error            string              `json:"error,omitempty"`
	ErrorDescription string              `json:"errorDescription,omitempty"`
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

func UploadSecret(ownerLogin string, repoName string, tag string, filePath string) {
	jwt := retrieveJwt()

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file %s: %v\n", filePath, err)
		os.Exit(1)
	}

	// Prepare request - encode content as base64 like WebApp does
	request := map[string]string{
		"tag":            tag,
		"envFileContent": base64.StdEncoding.EncodeToString(content),
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	// Make request
	url := fmt.Sprintf("%s/secrets/upload/%s/%s", getBackendURL(), ownerLogin, repoName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")

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

	var response UploadSecretResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if response.Error != "" {
		fmt.Printf("Error: %s", response.Error)
		if response.ErrorDescription != "" {
			fmt.Printf(" - %s", response.ErrorDescription)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("✅ Secret uploaded successfully!\n")
	fmt.Printf("   Secret ID: %d\n", response.SecretID)
	fmt.Printf("   Version: %d\n", response.Version)
	fmt.Printf("   Tag: %s\n", tag)
}

func DeleteSecret(ownerLogin string, repoName string, version int, tag string) {
	jwt := retrieveJwt()

	// Make request - build URL with version and/or tag parameters like WebApp
	var url string
	params := []string{}

	if version > 0 {
		params = append(params, fmt.Sprintf("version=%d", version))
	}
	if tag != "" {
		params = append(params, fmt.Sprintf("tag=%s", tag))
	}

	// If no specific version or tag provided, default to development tag
	if len(params) == 0 {
		params = append(params, "tag=development")
	}

	url = fmt.Sprintf("%s/secrets/delete/%s/%s?%s", getBackendURL(), ownerLogin, repoName, strings.Join(params, "&"))
	req, err := http.NewRequest("DELETE", url, nil)
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

	var response DeleteSecretResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if response.Error != "" {
		fmt.Printf("Error: %s", response.Error)
		if response.ErrorDescription != "" {
			fmt.Printf(" - %s", response.ErrorDescription)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("✅ Secret version %d deleted successfully!\n", version)
}

func DownloadSecret(ownerLogin string, repoName string, version int, tag string, outputPath string) {
	jwt := retrieveJwt()

	// Make request - build URL with version and/or tag parameters like WebApp
	var url string
	params := []string{}

	if version > 0 {
		params = append(params, fmt.Sprintf("version=%d", version))
	}
	if tag != "" {
		params = append(params, fmt.Sprintf("tag=%s", tag))
	}

	// If no specific version or tag provided, default to development tag
	if len(params) == 0 {
		params = append(params, "tag=development")
	}

	url = fmt.Sprintf("%s/secrets/download/%s/%s?%s", getBackendURL(), ownerLogin, repoName, strings.Join(params, "&"))
	req, err := http.NewRequest("GET", url, nil)
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

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: HTTP %d - %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Get filename from Content-Disposition header
	contentDisposition := resp.Header.Get("Content-Disposition")
	filename := ""
	if contentDisposition != "" {
		// Extract filename from "attachment; filename="filename""
		if strings.Contains(contentDisposition, "filename=") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				filename = strings.Trim(parts[1], `"`)
			}
		}
	}

	// If no filename from header, use default
	if filename == "" {
		filename = fmt.Sprintf("%s-%s-v%d.env", ownerLogin, repoName, version)
	}

	// Use provided output path or filename from header
	if outputPath == "" {
		outputPath = filename
	}

	// Read response body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Write content to file
	err = os.WriteFile(outputPath, content, 0644)
	if err != nil {
		fmt.Printf("Failed to write file %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	// Get metadata from headers
	secretVersion := resp.Header.Get("X-Secret-Version")
	secretTag := resp.Header.Get("X-Secret-Tag")

	fmt.Printf("✅ Secret downloaded successfully!\n")
	fmt.Printf("   Version: %s\n", secretVersion)
	fmt.Printf("   Tag: %s\n", secretTag)
	fmt.Printf("   Saved to: %s\n", outputPath)
}

func ListSecretVersions(ownerLogin string, repoName string) {
	jwt := retrieveJwt()

	// Make request
	url := fmt.Sprintf("%s/secrets/versions/%s/%s", getBackendURL(), ownerLogin, repoName)
	req, err := http.NewRequest("GET", url, nil)
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

	var response ListSecretVersionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Failed to parse response: %v\n", err)
		os.Exit(1)
	}

	if response.Error != "" {
		fmt.Printf("Error: %s", response.Error)
		if response.ErrorDescription != "" {
			fmt.Printf(" - %s", response.ErrorDescription)
		}
		fmt.Println()
		os.Exit(1)
	}

	fmt.Printf("Secret versions for %s/%s:\n", ownerLogin, repoName)
	if len(response.Versions) == 0 {
		fmt.Println("   No versions found")
		return
	}

	for _, version := range response.Versions {
		fmt.Printf("   v%d (%s) - %s\n", version.Version, version.Tag, version.CreatedAt)
		fmt.Printf("     Checksum: %s\n", version.Checksum)
		fmt.Println()
	}
}
