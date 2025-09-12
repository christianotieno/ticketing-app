.PHONY: help start build test clean

help:
	@echo "Available commands:"
	@echo "  make start       - Run the app"
	@echo "  make build       - Build the app"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean up files"

start:
	go run .

build:
	go build -o ticketing-app .

test:
	go test -v

clean:
	rm -f ticketing-app
	go clean