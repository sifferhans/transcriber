package main

import (
	"context"
	"os"

	"transcriber/internal/hfcache"
	"transcriber/internal/transcriber"
	"transcriber/internal/transcriber/chunked"
	"transcriber/internal/transcriber/stub"
	"transcriber/internal/transcriber/whispercpp"
)

// buildRegistry declares every Transcriber the server can use.
//
// To add a new model: import its adapter package and append another
// Register call. Multiple variants of the same backend live as separate
// entries with distinct IDs so a job request can pick between them via
// the `model` field.
//
// Model files are fetched from Hugging Face on first use and cached on
// disk via internal/hfcache. The matching `*_MODEL` env vars still work
// as local-path overrides for operator-pinned deployments.
func buildRegistry(defaultID string) *transcriber.Registry {
	r := transcriber.NewRegistry(defaultID)

	r.Register(stub.New("stub", "Stub Adapter"))

	cache := hfcache.Default()
	whisperBin := envOr("WHISPER_CPP_BIN", "/opt/homebrew/bin/whisper-cli")

	// OpenAI Whisper large-v3 via whisper.cpp. GGML pulled from ggerganov's
	// official mirror on first request.
	r.Register(chunked.New(
		whispercpp.New(whispercpp.Config{
			ID:          "whisper-cpp-large-v3",
			DisplayName: "OpenAI Whisper large-v3",
			Binary:      whisperBin,
			ModelFile:   os.Getenv("WHISPER_CPP_MODEL"),
			ResolveModel: func(ctx context.Context) (string, error) {
				return cache.Get(ctx, "ggerganov/whisper.cpp", "ggml-large-v3.bin")
			},
			Threads: 8,
		}),
		chunked.Config{},
	))

	// Norwegian-tuned Whisper large from NbAiLab. The repo itself ships a
	// GGML build (`ggml-model.bin`) alongside the safetensors weights.
	r.Register(chunked.New(
		whispercpp.New(whispercpp.Config{
			ID:          "nb-whisper-large",
			DisplayName: "NB-Whisper large (Norwegian)",
			Binary:      whisperBin,
			ModelFile:   os.Getenv("NB_WHISPER_MODEL"),
			ResolveModel: func(ctx context.Context) (string, error) {
				return cache.Get(ctx, "NbAiLab/nb-whisper-large", "ggml-model.bin")
			},
			Threads: 8,
		}),
		chunked.Config{},
	))

	// Drop-in aliases for the legacy Python API's allowlist. Sub-large sizes
	// upgrade to the large model — faithful to the old runner's behavior
	// of silently coercing anything outside its allowlist to large.
	r.Alias("openai/whisper-large-v3", "whisper-cpp-large-v3")
	r.Alias("openai/whisper-large-v2", "whisper-cpp-large-v3")
	r.Alias("openai/whisper-large", "whisper-cpp-large-v3")
	r.Alias("openai/whisper-medium", "whisper-cpp-large-v3")
	r.Alias("openai/whisper-small", "whisper-cpp-large-v3")
	r.Alias("openai/whisper-tiny", "whisper-cpp-large-v3")
	r.Alias("NbAiLab/nb-whisper-large", "nb-whisper-large")
	r.Alias("NbAiLab/nb-whisper-medium", "nb-whisper-large")
	r.Alias("NbAiLab/nb-whisper-small", "nb-whisper-large")
	r.Alias("NbAiLab/nb-whisper-tiny", "nb-whisper-large")

	return r
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
