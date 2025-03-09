APP_NAME=zsh-ai-suggestions

build:
	go build -o bin/${APP_NAME} cmd/main.go

run:
	go run cmd/main.go