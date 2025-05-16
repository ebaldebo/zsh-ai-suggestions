package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/gemini"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/ollama"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/ai/openai"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/logger"
)

type aiType string

const (
	OpenAI aiType = "openai"
	Ollama aiType = "ollama"
	Gemini aiType = "gemini"
)

func Run() {
	logger := logger.New()
	port := env.GetInt("SERVER_PORT", "5555")

	httpClient := createHttpClient()
	suggester := getSuggester(httpClient, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/suggest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		inputBytes, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error("failed to read request body: %w", err)
			http.Error(w, "failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		input := string(inputBytes)
		logger.Info("Received input: %s", input)

		suggestion, err := suggester.Suggest(r.Context(), input)
		if err != nil {
			logger.Error("failed to get suggestion: %w", err)
			http.Error(w, "failed to get suggestion", http.StatusInternalServerError)
			return
		}

		logger.Info("Generated suggestion: %s", suggestion)
		_, err = fmt.Fprintf(w, "%s", suggestion+"\n")
		if err != nil {
			logger.Error("failed to write response: %w", err)
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, "OK")
		if err != nil {
			http.Error(w, fmt.Sprintf("error: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}

	server.SetKeepAlivesEnabled(true)

	go func() {
		logger.Info("server listening on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced shutdown: %v", err)
	}

	logger.Info("server stopped gracefully")
}

func createHttpClient() *http.Client { // TODO: improve
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
