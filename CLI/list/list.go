package list

import (
	"Envini-CLI/auth"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserRepo struct {
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	NodeId   string `json:"node_id"`
	FullName string `json:"full_name"`
	Owner    Owner  `json:"owner"`
	Url      string `json:"html_url"`
}

type Owner struct {
	Login string `json:"login"`
	Url   string `json:"html_url"`
}

func ListRepos() {
	accessToken := auth.GetAccessToken()
	uri := "https://api.github.com/user/repos"

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodGet, uri, nil)
	r.Header.Add("Accept", "application/vnd.github+json")
	r.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	r.Header.Add("Authorization", "Bearer "+accessToken)

	resp, _ := client.Do(r)
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		panic("Wrong Response")
	}
	var userRepos []UserRepo
	json.Unmarshal(body, &userRepos)
	for i, val := range userRepos {
		i++
		fmt.Printf("%d"+"."+" "+val.Name+" "+val.Owner.Login+"\n", i)
	}
}
