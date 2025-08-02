package internal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	secretsservice "github.com/kurs0n/SecretOperationService/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	secretsservice.UnimplementedSecretsServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) ListRepos(ctx context.Context, req *secretsservice.ListReposRequest) (*secretsservice.ListReposResponse, error) {
	client := &http.Client{}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, os.Getenv("GITHUB_API_URL")+"/user/repos?per_page=100", nil)
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

func (s *Server) UploadSecret(ctx context.Context, req *secretsservice.UploadSecretRequest) (*secretsservice.UploadSecretResponse, error) {
	// Get audit info from context
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to the repository
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to list repos: " + listResp.Error,
		}, nil
	}

	if !HasRepoAccess(listResp.Repos, req.OwnerLogin, req.RepoName) {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "No access to repository")
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "No access to repository",
		}, nil
	}

	// 2. Find the repository in the list to get its details
	var targetRepo *secretsservice.Repo
	for _, repo := range listResp.Repos {
		if repo.OwnerLogin == req.OwnerLogin && repo.Name == req.RepoName {
			targetRepo = repo
			break
		}
	}

	if targetRepo == nil {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Repository not found")
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Repository not found",
		}, nil
	}

	// 3. Parse .env file content
	envData, err := s.parseEnvFile(req.EnvFileContent)
	if err != nil {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to parse .env file: "+err.Error())
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to parse .env file: " + err.Error(),
		}, nil
	}

	// 4. Convert env data to JSON
	envDataJSON, err := json.Marshal(envData)
	if err != nil {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to marshal env data: "+err.Error())
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to marshal env data: " + err.Error(),
		}, nil
	}

	// 5. Calculate checksum
	checksum := s.calculateChecksum(req.EnvFileContent)

	// 6. Get or create repository in database
	repo, err := GetOrCreateRepository(
		req.OwnerLogin,
		req.RepoName,
		targetRepo.Id,
		targetRepo.FullName,
		targetRepo.HtmlUrl,
		targetRepo.Description,
		targetRepo.Private,
	)
	if err != nil {
		LogAuditEvent("UPLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to get/create repository: "+err.Error())
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to get/create repository: " + err.Error(),
		}, nil
	}

	// 7. Get next version number
	version, err := GetNextVersion(repo.ID)
	if err != nil {
		LogAuditEvent("UPLOAD", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to get next version: "+err.Error())
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to get next version: " + err.Error(),
		}, nil
	}

	// 8. Create secret in database (with encryption enabled)
	secret, err := CreateSecret(repo.ID, version, req.Tag, string(envDataJSON), checksum, serviceName, true)
	if err != nil {
		LogAuditEvent("UPLOAD", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to create secret: "+err.Error())
		return &secretsservice.UploadSecretResponse{
			Success: false,
			Error:   "Failed to create secret: " + err.Error(),
		}, nil
	}

	// 9. Log successful operation
	LogAuditEvent("UPLOAD", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.UploadSecretResponse{
		Success:  true,
		Version:  int32(version),
		Checksum: checksum,
	}, nil
}

func (s *Server) ListSecretVersions(ctx context.Context, req *secretsservice.ListSecretVersionsRequest) (*secretsservice.ListSecretVersionsResponse, error) {
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to the repository
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("LIST_VERSIONS", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.ListSecretVersionsResponse{
			Error: "Failed to list repos: " + listResp.Error,
		}, nil
	}

	if !HasRepoAccess(listResp.Repos, req.OwnerLogin, req.RepoName) {
		LogAuditEvent("LIST_VERSIONS", nil, nil, serviceName, requestID, req.UserLogin, false, "No access to repository")
		return &secretsservice.ListSecretVersionsResponse{
			Error: "No access to repository",
		}, nil
	}

	// 2. Get repository from database
	var repo Repository
	result := DB.Where("owner_login = ? AND repo_name = ?", req.OwnerLogin, req.RepoName).First(&repo)
	if result.Error != nil {
		LogAuditEvent("LIST_VERSIONS", nil, nil, serviceName, requestID, req.UserLogin, false, "Repository not found in database")
		return &secretsservice.ListSecretVersionsResponse{
			Error: "Repository not found in database",
		}, nil
	}

	// 3. List secret versions
	secrets, err := ListSecretVersions(repo.ID)
	if err != nil {
		LogAuditEvent("LIST_VERSIONS", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to list secret versions: "+err.Error())
		return &secretsservice.ListSecretVersionsResponse{
			Error: "Failed to list secret versions: " + err.Error(),
		}, nil
	}

	// 4. Convert to proto format
	versions := make([]*secretsservice.SecretVersion, len(secrets))
	for i, secret := range secrets {
		versions[i] = &secretsservice.SecretVersion{
			Version:    int32(secret.Version),
			Tag:        secret.Tag,
			Checksum:   secret.Checksum,
			UploadedBy: secret.UploadedBy,
			CreatedAt:  secret.CreatedAt.Format(time.RFC3339),
		}
	}

	// 5. Log successful operation
	LogAuditEvent("LIST_VERSIONS", &repo.ID, nil, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.ListSecretVersionsResponse{
		Versions: versions,
	}, nil
}

func (s *Server) DownloadSecret(ctx context.Context, req *secretsservice.DownloadSecretRequest) (*secretsservice.DownloadSecretResponse, error) {
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to the repository
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("DOWNLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to list repos: " + listResp.Error,
		}, nil
	}

	if !HasRepoAccess(listResp.Repos, req.OwnerLogin, req.RepoName) {
		LogAuditEvent("DOWNLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "No access to repository")
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "No access to repository",
		}, nil
	}

	// 2. Get repository from database
	var repo Repository
	result := DB.Where("owner_login = ? AND repo_name = ?", req.OwnerLogin, req.RepoName).First(&repo)
	if result.Error != nil {
		LogAuditEvent("DOWNLOAD", nil, nil, serviceName, requestID, req.UserLogin, false, "Repository not found in database")
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Repository not found in database",
		}, nil
	}

	// 3. Get secret
	var secret *Secret
	var err2 error
	if req.Version == 0 {
		// Get latest version
		secret, err2 = GetLatestSecret(repo.ID)
	} else {
		// Get specific version
		secret, err2 = GetSecretByVersion(repo.ID, int(req.Version))
	}

	if err2 != nil {
		LogAuditEvent("DOWNLOAD", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to get secret: "+err2.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to get secret: " + err2.Error(),
		}, nil
	}

	// 4. Decrypt secret data if encrypted
	decryptedData, err := DecryptSecretData(secret)
	if err != nil {
		LogAuditEvent("DOWNLOAD", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, false, "Failed to decrypt secret: "+err.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to decrypt secret: " + err.Error(),
		}, nil
	}

	// 5. Convert JSON back to .env format
	var envData map[string]string
	if err := json.Unmarshal([]byte(decryptedData), &envData); err != nil {
		LogAuditEvent("DOWNLOAD", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, false, "Failed to unmarshal env data: "+err.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to unmarshal env data: " + err.Error(),
		}, nil
	}

	// 6. Convert to .env format
	envContent := s.convertToEnvFormat(envData)

	// 7. Log successful operation
	LogAuditEvent("DOWNLOAD", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.DownloadSecretResponse{
		Success:        true,
		Version:        int32(secret.Version),
		Tag:            secret.Tag,
		EnvFileContent: []byte(envContent),
		Checksum:       secret.Checksum,
		UploadedBy:     secret.UploadedBy,
		CreatedAt:      secret.CreatedAt.Format(time.RFC3339),
		IsEncrypted:    secret.IsEncrypted,
	}, nil
}

func (s *Server) DownloadSecretByTag(ctx context.Context, req *secretsservice.DownloadSecretByTagRequest) (*secretsservice.DownloadSecretResponse, error) {
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to the repository
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("DOWNLOAD_BY_TAG", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to list repos: " + listResp.Error,
		}, nil
	}

	if !HasRepoAccess(listResp.Repos, req.OwnerLogin, req.RepoName) {
		LogAuditEvent("DOWNLOAD_BY_TAG", nil, nil, serviceName, requestID, req.UserLogin, false, "No access to repository")
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "No access to repository",
		}, nil
	}

	// 2. Get repository from database
	var repo Repository
	result := DB.Where("owner_login = ? AND repo_name = ?", req.OwnerLogin, req.RepoName).First(&repo)
	if result.Error != nil {
		LogAuditEvent("DOWNLOAD_BY_TAG", nil, nil, serviceName, requestID, req.UserLogin, false, "Repository not found in database")
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Repository not found in database",
		}, nil
	}

	// 3. Get secret by tag
	secret, err := GetSecretByTag(repo.ID, req.Tag)
	if err != nil {
		LogAuditEvent("DOWNLOAD_BY_TAG", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to get secret by tag: "+err.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to get secret by tag: " + err.Error(),
		}, nil
	}

	// 4. Decrypt secret data if encrypted
	decryptedData, err := DecryptSecretData(secret)
	if err != nil {
		LogAuditEvent("DOWNLOAD_BY_TAG", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, false, "Failed to decrypt secret: "+err.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to decrypt secret: " + err.Error(),
		}, nil
	}

	// 5. Convert JSON back to .env format
	var envData map[string]string
	if err := json.Unmarshal([]byte(decryptedData), &envData); err != nil {
		LogAuditEvent("DOWNLOAD_BY_TAG", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, false, "Failed to unmarshal env data: "+err.Error())
		return &secretsservice.DownloadSecretResponse{
			Success: false,
			Error:   "Failed to unmarshal env data: " + err.Error(),
		}, nil
	}

	// 6. Convert to .env format
	envContent := s.convertToEnvFormat(envData)

	// 7. Log successful download
	LogAuditEvent("DOWNLOAD_BY_TAG", &repo.ID, &secret.ID, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.DownloadSecretResponse{
		Success:        true,
		Version:        int32(secret.Version),
		Tag:            secret.Tag,
		EnvFileContent: []byte(envContent),
		Checksum:       secret.Checksum,
		UploadedBy:     secret.UploadedBy,
		CreatedAt:      secret.CreatedAt.Format(time.RFC3339),
		IsEncrypted:    secret.IsEncrypted,
	}, nil
}

func (s *Server) DeleteSecret(ctx context.Context, req *secretsservice.DeleteSecretRequest) (*secretsservice.DeleteSecretResponse, error) {
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to the repository
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("DELETE", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.DeleteSecretResponse{
			Success: false,
			Error:   "Failed to list repos: " + listResp.Error,
		}, nil
	}

	if !HasRepoAccess(listResp.Repos, req.OwnerLogin, req.RepoName) {
		LogAuditEvent("DELETE", nil, nil, serviceName, requestID, req.UserLogin, false, "No access to repository")
		return &secretsservice.DeleteSecretResponse{
			Success: false,
			Error:   "No access to repository",
		}, nil
	}

	// 2. Get repository from database
	var repo Repository
	result := DB.Where("owner_login = ? AND repo_name = ?", req.OwnerLogin, req.RepoName).First(&repo)
	if result.Error != nil {
		LogAuditEvent("DELETE", nil, nil, serviceName, requestID, req.UserLogin, false, "Repository not found in database")
		return &secretsservice.DeleteSecretResponse{
			Success: false,
			Error:   "Repository not found in database",
		}, nil
	}

	// 3. Delete secrets
	var deletedVersions int
	var err2 error

	if req.Version == 0 {
		// Delete all versions
		err2 = DeleteAllSecrets(repo.ID)
		if err2 != nil {
			LogAuditEvent("DELETE", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to delete all secrets: "+err2.Error())
			return &secretsservice.DeleteSecretResponse{
				Success: false,
				Error:   "Failed to delete all secrets: " + err2.Error(),
			}, nil
		}
		deletedVersions = -1 // Indicates all versions deleted
	} else {
		// Delete specific version
		err2 = DeleteSecret(repo.ID, int(req.Version))
		if err2 != nil {
			LogAuditEvent("DELETE", &repo.ID, nil, serviceName, requestID, req.UserLogin, false, "Failed to delete secret: "+err2.Error())
			return &secretsservice.DeleteSecretResponse{
				Success: false,
				Error:   "Failed to delete secret: " + err2.Error(),
			}, nil
		}
		deletedVersions = 1
	}

	// 4. Log successful operation
	LogAuditEvent("DELETE", &repo.ID, nil, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.DeleteSecretResponse{
		Success:         true,
		DeletedVersions: int32(deletedVersions),
	}, nil
}

func (s *Server) ListAllRepositoriesWithVersions(ctx context.Context, req *secretsservice.ListAllRepositoriesWithVersionsRequest) (*secretsservice.ListAllRepositoriesWithVersionsResponse, error) {
	serviceName, requestID := s.getAuditInfo(ctx)

	// 1. Check if user has access to any repositories
	listResp, err := s.ListRepos(ctx, &secretsservice.ListReposRequest{AccessToken: req.AccessToken})
	if err != nil || listResp.Error != "" {
		LogAuditEvent("LIST_ALL_REPOS", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to list repos: "+listResp.Error)
		return &secretsservice.ListAllRepositoriesWithVersionsResponse{
			Error: "Failed to list repos: " + listResp.Error,
		}, nil
	}

	// 2. Get all repositories with versions from database
	reposWithVersions, err := ListAllRepositoriesWithVersions()
	if err != nil {
		LogAuditEvent("LIST_ALL_REPOS", nil, nil, serviceName, requestID, req.UserLogin, false, "Failed to get repositories with versions: "+err.Error())
		return &secretsservice.ListAllRepositoriesWithVersionsResponse{
			Error: "Failed to get repositories with versions: " + err.Error(),
		}, nil
	}

	// 3. Filter repositories to only include those the user has access to
	var accessibleRepos []*secretsservice.RepositoryWithVersions
	for _, repo := range reposWithVersions {
		// Check if user has access to this repository
		if HasRepoAccess(listResp.Repos, repo.OwnerLogin, repo.RepoName) {
			// Convert versions to proto format
			versions := make([]*secretsservice.SecretVersion, len(repo.Versions))
			for i, version := range repo.Versions {
				versions[i] = &secretsservice.SecretVersion{
					Version:     int32(version.Version),
					Tag:         version.Tag,
					Checksum:    version.Checksum,
					UploadedBy:  version.UploadedBy,
					CreatedAt:   version.CreatedAt.Format(time.RFC3339),
					IsEncrypted: version.IsEncrypted,
				}
			}

			accessibleRepos = append(accessibleRepos, &secretsservice.RepositoryWithVersions{
				Id:          uint32(repo.ID),
				OwnerLogin:  repo.OwnerLogin,
				RepoName:    repo.RepoName,
				RepoId:      repo.RepoID,
				FullName:    repo.FullName,
				HtmlUrl:     repo.HTMLURL,
				Description: repo.Description,
				IsPrivate:   repo.IsPrivate,
				CreatedAt:   repo.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   repo.UpdatedAt.Format(time.RFC3339),
				Versions:    versions,
			})
		}
	}

	// 4. Log successful operation
	LogAuditEvent("LIST_ALL_REPOS", nil, nil, serviceName, requestID, req.UserLogin, true, "")

	return &secretsservice.ListAllRepositoriesWithVersionsResponse{
		Repositories: accessibleRepos,
	}, nil
}

func RunGRPCServer() {
	// Initialize database
	if err := InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	secretsservice.RegisterSecretsServiceServer(grpcServer, NewServer())
	log.Println("gRPC SecretsService server listening on :50053")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func HasRepoAccess(repos []*secretsservice.Repo, ownerLogin, name string) bool {
	for _, repo := range repos {
		if repo.OwnerLogin == ownerLogin && repo.Name == name {
			return true
		}
	}
	return false
}

// Helper functions

func (s *Server) getAuditInfo(ctx context.Context) (serviceName, requestID string) {
	// Extract service name and request ID from context
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if serviceNames := md.Get("service-name"); len(serviceNames) > 0 {
			serviceName = serviceNames[0]
		}
		if requestIDs := md.Get("request-id"); len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}
	}

	// Set defaults if not provided
	if serviceName == "" {
		serviceName = "BackendGate" // Default service name
	}
	if requestID == "" {
		requestID = generateRequestID() // Generate a unique request ID
	}

	return serviceName, requestID
}

func generateRequestID() string {
	// Generate a unique request ID using timestamp and random bytes
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d-%x", timestamp, randomBytes)
}

func (s *Server) parseEnvFile(content []byte) (map[string]string, error) {
	envData := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
			envData[key] = value
		}
	}

	return envData, nil
}

func (s *Server) calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func (s *Server) convertToEnvFormat(envData map[string]string) string {
	var lines []string
	for key, value := range envData {
		// Escape special characters in value
		escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
		lines = append(lines, fmt.Sprintf("%s=\"%s\"", key, escapedValue))
	}
	return strings.Join(lines, "\n")
}
