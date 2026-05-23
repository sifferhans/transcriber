# transcriber

Go-based transcription API. Drop-in compatible with the existing Python
service (`POST /transcription/job`, `GET /transcription/job/{id}`), with
an adapter system that lets you swap the underlying ASR backend
(whisper.cpp, faster-whisper, stub, ...) per request via an additive
`model` field.

## Layout

```
cmd/transcriber/        entrypoint
internal/api/           HTTP server + DTOs
internal/jobs/          in-memory job store + priority queue
internal/worker/        worker goroutine pool
internal/transcriber/   adapter interface
  ├─ stub/              fake adapter for local dev / tests
  ├─ whispercpp/        shells out to whisper.cpp CLI
  └─ fasterwhisper/     shells out to whisper-ctranslate2
internal/formats/       json / srt / vtt / txt writers
internal/callback/      goroutine pool that POSTs webhooks
internal/web/           embedded SPA (//go:embed dist/* + SPA fallback)
frontend/               Nuxt 4 SPA (ssr: false) — pnpm generate output
                          is copied into internal/web/dist by `make frontend`
```

## Run

Two modes:

**Development** — Go API + Nuxt dev server with hot reload.

```sh
make dev            # both at once; Ctrl-C stops both
# or in separate terminals:
make dev-api        # Go on :8888
make dev-frontend   # Nuxt on :3000 (proxies API calls)
```

**Single-binary** — SPA embedded in the Go binary, both served from `:8888`.

```sh
make build          # pnpm generate → internal/web/dist → go build
./transcriber
```

Both default to the `stub` adapter so they work without any ASR backend
installed. `internal/web/dist/` must be populated before the Go side will
compile (the `//go:embed` directive needs at least one file) — run
`make frontend` once after cloning, then `go run ./cmd/transcriber` works
on its own for API-only iteration.

## Configuration

The set of registered models lives in `cmd/transcriber/models.go` as typed
Go code. Server settings come from flags; per-machine paths from env vars.

| Flag                | Default | Meaning                                        |
| ------------------- | ------- | ---------------------------------------------- |
| `-port`             | `8888`  | HTTP listen port                               |
| `-workers`          | `2`     | concurrent transcription jobs                  |
| `-callback-workers` | `2`     | webhook delivery goroutines                    |
| `-default-model`    | `stub`  | adapter ID used when the request omits `model` |

| Env var                       | Default                              | Meaning                             |
| ----------------------------- | ------------------------------------ | ----------------------------------- |
| `WHISPER_CPP_BIN`             | `/opt/homebrew/bin/whisper-cli`      | whisper.cpp binary                  |
| `WHISPER_CPP_MODEL`           | `$HOME/models/ggml-base.bin`         | whisper.cpp model file              |
| `FASTER_WHISPER_BIN`          | `/usr/local/bin/whisper-ctranslate2` | faster-whisper CLI                  |
| `FASTER_WHISPER_COMPUTE_TYPE` | `float16`                            | float16 / int8_float16 / int8 / ... |
| `FASTER_WHISPER_DEVICE`       | `cuda`                               | cuda / cpu / auto                   |

## API

### `POST /transcription/job`

```json
{
  "path": "/mnt/storage/audio/foo.wav",
  "language": "no",
  "format": "all",
  "output_path": "/mnt/storage/out/foo/",
  "priority": 5,
  "callback": "https://example.com/hook",
  "model": "whisper-cpp-large-v3"
}
```

`model` is optional — omit to use the default. `format: "all"` writes
json+srt+vtt+txt; or pass a comma-separated subset like `"json,srt"`.

### `GET /transcription/job/{id}`

Returns the current job state. `status` is one of `PENDING`, `RUNNING`,
`COMPLETED`, `FAILED`, `CANCELED`. `progress` is 0–100. `result` is the
path to `transcript.json` once `COMPLETED`.

### Additive endpoints

- `GET  /transcription/jobs` — list all jobs
- `DELETE /transcription/job/{id}` — cancel a queued or running job
- `GET  /models` — list registered adapters
- `GET  /healthz`, `GET /readyz`

## Adding a new backend

1. Implement `transcriber.Transcriber` in `internal/transcriber/<name>/`.
2. Add another `r.Register(...)` call in `cmd/transcriber/models.go` with
   the adapter's typed `Config`. Use a distinct ID per variant
   (e.g. `whisper-cpp-large-v3`, `whisper-cpp-medium`) so callers can A/B
   test by passing `"model": "..."` in the request body.
