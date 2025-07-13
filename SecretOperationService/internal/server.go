package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	secretsservice "github.com/kurs0n/SecretOperationService/proto"
	"google.golang.org/grpc"
)

type Server struct {
	secretsservice.UnimplementedSecretsServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ListRepos(ctx context.Context, req *secretsservice.ListReposRequest) (*secretsservice.ListReposResponse, error) {
	client := &http.Client{}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/repos?per_page=100", nil)
	if err != nil {
		return &secretsservice.ListReposResponse{
			Error: fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}

	r.Header.Add("Authorization", "Bearer "+req.AccessToken)
	r.Header.Add("Accept", "application/vnd.github+json")
	r.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(r)
	if err != nil {
		return &secretsservice.ListReposResponse{
			Error: fmt.Sprintf("Failed to make request: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &secretsservice.ListReposResponse{
			Error: fmt.Sprintf("GitHub API returned status %d", resp.StatusCode),
		}, nil
	}

	var githubRepos []struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		HTMLURL     string `json:"html_url"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
		Owner       struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		} `json:"owner"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubRepos); err != nil {
		return &secretsservice.ListReposResponse{
			Error: fmt.Sprintf("Failed to decode response: %v", err),
		}, nil
	}

	repos := make([]*secretsservice.Repo, len(githubRepos))
	for i, repo := range githubRepos {
		repos[i] = &secretsservice.Repo{
			Id:             repo.ID,
			Name:           repo.Name,
			FullName:       repo.FullName,
			HtmlUrl:        repo.HTMLURL,
			Description:    repo.Description,
			Private:        repo.Private,
			OwnerLogin:     repo.Owner.Login,
			OwnerAvatarUrl: repo.Owner.AvatarURL,
		}
	}

	return &secretsservice.ListReposResponse{
		Repos: repos,
	}, nil
}

func RunGRPCServer() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	secretsservice.RegisterSecretsServiceServer(grpcServer, NewServer())
	log.Println("gRPC SecretsService server listening on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
