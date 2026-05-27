package main

import (
	"context"
	"os"

	"transcriber/internal/hfcache"
	"transcriber/internal/transcriber"
	"transcriber/internal/transcriber/chunked"
	"transcriber/internal/transcriber/fasterwhisper"
	"transcriber/internal/transcriber/stub"
	"transcriber/internal/transcriber/whispercpp"
)

// buildRegistry declares every Transcriber the server can use.
//
// To add a new model: import its adapter package and append another
// Register call. Multiple variants of the same backend (e.g. medium vs
// large) live as separate entries with distinct IDs so callers can pick
// per request via the `model` field.
//
// Per-machine paths (binaries, model files) are read from env vars so the
// same binary works in dev and prod without a recompile.
func buildRegistry(defaultID string) *transcriber.Registry {
	r := transcriber.NewRegistry(defaultID)

	r.Register(stub.New("stub", "Stub Adapter"))

	// Model files are fetched from Hugging Face on first use and cached on
	// disk. `WHISPER_CPP_MODEL` still works as a path override for operators
	// who want to pin a specific local file (air-gapped boxes, custom models).
	cache := hfcache.Default()

	// Long-form inputs (sermons, lectures) hit chunking; short files pass
	// through the wrapper unchanged. ffmpeg/ffprobe must be on PATH.
	r.Register(chunked.New(
		whispercpp.New(whispercpp.Config{
			ID:        "whisper-cpp-large-v3",
			Binary:    envOr("WHISPER_CPP_BIN", "/opt/homebrew/bin/whisper-cli"),
			ModelFile: os.Getenv("WHISPER_CPP_MODEL"),
			ResolveModel: func(ctx context.Context) (string, error) {
				return cache.Get(ctx, "ggerganov/whisper.cpp", "ggml-large-v3.bin")
			},
			Threads: 8,
		}),
		chunked.Config{},
	))

	r.Register(chunked.New(
		fasterwhisper.New(fasterwhisper.Config{
			ID:          "faster-whisper-large-v3",
			Binary:      envOr("FASTER_WHISPER_BIN", "/usr/local/bin/whisper-ctranslate2"),
			Model:       "large-v3",
			ComputeType: envOr("FASTER_WHISPER_COMPUTE_TYPE", "float16"),
			Device:      envOr("FASTER_WHISPER_DEVICE", "cuda"),
		}),
		chunked.Config{},
	))

	// Norwegian-tuned model from NbAiLab. The Model string is passed straight
	// to `whisper-ctranslate2 --model`, which accepts either a CTranslate2
	// HF repo ID (downloaded + cached on first use) or a local path.
	r.Register(chunked.New(
		fasterwhisper.New(fasterwhisper.Config{
			ID:          "nb-whisper-large",
			Binary:      envOr("FASTER_WHISPER_BIN", "/usr/local/bin/whisper-ctranslate2"),
			Model:       envOr("NB_WHISPER_MODEL", "NbAiLab/nb-whisper-large"),
			ComputeType: envOr("FASTER_WHISPER_COMPUTE_TYPE", "float16"),
			Device:      envOr("FASTER_WHISPER_DEVICE", "cuda"),
		}),
		chunked.Config{},
	))

	// Drop-in aliases for the legacy ai-api allowlist. Sub-large OpenAI sizes
	// and sub-large NbAiLab sizes intentionally upgrade to the large model —
	// faithful to the old runner's behavior of silently coercing anything
	// outside its allowlist to openai/whisper-large-v3.
	r.Alias("openai/whisper-large-v3", "faster-whisper-large-v3")
	r.Alias("openai/whisper-large-v2", "faster-whisper-large-v3")
	r.Alias("openai/whisper-large", "faster-whisper-large-v3")
	r.Alias("openai/whisper-medium", "faster-whisper-large-v3")
	r.Alias("openai/whisper-small", "faster-whisper-large-v3")
	r.Alias("openai/whisper-tiny", "faster-whisper-large-v3")
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
