# syntax=docker/dockerfile:1

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
WORKDIR /workspace

# Mod cache
COPY SecretOperationService/go.mod SecretOperationService/go.sum SecretOperationService/
WORKDIR /workspace/SecretOperationService
RUN go mod download

# Copy the repo (for proto and internal)
WORKDIR /workspace
COPY . .

# Build
WORKDIR /workspace/SecretOperationService
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/secretoperationservice ./main.go

# Runtime stage
FROM gcr.io/distroless/base-debian12
WORKDIR /

COPY --from=builder /out/secretoperationservice /secretoperationservice

ENV PORT=50053
EXPOSE 50053

USER nonroot:nonroot
ENTRYPOINT ["/secretoperationservice"]
