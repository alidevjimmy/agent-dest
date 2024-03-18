all: build test

build:
	@echo "building binaries"
	go build -o main

test:
	@echo "testing project"
	go test ./... -v

run:
	@echo "running project"
	go run .