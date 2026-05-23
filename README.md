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
internal/config/        JSON config loader
```

## Run

```sh
cp config.example.json config.json
go run ./cmd/transcriber -config config.json
```

The example config defaults to the `stub` adapter so the server works without
any backend installed.

## API

### `POST /transcription/job`

```json
{
  "path":        "/mnt/storage/audio/foo.wav",
  "language":    "no",
  "format":      "all",
  "output_path": "/mnt/storage/out/foo/",
  "priority":    5,
  "callback":    "https://example.com/hook",
  "model":       "whisper-cpp-large-v3"
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
2. Add a `case "<adapter-id>":` to `buildAdapter` in `cmd/transcriber/main.go`.
3. Add an entry to `config.json`.
