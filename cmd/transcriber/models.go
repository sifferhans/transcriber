package main

import (
	"os"
	"path/filepath"

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

	home, _ := os.UserHomeDir()

	r.Register(stub.New("stub", "Stub Adapter"))

	// Long-form inputs (sermons, lectures) hit chunking; short files pass
	// through the wrapper unchanged. ffmpeg/ffprobe must be on PATH.
	r.Register(chunked.New(
		whispercpp.New(whispercpp.Config{
			ID:        "whisper-cpp-large-v3",
			Binary:    envOr("WHISPER_CPP_BIN", "/opt/homebrew/bin/whisper-cli"),
			ModelFile: envOr("WHISPER_CPP_MODEL", filepath.Join(home, "models", "ggml-base.bin")),
			Threads:   8,
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

	return r
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
