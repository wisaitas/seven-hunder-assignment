.PHONY: build run 

build:
	go build -o app cmd/app/main.go

run:
	go run cmd/app/main.go