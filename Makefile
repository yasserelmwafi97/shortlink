.PHONY: run test build lint docker-build docker-run tidy

run:
	go run ./cmd/server

test:
	go test ./... -race -cover

build:
	go build -o bin/shortlink ./cmd/server

lint:
	golangci-lint run

tidy:
	go mod tidy

docker-build:
	docker build -t shortlink .

docker-run:
	docker run --rm -p 8080:8080 -v shortlink-data:/data shortlink
