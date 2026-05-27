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
  ├─ fasterwhisper/     shells out to whisper-ctranslate2
  └─ chunked/           wrapper that splits long-form audio into
                        overlapping chunks (ffmpeg/ffprobe), transcribes
                        each via an inner adapter, and stitches results
                        back into one timeline
internal/formats/       json / srt / vtt / txt writers
                        testdata/golden/ — snapshot fixtures shared by
                        format and adapter parser tests as the
                        cross-adapter output contract
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

The real adapters (`whisper-cpp-*`, `faster-whisper-*`) require **ffmpeg**
and **ffprobe** on `$PATH`: the chunked wrapper uses ffprobe to read the
input duration and ffmpeg to extract each chunk to a 16kHz mono wav. The
`stub` adapter has no external dependencies.

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

```sh
# Submit a job, then poll until it completes.
JOB=$(curl -sS -X POST http://localhost:8888/transcription/job \
    -H 'content-type: application/json' \
    -d '{
        "path": "/mnt/storage/audio/foo.wav",
        "language": "no",
        "format": "all",
        "output_path": "/mnt/storage/out/foo/",
        "model": "whisper-cpp-large-v3"
    }' | jq -r .id)

while :; do
    curl -sS "http://localhost:8888/transcription/job/$JOB" | jq '{status, progress, result}'
    sleep 2
done
```

### `GET /transcription/job/{id}`

Returns the current job state. `status` is one of `PENDING`, `RUNNING`,
`COMPLETED`, `FAILED`, `CANCELED`. `progress` is 0–100. `result` is the
path to `transcript.json` once `COMPLETED`.

The JSON result includes word-level timestamps. Each segment carries a
`words` array (`text` / `start` / `end`) in addition to the segment-level
`text`/`start`/`end`. See
`internal/formats/testdata/golden/transcript.json` for the canonical
shape — this fixture is the source of truth that every adapter must
serialize to.

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
3. Wrap the adapter in `chunked.New(inner, chunked.Config{})` at
   registration time if it should handle long-form audio — the wrapper
   passes short files (≤ `ChunkLengthSec`, default 5 min) through
   unchanged and chunks longer files transparently.
4. Add a `testdata/raw.json` fixture and a parser test in the adapter
   package that round-trips the parsed `Transcription` through
   `formats.Write` and byte-compares each output against
   `internal/formats/testdata/golden/transcript.<ext>`. This is the
   contract every adapter is held to.
