.PHONY: help start build test docker-build docker-run clean

help:
	@echo "Available commands:"
	@echo "  make start       - Run the app"
	@echo "  make build       - Build the app"
	@echo "  make test        - Run tests"
	@echo "  make docker      - Build and run with Docker"
	@echo "  make clean       - Clean up files"

start:
	go run .

build:
	go build -o ticketing-app .

test:
	go test -v

docker-build:
	docker build -t ticketing-app .

docker-run:
	docker run --rm ticketing-app

docker: docker-build docker-run

clean:
	rm -f ticketing-app
	go clean