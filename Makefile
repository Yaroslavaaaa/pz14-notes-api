.PHONY: run swagger

run:
	go run ./cmd/api

swagger:
	swag init -g cmd/api/main.go -o docs