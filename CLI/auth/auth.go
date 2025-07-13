package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const BackendGateURL = "http://localhost:3000"

type BackendGateAuthResponse struct {
	VerificationUri string `json:"verificationUri"`
	UserCode        string `json:"userCode"`
	DeviceCode      string `json:"deviceCode"`
	ExpiresIn       int    `json:"expiresIn"`
	Interval        int    `json:"interval"`
}

type BackendGateTokenResponse struct {
	SessionId        string `json:"sessionId"`
	Jwt              string `json:"jwt"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type BackendGateAuthTokenResponse struct {
	AccessToken      string `json:"accessToken"`
	TokenType        string `json:"tokenType"`
	Scope            string `json:"scope"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"errorDescription,omitempty"`
}

type StoredAuthData struct {
	Jwt string `json:"jwt"`
}

// isWSL checks if the Go program is running inside Windows Subsystem for Linux
func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		if isWSL() {
			cmd = "cmd.exe"
			args = []string{"/c", "start", url}
		} else {
			cmd = "xdg-open"
			args = []string{url}
		}
	}
	if len(args) > 1 {
		args = append(args[:1], append([]string{""}, args[1:]...)...)
	}
	return exec.Command(cmd, args...).Start()
}

func startGitHubAuth() (*BackendGateAuthResponse, error) {
	resp, err := http.Post(BackendGateURL+"/auth/github/start", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start GitHub auth: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var authResponse BackendGateAuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &authResponse, nil
}

func pollForToken(deviceCode string) (*BackendGateTokenResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/auth/github/poll?deviceCode=%s", BackendGateURL, deviceCode))
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var tokenResponse BackendGateTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &tokenResponse, nil
}

func startSpinner(stopChan <-chan struct{}) {
	spinner := []rune{'|', '/', '-', '\\'}
	idx := 0
	for {
		select {
		case <-stopChan:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\rWaiting for authorization... %c", spinner[idx%len(spinner)])
			time.Sleep(120 * time.Millisecond)
			idx++
		}
	}
}

func checkForTokens(authResponse *BackendGateAuthResponse) *BackendGateTokenResponse {
	stopSpinner := make(chan struct{})
	go startSpinner(stopSpinner)
	defer func() {
		close(stopSpinner)
		fmt.Print("\r")
	}()

	for {
		tokenResponse, err := pollForToken(authResponse.DeviceCode)
		if err != nil {
			fmt.Printf("\nError polling for token: %v\n", err)
			time.Sleep(time.Duration(authResponse.Interval) * time.Second)
			continue
		}

		if tokenResponse.Error == "" {
			return tokenResponse
		}

		switch tokenResponse.Error {
		case "authorization_pending":
			time.Sleep(time.Duration(authResponse.Interval) * time.Second)
		case "slow_down":
			time.Sleep(time.Duration(authResponse.Interval) * time.Second)
		case "expired_token":
			fmt.Println("\nThe device code has expired. Please run `auth` again.")
			os.Exit(1)
		case "access_denied":
			fmt.Println("\nLogin cancelled by user.")
			os.Exit(1)
		default:
			fmt.Printf("\nError: %s - %s\n", tokenResponse.Error, tokenResponse.ErrorDescription)
			time.Sleep(time.Duration(authResponse.Interval) * time.Second)
		}
	}
}

func writeJwtToFile(jwt string) {
	err := os.MkdirAll("./temp", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	authData := StoredAuthData{Jwt: jwt}
	jsonBytes, err := json.Marshal(authData)
	if err != nil {
		fmt.Println("Error marshaling JWT:", err)
		return
	}

	err = os.WriteFile("./temp/auth.json", jsonBytes, 0644)
	if err != nil {
		fmt.Println("Error writing auth file:", err)
		return
	}
}

func clearAuth() {
	if _, err := os.Stat("./temp/auth.json"); err == nil {
		err := os.Remove("./temp/auth.json")
		if err != nil {
			fmt.Println("Error removing auth file:", err)
		}
	}
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

func validateSession(jwt string) bool {
	resp, err := http.Get(fmt.Sprintf("%s/auth/validate?jwt=%s", BackendGateURL, jwt))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	var validationResponse struct {
		Valid bool   `json:"valid"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &validationResponse); err != nil {
		return false
	}

	return validationResponse.Valid
}

func getAuthToken(jwt string) (*BackendGateAuthTokenResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/auth/token?jwt=%s", BackendGateURL, jwt))
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var tokenResponse BackendGateAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &tokenResponse, nil
}

func Authorize() {
	clearAuth()

	authResponse, err := startGitHubAuth()
	if err != nil {
		fmt.Printf("Failed to start GitHub auth: %v\n", err)
		os.Exit(1)
	}

	openURL(authResponse.VerificationUri)
	fmt.Println("Verification URL: " + authResponse.VerificationUri)
	fmt.Println("Your Authorization Code: " + authResponse.UserCode)

	tokenResponse := checkForTokens(authResponse)
	writeJwtToFile(tokenResponse.Jwt)

	fmt.Println("Authentication successful!")
}

func IfRefreshIsRequired() bool {
	jwt := retrieveJwt()
	return !validateSession(jwt)
}

func GetJwt() string {
	return retrieveJwt()
}

func GetAccessToken() string {
	jwt := retrieveJwt()
	tokenResponse, err := getAuthToken(jwt)
	if err != nil {
		fmt.Printf("Failed to get access token: %v\n", err)
		os.Exit(1)
	}

	if tokenResponse.Error != "" {
		fmt.Printf("Error getting access token: %s - %s\n", tokenResponse.Error, tokenResponse.ErrorDescription)
		os.Exit(1)
	}

	return tokenResponse.AccessToken
}
