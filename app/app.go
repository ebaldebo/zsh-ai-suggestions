package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/gemini"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/ollama"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/openai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/env"
)

type aiType string

const (
	OpenAI aiType = "openai"
	Ollama aiType = "ollama"
	Gemini aiType = "gemini"
)

func Run() {
	httpClient := createHttpClient()

	suggester := getSuggester(httpClient)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		ctx := context.Background()
		suggestion, err := suggester.Suggest(ctx, input)
		if err != nil {
			log.Printf("failed to suggest: %v", err)
			continue
		}

		fmt.Println(suggestion)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read input: %v", err)
	}
}

func createHttpClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func getSuggester(httpClient *http.Client) ai.Suggester {
	aiType := aiType(env.Get(envAIType, defaultAIType))
	var suggester ai.Suggester
	switch aiType {
	case OpenAI:
		suggester = openai.New(httpClient)
	case Ollama:
		suggester = ollama.New(httpClient)
	case Gemini:
		suggester = gemini.New(httpClient)
	default:
		log.Fatalf("unsupported AI type: %s", aiType)
	}
	return suggester
}
