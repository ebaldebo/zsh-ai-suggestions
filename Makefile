APP_NAME=zsh-ai-suggestions

build/fipc:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${APP_NAME} cmd/fipc/main.go

run:
	go run cmd/main.go

docker/fipc: build/fipc
	@docker compose build
	@docker compose run --rm --remove-orphans zsh-ai-suggestions-playground
