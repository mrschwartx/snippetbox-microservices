ifneq (,$(wildcard .env))
include .env
export
endif

COVER_PKGS := ./internal/...

.PHONY: tidy run build fmt vet test test/race test/integration test/e2e test/all test/cover

tidy:
	go mod tidy

run:
	go run ./cmd/backend

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/backend ./cmd/backend

fmt:
	gofmt -s -w .

vet:
	go vet -tags=integration,e2e ./...

test:
	go test -count=1 ./...

test/race:
	go test -race -count=1 ./...

test/integration:
	go test -count=1 -tags=integration ./test/integration/...

test/e2e:
	go test -count=1 -tags=e2e ./test/e2e/...

test/all:
	go test -count=1 -tags=integration,e2e ./...

test/cover:
	go test -count=1 -tags=integration,e2e -coverpkg=$(COVER_PKGS) -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1
