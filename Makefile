
install-deps:
	go mod tidy
	go mod download

lint:
	golangci-lint run --config=.golangci.yml

build:
	go build -o bin/ffmpegbox ./cmd/ffmpegbox
