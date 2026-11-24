// Command ratings-api serves product ratings for the Demo Commerce Platform.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/devexio/demo-ratings-api/internal/config"
	"github.com/devexio/demo-ratings-api/internal/handler"
	"github.com/devexio/demo-ratings-api/internal/metrics"
	"github.com/devexio/demo-ratings-api/internal/middleware"
	"github.com/devexio/demo-ratings-api/internal/store"
)

func main() {
	cfg := config.FromEnv()

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		level = slog.LevelInfo
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	st := store.New()
	m := metrics.New()
	api := handler.NewAPI(st, m, func() bool { return !cfg.FailReadiness })

	// Context must be the outermost middleware so the enriched request
	// (request ID + trace ID) is visible to the logging/metrics layers.
	root := middleware.Context(
		middleware.Logging(log)(
			middleware.Metrics(m)(api.Mux())))

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("ratings-api listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}
