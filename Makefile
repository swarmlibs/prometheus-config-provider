make:
	go mod tidy

build:
	go build -o bin/prometheus-configs-provider main.go

run:
	go run main.go

clean:
	rm -rf bin
