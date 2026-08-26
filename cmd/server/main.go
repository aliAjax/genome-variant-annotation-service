package main

import (
	"context"
	"github.com/example/genome-variant-annotation/internal/annotation"
	"github.com/example/genome-variant-annotation/internal/httpapi"
	"github.com/example/genome-variant-annotation/internal/job"
	"github.com/example/genome-variant-annotation/internal/normalization"
	"github.com/example/genome-variant-annotation/internal/platform"
	"github.com/example/genome-variant-annotation/internal/reference"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := platform.LoadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	clock := platform.SystemClock{}
	referenceRepo := reference.NewMemoryRepository()
	referenceService := reference.NewService(referenceRepo, clock)
	sequence := normalization.NewMemoryReference()
	sequence.Set("1", repeat("ACGT", 100000))
	normalizer := normalization.NewService(sequence)
	annotator := annotation.NewService(normalizer, referenceService)
	jobRepo := job.NewMemoryRepository()
	jobService := job.NewService(jobRepo, annotator, clock, 1024)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	jobService.Run(ctx, cfg.WorkerCount)
	handler := httpapi.NewServer(referenceService, jobService, normalizer, cfg.AuthToken).Handler()
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		logger.Info("server started", "address", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			cancel()
		}
	}()
	<-ctx.Done()
	shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if err := server.Shutdown(shutdown); err != nil {
		logger.Error("shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
func repeat(value string, count int) string {
	result := make([]byte, 0, len(value)*count)
	for i := 0; i < count; i++ {
		result = append(result, value...)
	}
	return string(result)
}
