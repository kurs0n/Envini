.PHONY: proto run_services stop_services clean

proto:
	protoc --proto_path=proto --go_out=paths=source_relative:AuthService/proto --go-grpc_out=paths=source_relative:AuthService/proto proto/auth.proto
	protoc --proto_path=proto --go_out=paths=source_relative:SecretOperationService/proto --go-grpc_out=paths=source_relative:SecretOperationService/proto proto/secrets.proto

run_services:
	@echo "Starting Auth DB..."
	make -C Database_AuthService run
	@echo "Waiting for Auth DB to initialize..."
	@sleep 5

	@echo "Starting Audit DB..."
	make -C Database_SecretService run &
	@echo "Waiting for Audit DB to initialize..."
	@sleep 5

	@echo "Starting AuthService..."
	cd AuthService && nohup go run main.go > service.log 2>&1 &
	@echo "Waiting for AuthService to initialize..."
	@sleep 5

	@echo "Starting SecretOperationService..."
	cd SecretOperationService && nohup go run main.go > service.log 2>&1 &
	@echo "Waiting for SecretOperationService to initialize..."
	@sleep 5

	@echo "Starting BackendGate..."
	cd BackendGate && npm run start:dev

stop_services:
	make -C Database_AuthService stop
	make -C Database_SecretService stop
	@pkill -f "go run main.go" || true

clean:
	docker rm -f envini-auth-postgres || true
	docker rm -f envini-audit-postgres || true