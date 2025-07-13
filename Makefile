.PHONY: proto

proto:
	protoc --proto_path=proto --go_out=paths=source_relative:AuthService/proto --go-grpc_out=paths=source_relative:AuthService/proto proto/auth.proto
	protoc --proto_path=proto --go_out=paths=source_relative:github-auth-client/proto --go-grpc_out=paths=source_relative:github-auth-client/proto proto/auth.proto