APP_NAME=zsh-ai-suggestions

build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/${APP_NAME} cmd/main.go

build/arm64:
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/${APP_NAME} cmd/main.go

run:
	go run cmd/main.go

docker: build
	@docker compose build --no-cache
	@docker compose run --rm --remove-orphans zsh-ai-suggestions-playground