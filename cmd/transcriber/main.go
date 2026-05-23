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
	"github.com/bcc-code/transcriber/internal/config"
	"github.com/bcc-code/transcriber/internal/jobs"
	"github.com/bcc-code/transcriber/internal/transcriber"
	"github.com/bcc-code/transcriber/internal/transcriber/fasterwhisper"
	"github.com/bcc-code/transcriber/internal/transcriber/stub"
	"github.com/bcc-code/transcriber/internal/transcriber/whispercpp"
	"github.com/bcc-code/transcriber/internal/worker"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	registry := transcriber.NewRegistry(cfg.DefaultModel)
	for id, mc := range cfg.Models {
		adapter, err := buildAdapter(id, mc)
		if err != nil {
			slog.Error("build adapter", "id", id, "err", err)
			os.Exit(1)
		}
		registry.Register(adapter)
	}

	store := jobs.NewStore()
	queue := jobs.NewQueue()
	notifier := callback.NewNotifier(cfg.Server.CallbackWorkers, 256)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier.Start(ctx)

	pool := worker.New(cfg.Server.Workers, store, queue, registry, notifier, func(j jobs.Job) any {
		return api.ToDTO(j)
	})
	pool.Start(ctx)

	srv := api.NewServer(store, queue, registry)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("transcriber api listening",
			"addr", addr,
			"workers", cfg.Server.Workers,
			"default_model", cfg.DefaultModel,
			"models", len(cfg.Models),
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

	queue.Close() // unblock workers waiting for a job
	cancel()      // cancel running jobs
	pool.Wait()
	notifier.Shutdown()
	slog.Info("shutdown complete")
}

func buildAdapter(id string, mc config.ModelConfig) (transcriber.Transcriber, error) {
	switch mc.Adapter {
	case "stub":
		return stub.New(id, "Stub: "+id), nil
	case "whisper-cpp", "whisper.cpp":
		return whispercpp.New(whispercpp.Config{
			ID:        id,
			Binary:    mc.Binary,
			ModelFile: mc.ModelFile,
			Threads:   mc.Threads,
		}), nil
	case "faster-whisper":
		return fasterwhisper.New(fasterwhisper.Config{
			ID:          id,
			Binary:      mc.Binary,
			Model:       mc.Model,
			ComputeType: mc.ComputeType,
			Device:      mc.Device,
		}), nil
	default:
		return nil, fmt.Errorf("unknown adapter: %s", mc.Adapter)
	}
}
