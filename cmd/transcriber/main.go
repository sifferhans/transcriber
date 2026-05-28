package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"transcriber/internal/api"
	"transcriber/internal/callback"
	"transcriber/internal/jobs"
	"transcriber/internal/web"
	"transcriber/internal/worker"
)

func main() {
	port := flag.Int("port", 8888, "HTTP listen port")
	workers := flag.Int("workers", 2, "number of transcription worker goroutines")
	callbackWorkers := flag.Int("callback-workers", 2, "number of webhook delivery goroutines")
	defaultModel := flag.String("default-model", "stub", "model adapter ID to use when the request omits `model`")
	defaultPromptFile := flag.String("default-prompt-file", "prompt.txt", "path to a file whose contents are used as the prompt when the request omits one (missing file = no default prompt)")
	maxTerminalJobs := flag.Int("max-terminal-jobs", 20, "how many finished jobs (completed/failed/canceled) to retain in memory; <= 0 disables the cap")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	var defaultPrompt string
	if *defaultPromptFile != "" {
		b, err := os.ReadFile(*defaultPromptFile)
		switch {
		case err == nil:
			defaultPrompt = strings.TrimSpace(string(b))
			slog.Info("loaded default prompt", "path", *defaultPromptFile, "bytes", len(defaultPrompt))
		case os.IsNotExist(err):
			slog.Info("no default prompt file found", "path", *defaultPromptFile)
		default:
			slog.Error("reading default prompt file", "path", *defaultPromptFile, "err", err)
			os.Exit(1)
		}
	}

	registry := buildRegistry(*defaultModel)
	if _, ok := registry.Default(); !ok {
		slog.Error("default model not registered", "id", *defaultModel)
		os.Exit(1)
	}

	store := jobs.NewStore(*maxTerminalJobs)
	queue := jobs.NewQueue()
	notifier := callback.NewNotifier(*callbackWorkers, 256)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier.Start(ctx)

	pool := worker.New(*workers, store, queue, registry, notifier, func(j jobs.Job) any {
		return api.ToDTO(j)
	})
	pool.Start(ctx)

	srv := api.NewServer(store, queue, registry, defaultPrompt)
	addr := fmt.Sprintf(":%d", *port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(web.Handler()),
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
