make:
	go mod tidy

build:
	go build -o bin/dockerswarm-configs-provider main.go

run:
	go run main.go
