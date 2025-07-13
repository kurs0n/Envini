package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	authservice "github.com/kurs0n/github-auth-client/proto"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := authservice.NewAuthServiceClient(conn)

	ctx := context.Background()
	resp, err := client.StartDeviceFlow(ctx, &authservice.StartDeviceFlowRequest{})
	if err != nil {
		log.Fatalf("StartDeviceFlow error: %v", err)
	}
	fmt.Printf("Go to: %s\nEnter code: %s\n", resp.VerificationUri, resp.UserCode)

	var jwt string
	for {
		time.Sleep(time.Duration(resp.Interval) * time.Second)
		tok, err := client.PollForToken(ctx, &authservice.PollForTokenRequest{
			DeviceCode: resp.DeviceCode,
		})
		if err != nil {
			log.Fatalf("PollForToken error: %v", err)
		}
		if tok.Error != "" {
			if tok.Error == "authorization_pending" || tok.Error == "slow_down" {
				fmt.Println("Waiting for user authorization...")
				continue
			}
			log.Fatalf("Token error: %s - %s", tok.Error, tok.ErrorDescription)
		}
		jwt = tok.Jwt
		fmt.Printf("Authenticated! JWT: %s\n", jwt)
		break
	}

} 