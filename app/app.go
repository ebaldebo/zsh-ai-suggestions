package app

import (
	"bufio"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/gemini"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/ollama"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/ai/openai"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

const tmpDir = "/tmp/zsh-ai-suggestions"

type aiType string

const (
	OpenAI aiType = "openai"
	Ollama aiType = "ollama"
	Gemini aiType = "gemini"
)

var (
	processingFiles = make(map[string]bool)
	processingMutex sync.Mutex
)

func Run() {
	logger := logger.New()

	cleanupStaleFiles(logger)

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		logger.Error("failed to create tmp directory: %v", err)
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("failed to create watcher: %v", err)
		os.Exit(1)
	}
	defer watcher.Close()

	logger.Info("suggestion server started")
	err = watcher.Add(tmpDir)
	if err != nil {
		logger.Error("failed to watch tmp directory: %v", err)
		os.Exit(1)
	}

	httpClient := createHttpClient()
	suggester := getSuggester(httpClient, logger)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&(fsnotify.Write) != 0 && shouldProcessFile(event.Name, logger) {
				go processInput(suggester, event.Name, logger)
			}
		case err := <-watcher.Errors:
			logger.Warn("watcher error: %v", err)
		}
	}
}

func shouldProcessFile(filename string, logger logger.Logger) bool {
	if strings.Contains(filename, "-output-") || strings.HasSuffix(filename, ".tmp") {
		return false
	}

	if !strings.Contains(filename, "-input-") {
		return false
	}

	processingMutex.Lock()
	defer processingMutex.Unlock()

	if processingFiles[filename] {
		logger.Debug("file is already being processed: %s", filepath.Base(filename))
		return false
	}

	processingFiles[filename] = true
	return true
}

func finishProcessing(filename string) {
	processingMutex.Lock()
	defer processingMutex.Unlock()
	delete(processingFiles, filename)
}

func processInput(suggester ai.Suggester, inputFile string, logger logger.Logger) {
	defer finishProcessing(inputFile)

	basename := filepath.Base(inputFile)
	logger.Debug("processing: %s", basename)

	outPutFile := strings.Replace(inputFile, "-input", "-output", 1)

	if _, err := os.Stat(inputFile); err != nil {
		logger.Debug("file not accessible: %s", basename)
		return
	}

	time.Sleep(50 * time.Millisecond)

	file, err := os.Open(inputFile)
	if err != nil {
		logger.Warn("failed to open file: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		logger.Debug("empty file: %s", basename)
		return
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		logger.Debug("empty input in file: %s", basename)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	suggestion, err := suggester.Suggest(ctx, input)
	if err != nil {
		logger.Error("failed to get suggestion: %v", err)
		return
	}

	if err = os.WriteFile(outPutFile, []byte(suggestion), 0o600); err != nil {
		logger.Error("failed to write output file: %v", err)
		return
	}

	logger.Debug("suggestion generated for: %s", basename)

	os.Remove(inputFile)
}

func cleanupStaleFiles(logger logger.Logger) {
	files, err := filepath.Glob(filepath.Join(tmpDir, "zsh-ai-input-*"))
	if err == nil && len(files) > 0 {
		logger.Debug("cleaning up %d stale files", len(files))
		for _, file := range files {
			os.Remove(file)
		}
	}

	files, err = filepath.Glob(filepath.Join(tmpDir, "*.tmp"))
	if err == nil && len(files) > 0 {
		logger.Debug("cleaning up %d stale files", len(files))
		for _, file := range files {
			os.Remove(file)
		}
	}
}

func createHttpClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func getSuggester(httpClient *http.Client, logger logger.Logger) ai.Suggester {
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
		logger.Error("unknown AI type: %s", aiType)
		os.Exit(1)
	}

	return suggester
}
