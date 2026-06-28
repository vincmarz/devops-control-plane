# DevOps Control Plane container image
# Multi-stage build: compile a static Go binary, then run it as non-root.

FROM docker.io/library/golang:1.22 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/devops-control-plane \
    ./cmd/devops-control-plane

FROM registry.access.redhat.com/ubi9/ubi-micro:latest
WORKDIR /app

COPY --from=builder /out/devops-control-plane /app/devops-control-plane

ENV HTTP_ADDR=:8080 \
    LOG_LEVEL=info

EXPOSE 8080
USER 1001

ENTRYPOINT ["/app/devops-control-plane"]
