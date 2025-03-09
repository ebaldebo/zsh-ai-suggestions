package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/openai"
)

func main() {
	suggester := openai.NewOpenAIClient()

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
