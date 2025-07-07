package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	Interval        int    `json:"interval"`
}

type OauthAccessTokenSuccessResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiresIn             int64  `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
	TokenType             string `json:"token_type"`
	Scope                 string `json:"scope"`
}

type BadCredentialsResponse struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
	Status           string `json:"status"`
}

type OauthAccessTokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
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

func requestDeviceCode() DeviceCodeResponse {
	uri := "https://github.com/login/device/code"
	data := url.Values{}

	data.Set("client_id", os.Getenv("CLIENT_ID"))
	data.Set("scope", "codespace")

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, uri, strings.NewReader(data.Encode()))
	r.Header.Add("Accept", "application/json")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Wrong response")
	}
	var deviceCodeResponse DeviceCodeResponse
	json.Unmarshal(body, &deviceCodeResponse)
	return deviceCodeResponse
}

func requestToken(deviceCode string) (*OauthAccessTokenSuccessResponse, *OauthAccessTokenErrorResponse) {
	uri := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("CLIENT_ID"))
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, uri, strings.NewReader(data.Encode()))
	r.Header.Add("Accept", "application/json")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Wrong response")
	}

	var oauthAccessTokenErrorResponse OauthAccessTokenErrorResponse
	var oauthAccessTokenSuccessResponse OauthAccessTokenSuccessResponse
	err = json.Unmarshal(body, &oauthAccessTokenErrorResponse)
	if oauthAccessTokenErrorResponse.Error == "" {
		json.Unmarshal(body, &oauthAccessTokenSuccessResponse)
		return &oauthAccessTokenSuccessResponse, nil

	} else if err != nil {
		panic("Something went wrong")
	} else {
		return nil, &oauthAccessTokenErrorResponse
	}
}

func checkForTokens(deviceCodeResponse DeviceCodeResponse, interval int) *OauthAccessTokenSuccessResponse {
	for {
		successResponse, errorResponse := requestToken(deviceCodeResponse.DeviceCode)
		if successResponse != nil {
			return successResponse
		}

		switch errorResponse.Error {
		case "authorization_pending":
			time.Sleep(time.Duration(interval) * time.Second)
		case "slow_down":
			time.Sleep(time.Duration(interval) * time.Second)
		case "expired_token":
			fmt.Println("The device code has expired. Please run `login` again.")
			os.Exit(1)
		case "access_denied":
			fmt.Println("Login cancelled by user.")
			panic("error")
		default:
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func writeTokensToFile(oauthAccessTokenSuccessResponse *OauthAccessTokenSuccessResponse) {
	err := os.MkdirAll("./temp", 0755)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}
	jsonBytes, err := json.Marshal(oauthAccessTokenSuccessResponse)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = os.WriteFile("./temp/tokens.json", jsonBytes, 0644)
	if err != nil {
		fmt.Println("There is no file tokens.json. Please authenticate first using the auth command.")
		return
	}
}

func clearTokens() {
	if _, err := os.Stat("./temp/tokens.json"); err == nil {
		err := os.Remove("./temp/tokens.json")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func retrieveTokens() OauthAccessTokenSuccessResponse {
	bytes, err := os.ReadFile("./temp/tokens.json")
	if err != nil {
		fmt.Println("There is no file tokens.json. Please authenticate first using the auth command.")
		os.Exit(0)
	}
	var data OauthAccessTokenSuccessResponse
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}
	return data
}

func RefreshTokens() { // write additional logic to check whenever refresh_token is expired
	oauthAccessTokenSuccessResponse := retrieveTokens()
	uri := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("CLIENT_ID"))
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", oauthAccessTokenSuccessResponse.RefreshToken)
	data.Set("scope", "codespace")

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, uri, strings.NewReader(data.Encode()))
	r.Header.Add("Accept", "application/json")
	resp, _ := client.Do(r)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Wrong response")
	}
	err = json.Unmarshal(body, &oauthAccessTokenSuccessResponse)
	if err != nil {
		panic("Wrong parsing")
	}
	writeTokensToFile(&oauthAccessTokenSuccessResponse)
}

func IfRefreshIsRequired() bool {
	OauthAccessTokenSuccessResponse := retrieveTokens()
	uri := "https://api.github.com/user/repos"

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodGet, uri, nil)
	r.Header.Add("Accept", "application/vnd.github+json")
	r.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	r.Header.Add("Authorization", "Bearer "+OauthAccessTokenSuccessResponse.AccessToken)

	resp, _ := client.Do(r)
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		panic("Wrong Response")
	}
	var badCredentialsResponse BadCredentialsResponse
	json.Unmarshal(body, &badCredentialsResponse)
	if badCredentialsResponse.Message == "Bad credentials" {
		return true
	}
	return false
}

func GetAccessToken() string {
	return retrieveTokens().AccessToken
}

func Authorize() {
	clearTokens()
	deviceCodeResponse := requestDeviceCode()
	openURL(deviceCodeResponse.VerificationUri)
	fmt.Println("Verification Url: " + deviceCodeResponse.VerificationUri)
	fmt.Println("Your Authorization Code: " + deviceCodeResponse.UserCode)
	oauthAccessTokenSuccessResponse := checkForTokens(deviceCodeResponse, 5)
	writeTokensToFile(oauthAccessTokenSuccessResponse)
}
