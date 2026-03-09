.PHONY: run build tidy

APP_NAME=jedug-api

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

tidy:
	go mod tidy
