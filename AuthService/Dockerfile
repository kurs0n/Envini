# syntax=docker/dockerfile:1

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23 AS builder
WORKDIR /workspace

# Copy go mod/sum first for caching
COPY AuthService/go.mod AuthService/go.sum AuthService/
WORKDIR /workspace/AuthService
RUN go mod download

# Copy the rest (including shared proto from repo root)
WORKDIR /workspace
COPY . .

# Build binary (replace path to main)
WORKDIR /workspace/AuthService
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/authservice ./main.go

# Runtime stage
FROM gcr.io/distroless/base-debian12
WORKDIR /

COPY --from=builder /out/authservice /authservice

ENV PORT=50052
EXPOSE 50052

USER nonroot:nonroot
ENTRYPOINT ["/authservice"]
