# transcriber

Go-based transcription API. Drop-in compatible with the existing Python
service consumed by `example.md` (`POST /transcription/job`, `GET
/transcription/job/{id}`), with an adapter system that lets you swap the
underlying ASR backend (whisper.cpp, faster-whisper, stub, ...) per request
via an additive `model` field.

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
```

## Run

```sh
go run ./cmd/transcriber
```

That's it — defaults to the `stub` adapter on `:8888` with 2 workers, no
external backend required.

## Configuration

The set of registered models lives in `cmd/transcriber/models.go` as typed
Go code. Server settings come from flags; per-machine paths from env vars.

| Flag | Default | Meaning |
| --- | --- | --- |
| `-port` | `8888` | HTTP listen port |
| `-workers` | `2` | concurrent transcription jobs |
| `-callback-workers` | `2` | webhook delivery goroutines |
| `-default-model` | `stub` | adapter ID used when the request omits `model` |

| Env var | Default | Meaning |
| --- | --- | --- |
| `WHISPER_CPP_BIN` | `/opt/homebrew/bin/whisper-cli` | whisper.cpp binary |
| `WHISPER_CPP_MODEL` | `/models/ggml-large-v3.bin` | whisper.cpp model file |
| `FASTER_WHISPER_BIN` | `/usr/local/bin/whisper-ctranslate2` | faster-whisper CLI |
| `FASTER_WHISPER_COMPUTE_TYPE` | `float16` | float16 / int8_float16 / int8 / ... |
| `FASTER_WHISPER_DEVICE` | `cuda` | cuda / cpu / auto |

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
