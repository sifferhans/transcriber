package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bcc-code/transcriber/internal/api"
	"github.com/bcc-code/transcriber/internal/callback"
	"github.com/bcc-code/transcriber/internal/jobs"
	"github.com/bcc-code/transcriber/internal/worker"
)

func main() {
	port := flag.Int("port", 8888, "HTTP listen port")
	workers := flag.Int("workers", 2, "number of transcription worker goroutines")
	callbackWorkers := flag.Int("callback-workers", 2, "number of webhook delivery goroutines")
	defaultModel := flag.String("default-model", "stub", "model adapter ID to use when the request omits `model`")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	registry := buildRegistry(*defaultModel)
	if _, ok := registry.Default(); !ok {
		slog.Error("default model not registered", "id", *defaultModel)
		os.Exit(1)
	}

	store := jobs.NewStore()
	queue := jobs.NewQueue()
	notifier := callback.NewNotifier(*callbackWorkers, 256)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier.Start(ctx)

	pool := worker.New(*workers, store, queue, registry, notifier, func(j jobs.Job) any {
		return api.ToDTO(j)
	})
	pool.Start(ctx)

	srv := api.NewServer(store, queue, registry)
	addr := fmt.Sprintf(":%d", *port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("transcriber api listening",
			"addr", addr,
			"workers", *workers,
			"default_model", *defaultModel,
			"models", len(registry.List()),
		)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	slog.Info("shutdown signal received")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = httpSrv.Shutdown(shutdownCtx)

	queue.Close()
	cancel()
	pool.Wait()
	notifier.Shutdown()
	slog.Info("shutdown complete")
}
