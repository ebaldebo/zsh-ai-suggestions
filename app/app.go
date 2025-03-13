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
	"github.com/fsnotify/fsnotify"
)

const tmpDir = "/tmp/zsh-ai-suggestions"

type aiType string

const (
	OpenAI aiType = "openai"
	Ollama aiType = "ollama"
	Gemini aiType = "gemini"
)

func Run() {
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		log.Fatalf("failed to create tmp directory: %v", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	fmt.Println("suggestion server started")

	err = watcher.Add(tmpDir)
	if err != nil {
		log.Fatalf("failed to watch tmp directory: %v", err)
	}

	httpClient := createHttpClient()
	suggester := getSuggester(httpClient)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write) != 0 {
				log.Printf("processing input file: %s", event.Name)
				processInput(suggester, event.Name)
			}
		case err := <-watcher.Errors:
			log.Printf("watcher error: %v", err)
		}
	}
}

func processInput(suggester ai.Suggester, inputFile string) {
	log.Printf("Inside processInput() for file: %s", inputFile)
	if !strings.Contains(inputFile, "-input-") {
		return
	}

	outPutFile := strings.Replace(inputFile, "-input", "-output", 1)

	if _, err := os.Stat(inputFile); err != nil {
		log.Printf("skipping: file does not exist or is unredable %s", inputFile)
		return
	}

	file, err := os.Open(inputFile)
	if err != nil {
		log.Printf("failed to open file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suggestion, err := suggester.Suggest(ctx, input)
	if err != nil {
		log.Printf("failed to get suggestion: %v", err)
		return
	}

	tmpOutput := outPutFile + ".tmp"

	if err = os.WriteFile(tmpOutput, []byte(suggestion), 0o600); err != nil {
		log.Printf("failed to write temporary file: %v", err)
		return
	}

	if err = os.Rename(tmpOutput, outPutFile); err != nil {
		log.Printf("failed to rename temporary file: %v", err)
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
