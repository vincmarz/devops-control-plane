APP_NAME := devops-control-plane
PKG := ./...

.PHONY: run build test fmt vet clean migrate-up migrate-down migrate-runtime-state-up migrate-runtime-state-down

run:
	go run ./cmd/devops-control-plane

build:
	mkdir -p bin
	go build -o bin/$(APP_NAME) ./cmd/devops-control-plane

test:
	go test $(PKG)

fmt:
	go fmt $(PKG)

vet:
	go vet $(PKG)

migrate-up:
	go run ./cmd/devops-control-plane-migrate -direction up

migrate-down:
	go run ./cmd/devops-control-plane-migrate -direction down

migrate-runtime-state-up:
	go run ./cmd/devops-control-plane-migrate -direction up -up migrations/000002_change_runtime_states.up.sql

migrate-runtime-state-down:
	go run ./cmd/devops-control-plane-migrate -direction down -down migrations/000002_change_runtime_states.down.sql

clean:
	rm -rf bin coverage.out coverage.html
