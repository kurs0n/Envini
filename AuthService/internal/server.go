package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"bytes"
	"io"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	authservice "github.com/kurs0n/AuthService/proto"
	"google.golang.org/grpc"
)

type Server struct {
	authservice.UnimplementedAuthServiceServer
	Sessions *SessionStore
}

func NewServer() *Server {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "envini"),
		getenv("DB_PASSWORD", "envini"),
		getenv("DB_NAME", "envini"),
		getenv("DB_SSL_MODE", "disable"),
	)
	store, err := NewSessionStore(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return &Server{Sessions: store}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		_, err := fmt.Sscanf(v, "%d", &i)
		if err == nil {
			return i
		}
	}
	return fallback
}

func (s *Server) StartDeviceFlow(ctx context.Context, req *authservice.StartDeviceFlowRequest) (*authservice.StartDeviceFlowResponse, error) {
	uri := "https://github.com/login/device/code"
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))

	client := &http.Client{}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var deviceCodeResp struct {
		DeviceCode      string `json:"device_code"`
		ExpiresIn       int32  `json:"expires_in"`
		UserCode        string `json:"user_code"`
		VerificationUri string `json:"verification_uri"`
		Interval        int32  `json:"interval"`
	}
	if err := json.Unmarshal(body, &deviceCodeResp); err != nil {
		return nil, err
	}
	return &authservice.StartDeviceFlowResponse{
		DeviceCode:      deviceCodeResp.DeviceCode,
		UserCode:        deviceCodeResp.UserCode,
		VerificationUri: deviceCodeResp.VerificationUri,
		ExpiresIn:       deviceCodeResp.ExpiresIn,
		Interval:        deviceCodeResp.Interval,
	}, nil
}

var jwtSecret = []byte(getenv("JWT_SECRET", "supersecretkey"))

func generateJWT(sessionID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"session_id": sessionID.String(),
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func getGithubUserID(accessToken string) (int64, error) {
	client := &http.Client{}
	r, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return 0, err
	}
	r.Header.Add("Authorization", "Bearer "+accessToken)
	r.Header.Add("Accept", "application/vnd.github+json")
	r.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(r)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	var user struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (s *Server) PollForToken(ctx context.Context, req *authservice.PollForTokenRequest) (*authservice.PollForTokenResponse, error) {
	uri := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("device_code", req.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	client := &http.Client{}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tokenSuccess struct {
		AccessToken           string `json:"access_token"`
		ExpiresIn             int64  `json:"expires_in"`
		RefreshToken          string `json:"refresh_token"`
		RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
		TokenType             string `json:"token_type"`
		Scope                 string `json:"scope"`
	}
	var tokenError struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		ErrorUri         string `json:"error_uri"`
	}
	_ = json.Unmarshal(body, &tokenError)
	if tokenError.Error == "" {
		_ = json.Unmarshal(body, &tokenSuccess)
		// Fetch GitHub user ID using the access token
		githubUserID, err := getGithubUserID(tokenSuccess.AccessToken)
		if err != nil {
			return &authservice.PollForTokenResponse{
				Error:            "github_user_fetch_failed",
				ErrorDescription: err.Error(),
			}, nil
		}
		sessionID := uuid.New()
		sess := &Session{
			SessionID:             sessionID,
			GithubUserID:          githubUserID,
			AccessToken:           tokenSuccess.AccessToken,
			RefreshToken:          tokenSuccess.RefreshToken,
			ExpiresAt:             time.Now().Add(time.Duration(tokenSuccess.ExpiresIn) * time.Second),
			RefreshTokenExpiresAt: time.Now().Add(time.Duration(tokenSuccess.RefreshTokenExpiresIn) * time.Second),
			CreatedAt:             time.Now(),
		}
		err = s.Sessions.UpsertByGithubUserID(ctx, sess)
		if err != nil {
			return &authservice.PollForTokenResponse{
				Error:            "internal_error",
				ErrorDescription: err.Error(),
			}, nil
		}
		jwtToken, err := generateJWT(sessionID)
		if err != nil {
			return &authservice.PollForTokenResponse{
				Error:            "internal_error",
				ErrorDescription: err.Error(),
			}, nil
		}
		return &authservice.PollForTokenResponse{
			SessionId: sessionID.String(),
			Jwt:       jwtToken,
		}, nil
	} else {
		return &authservice.PollForTokenResponse{
			Error:            tokenError.Error,
			ErrorDescription: tokenError.ErrorDescription,
		}, nil
	}
}

func (s *Server) RefreshToken(ctx context.Context, req *authservice.RefreshTokenRequest) (*authservice.RefreshTokenResponse, error) {
	token, err := jwt.Parse(req.Jwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &authservice.RefreshTokenResponse{
			Error:            "invalid_jwt",
			ErrorDescription: "JWT is invalid or expired",
		}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authservice.RefreshTokenResponse{
			Error:            "invalid_jwt_claims",
			ErrorDescription: "JWT claims are not in the expected format",
		}, nil
	}
	sessionIDStr, ok := claims["session_id"].(string)
	if !ok {
		return &authservice.RefreshTokenResponse{
			Error:            "session_id_missing_in_jwt",
			ErrorDescription: "Session ID is missing in JWT claims",
		}, nil
	}
	sid, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return &authservice.RefreshTokenResponse{
			Error:            "invalid_session_id",
			ErrorDescription: "Session ID in JWT is not a valid UUID",
		}, nil
	}
	sess, err := s.Sessions.GetBySessionID(ctx, sid)
	if err != nil || sess == nil {
		return &authservice.RefreshTokenResponse{
			Error:            "session_not_found",
			ErrorDescription: "Session not found or expired",
		}, nil
	}
	uri := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", sess.RefreshToken)
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("scope", "user:email")

	client := &http.Client{}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Accept", "application/json")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tokenSuccess struct {
		AccessToken           string `json:"access_token"`
		ExpiresIn             int64  `json:"expires_in"`
		RefreshToken          string `json:"refresh_token"`
		RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
	}
	var tokenError struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(body, &tokenError)
	if tokenError.Error == "" {
		_ = json.Unmarshal(body, &tokenSuccess)
		// Update session for the same github_user_id
		sess.AccessToken = tokenSuccess.AccessToken
		sess.RefreshToken = tokenSuccess.RefreshToken
		sess.ExpiresAt = time.Now().Add(time.Duration(tokenSuccess.ExpiresIn) * time.Second)
		sess.RefreshTokenExpiresAt = time.Now().Add(time.Duration(tokenSuccess.RefreshTokenExpiresIn) * time.Second)
		sess.CreatedAt = time.Now()
		err = s.Sessions.UpsertByGithubUserID(ctx, sess)
		if err != nil {
			return &authservice.RefreshTokenResponse{
				Error:            "internal_error",
				ErrorDescription: err.Error(),
			}, nil
		}
		return &authservice.RefreshTokenResponse{}, nil
	} else {
		return &authservice.RefreshTokenResponse{
			Error:            tokenError.Error,
			ErrorDescription: tokenError.ErrorDescription,
		}, nil
	}
}

func (s *Server) ValidateSession(ctx context.Context, req *authservice.ValidateSessionRequest) (*authservice.ValidateSessionResponse, error) {
	token, err := jwt.Parse(req.Jwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &authservice.ValidateSessionResponse{Valid: false, Error: "invalid_jwt"}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authservice.ValidateSessionResponse{Valid: false, Error: "invalid_jwt_claims"}, nil
	}
	sessionIDStr, ok := claims["session_id"].(string)
	if !ok {
		return &authservice.ValidateSessionResponse{Valid: false, Error: "session_id_missing_in_jwt"}, nil
	}
	sid, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return &authservice.ValidateSessionResponse{Valid: false, Error: "invalid_session_id"}, nil
	}
	sess, err := s.Sessions.GetBySessionID(ctx, sid)
	if err != nil || sess == nil {
		return &authservice.ValidateSessionResponse{Valid: false, Error: "Session not found or expired"}, nil
	}
	return &authservice.ValidateSessionResponse{Valid: true}, nil
}

func (s *Server) GetAuthToken(ctx context.Context, req *authservice.GetAuthTokenRequest) (*authservice.GetAuthTokenResponse, error) {
	token, err := jwt.Parse(req.Jwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &authservice.GetAuthTokenResponse{
			Error:            "invalid_jwt",
			ErrorDescription: "JWT is invalid or expired",
		}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authservice.GetAuthTokenResponse{
			Error:            "invalid_jwt_claims",
			ErrorDescription: "JWT claims are not in the expected format",
		}, nil
	}
	sessionIDStr, ok := claims["session_id"].(string)
	if !ok {
		return &authservice.GetAuthTokenResponse{
			Error:            "session_id_missing_in_jwt",
			ErrorDescription: "Session ID is missing in JWT claims",
		}, nil
	}
	sid, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return &authservice.GetAuthTokenResponse{
			Error:            "invalid_session_id",
			ErrorDescription: "Session ID in JWT is not a valid UUID",
		}, nil
	}
	sess, err := s.Sessions.GetBySessionID(ctx, sid)
	if err != nil || sess == nil {
		return &authservice.GetAuthTokenResponse{
			Error:            "session_not_found",
			ErrorDescription: "Session not found or expired",
		}, nil
	}

	if time.Now().After(sess.ExpiresAt) {
		refreshReq := &authservice.RefreshTokenRequest{Jwt: req.Jwt}
		refreshResp, err := s.RefreshToken(ctx, refreshReq)
		if err != nil {
			return &authservice.GetAuthTokenResponse{
				Error:            "internal_error",
				ErrorDescription: err.Error(),
			}, nil
		}
		if refreshResp.Error != "" {
			return &authservice.GetAuthTokenResponse{
				Error:            refreshResp.Error,
				ErrorDescription: refreshResp.ErrorDescription,
			}, nil
		}

		sess, err = s.Sessions.GetBySessionID(ctx, sid)
		if err != nil || sess == nil {
			return &authservice.GetAuthTokenResponse{
				Error:            "session_not_found_after_refresh",
				ErrorDescription: "Session not found after token refresh",
			}, nil
		}
	}

	return &authservice.GetAuthTokenResponse{
		AccessToken: sess.AccessToken,
		TokenType:   "Bearer",
		Scope:       "user:email",
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *authservice.LogoutRequest) (*authservice.LogoutResponse, error) {
	token, err := jwt.Parse(req.Jwt, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &authservice.LogoutResponse{Success: false, Error: "invalid_jwt"}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authservice.LogoutResponse{Success: false, Error: "invalid_jwt_claims"}, nil
	}
	sessionIDStr, ok := claims["session_id"].(string)
	if !ok {
		return &authservice.LogoutResponse{Success: false, Error: "session_id_missing_in_jwt"}, nil
	}
	sid, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return &authservice.LogoutResponse{Success: false, Error: "invalid_session_id"}, nil
	}
	err = s.Sessions.Delete(ctx, sid)
	if err != nil {
		return &authservice.LogoutResponse{Success: false, Error: err.Error()}, nil
	}
	return &authservice.LogoutResponse{Success: true}, nil
}

func RunGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	authservice.RegisterAuthServiceServer(grpcServer, NewServer())
	log.Println("gRPC AuthService server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
