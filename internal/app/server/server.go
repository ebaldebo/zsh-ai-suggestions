package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/env"
	"github.com/ebaldebo/zsh-ai-suggestions/internal/pkg/logger"
)

func Run() {
	logger := logger.New()
	port := env.GetInt("SERVER_PORT", "5555")

	mux := http.NewServeMux()

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
