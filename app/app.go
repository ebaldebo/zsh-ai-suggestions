package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/ollama"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/openai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/env"
)

type aiType string

const (
	OpenAI aiType = "openai"
	Ollama aiType = "ollama"
)

func Run() {
	suggester := getSuggester()

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

func getSuggester() ai.Suggester {
	aiType := aiType(env.Get(envAIType, defaultAIType))
	var suggester ai.Suggester
	switch aiType {
	case OpenAI:
		suggester = openai.New()
	case Ollama:
		suggester = ollama.New()
	default:
		log.Fatalf("unsupported AI type: %s", aiType)
	}
	return suggester
}
