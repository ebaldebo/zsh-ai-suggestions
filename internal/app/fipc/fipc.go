package fipc

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/gemini"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/ollama"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/openai"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/logger"
	"github.com/fsnotify/fsnotify"
)

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

	tmpDir := env.Get(envTmpDir, defaultTmpDir)

	cleanTempDirectory(tmpDir, logger)

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		logger.Error("failed to create tmp directory: %v", err)
		os.Exit(1)
	}

	setupCleanOnExit(tmpDir, logger)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error("failed to create watcher: %v", err)
		os.Exit(1)
	}
	defer watcher.Close()

	logger.Info("suggestion server started with tmp dir: %s", tmpDir)
	err = watcher.Add(tmpDir)
	if err != nil {
		logger.Error("failed to watch tmp directory: %v", err)
		os.Exit(1)
	}

	httpClient := createHttpClient()
	suggester := getSuggester(httpClient, logger)

	cleanUpOnExit := env.Get(envCleanupOnExit, defaultCleanupOnExit) == "true"
	if cleanUpOnExit {
		logger.Info("clean up on exit enabled")
		go monitorTerminals(tmpDir, logger)
	}

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

func monitorTerminals(tmpDir string, logger logger.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		count, err := countZshProcesses()
		if err != nil {
			logger.Warn("failed to count zsh processes: %v", err)
			continue
		}

		logger.Debug("zsh processes: %d", count)

		if count == 0 {
			logger.Info("no zsh processes found, cleaning up")
			cleanTempDirectory(tmpDir, logger)
			os.Exit(0)
		}
	}
}

func countZshProcesses() (int, error) {
	var cmd *exec.Cmd

	if _, err := os.Stat("/proc"); err == nil {
		cmd = exec.Command("pgrep", "-c", "zsh")
	} else {
		cmd = exec.Command("sh", "-c", "ps -e | grep -c '[z]sh'")
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to count zsh processes: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse zsh process count: %w", err)
	}

	return count, nil
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

func setupCleanOnExit(tmpDir string, logger logger.Logger) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("shutdown signal received, cleaning up")
		cleanTempDirectory(tmpDir, logger)
		os.Exit(0)
	}()
}

func cleanTempDirectory(tmpDir string, logger logger.Logger) {
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		return
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		logger.Warn("failed to read temp directory: %v", err)
		return
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(tmpDir, entry.Name())
		if err := os.Remove(path); err != nil {
			logger.Debug("failed to remove file %s: %v", path, err)
		} else {
			count++
		}
	}

	if count > 0 {
		logger.Info("cleaned up %d files", count)
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
