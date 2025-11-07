.PHONY: build run test

build:
	go build -o app cmd/app/main.go

run:
	go run cmd/app/main.go

test:
	go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep total:
	go tool cover -html=coverage.out

compose-up:
	docker-compose up -d

compose-down:
	docker-compose down