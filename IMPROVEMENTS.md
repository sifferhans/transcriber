# Potential improvements

Working notes for hardening this into a production API-first service that
backs a VOD pipeline. Ordered roughly by ROI for that use case; the
chunking work for long-form audio is already in `internal/transcriber/chunked`.

## Persistent job store

`internal/jobs.Store` is in-memory: a server restart loses the queue and all
history. For an API behind a pipeline this is unacceptable — retries can't
deduplicate, callers can't query status across restarts, and a crash mid-job
silently drops work.

Approach: drop in SQLite via `modernc.org/sqlite` (pure Go, no CGO). The
existing `Store` interface fits a SQL backing with `jobs` + `job_events`
tables. Restore in-flight jobs on boot (mark `RUNNING`-but-unowned ones as
`PENDING` and requeue). Half-day swap.

## Idempotency keys

Pipelines retry on infra hiccups, and without deduplication the API will
happily transcribe the same file twice. Accept `Idempotency-Key` (header or
body field) on POST; cache the resulting job id keyed by (api-key,
idempotency-key) for a fixed window (24h is conventional). Repeats inside
the window return the original job's id with the original status.

## Authentication + per-key rate limits

Currently anyone reachable on the port can submit jobs. Minimum viable:
opaque API tokens checked at middleware, scoped per consumer; per-token
rate limit (token bucket, e.g. 60 req/min) and concurrent-job quota. Logs
must include the token id so abuse is traceable.

## URL-based input + signed-URL output

The API takes a server-local file path, which forces callers to drop files
on the box before submitting. For a pipeline that's a step nobody wants.

- Input: accept `source_url` (`http(s)://` or `s3://`); download to a scratch
  dir, transcribe, clean up. Bound max size + content-type, respect signed
  URL TTLs.
- Output: optionally PUT the result files (json/srt/vtt/txt/zip) to a
  caller-supplied signed URL or S3 prefix. The VOD service then never
  reads back through this API.

Once these land, the box is stateless from the caller's perspective.

## Subtitle line splitting (SRT/VTT)

The current SRT/VTT writers emit one cue per Whisper segment, which is
sentence-sized. Real VOD subtitles need cues of ≤42 chars, ≤2 lines, ≤6s.
Use the word timestamps to split each segment cleanly:

- Greedy fill: pack words into the current line until adding the next word
  would exceed `max_chars_per_line`; break, add a `\n`, fill again; flush
  the cue when `max_lines_per_cue` or `max_cue_duration_ms` is reached.
- Snap line breaks to whitespace; never split mid-word.
- Configurable per-call (`subtitle_options` on the job) so different
  consumers can ask for different rules.

This is the single highest-leverage output change for VOD use.

## Webhook signing + richer payload

`internal/callback` fires on status changes but the payload isn't signed
and is fairly bare. For pipeline integration:

- Sign with HMAC-SHA256 over the raw body using a per-consumer secret;
  include `X-Signature` + `X-Timestamp` headers and reject replays older
  than ~5 min on the receiver side.
- On `COMPLETED`, include the result inline (small) or a result URL the
  receiver can pull, plus `duration`, `language`, `model`, `word_count`.
- Retries: exponential backoff with a dead-letter after N attempts.

## OpenAPI spec

API-first means a published contract. Generate `openapi.yaml` from the
handlers (or hand-write, the surface is small) and serve it at
`/openapi.yaml`. Use it to generate clients for the VOD service in its
native language.

## Initial prompt / custom vocabulary

Whisper supports an `--initial_prompt` that biases the decoder toward
specific terminology. For domain content (proper nouns, theology terms in
the church-gathering corpus, brand names, host names on a recurring
podcast) this is a one-line CLI flag with a noticeable quality win.

Add an `initial_prompt` field to the job request and forward it to both
adapters. Per-API-key default prompts are a cheap extension that lets a
consumer set vocabulary once and forget.

## Forced alignment mode

If the caller already has a clean human-edited transcript (very common in
VOD — there's a script), aligning it to audio gives sub-100ms timestamps
with zero transcription errors. This is the killer feature for production
subtitle workflows.

Implementation options:
- Wrap WhisperX or stable-ts as a third adapter (`task=align`); request
  takes `audio` + `transcript`, returns word-level timestamps for the
  provided text.
- Build it natively against whisper.cpp's DTW (see DTW section below) —
  more work but no Python dependency.

## Speaker diarization

whisper.cpp `-tdrz` (tinydiarize, requires a `-tdrz` model) or pyannote
via faster-whisper labels speakers. Add a `speaker` field to `Segment`
and `Word`. Useful for multi-voice content (interviews, panels, church
services with multiple readers).

## Translation task

Whisper has a built-in `task=translate` that translates any source
language to English in one pass. For a multilingual content library this
eliminates a second processing step. Add a `task` field (`transcribe` /
`translate`) on the job request.

## Job timeouts + orphan-process cleanup

A hung `whisper-cli` today won't be killed unless someone cancels the
job manually. Two changes:

1. Wall-clock timeout per job (config + per-request override). On expiry,
   cancel context and mark the job `FAILED` with reason `timeout`.
2. Verify `exec.CommandContext` propagates SIGKILL to the whisper process
   tree on context cancel — set `Setpgid: true` and kill the process group
   on Unix to be safe.

## Prometheus metrics + structured JSON logging

For operability without poking the API:

- `/metrics` endpoint with queue depth gauge, job duration histogram
  (by model + status), job throughput counter, callback delivery
  success/failure counts.
- Switch the `slog` text handler to JSON, attach `job_id` (and
  `api_key_id` once auth lands) to every log line via context.

## Audio preprocessing pipeline

Right now each adapter does its own decoding. Centralize a preprocessing
step: normalize loudness, downmix to mono, resample to 16kHz, optional
silence trim. Adapters then receive a known-good wav and never see raw
mp3/m4a/etc. Also makes the chunked extractor's output the *only* shape
adapters need to support.

## Parallel chunk transcription

`chunked.Config.Parallelism` is wired but defaults to 1. On CPU
(whisper.cpp) or multi-GPU hosts, raising this gives substantial
wall-time improvement on long files. Needs measurement per backend
(whisper.cpp's internal threading vs. process-level concurrency) before
choosing a default > 1.

## Persistent model worker

Each `whisper-cli` / `whisper-ctranslate2` invocation reloads the model
from disk (~5s for large-v3). When chunking a 1-hour file into 12 chunks
that's a meaningful tax. Long-running per-model worker processes (whisper.cpp
ships a `server` binary; faster-whisper can wrap in a Python sidecar)
amortize the load cost across all chunks of a job and across jobs.

## Word-level timestamp accuracy via DTW (whisper.cpp adapter)

Word timestamps from the whisper.cpp adapter currently come from each token's
`offsets.from/to` in the `--output-json-full` output. These are derived from
the decoder's per-step time anchors, which can be coarse: it's common to see
several adjacent tokens share the same offset when the model emits them in a
single inference step.

whisper.cpp ships a more precise alternative: DTW (dynamic time warping) over
the cross-attention weights, enabled via `-dtw <preset>` on `whisper-cli`. When
enabled, each token gets a `t_dtw` value (milliseconds) that's typically
sub-100ms accurate.

To wire it up:

1. Add a `DTWModel string` field to `whispercpp.Config` and append
   `-dtw <DTWModel>` to the args in `Transcribe` when set.
2. Surface it via env (e.g. `WHISPER_CPP_DTW`) in `cmd/transcriber/models.go`,
   or derive the preset name from the model filename (`ggml-base.bin` →
   `base`, `ggml-large-v3.bin` → `large.v3` — note: dots, not dashes).
3. Update `tokensToWords` in `internal/transcriber/whispercpp/whispercpp.go`
   to prefer `t_dtw` for the word's start when it's not `-1`, falling back to
   `offsets.from` otherwise. End time can come from the next token's `t_dtw`
   (or the segment end for the last token).

Caveats:
- The preset name must match the model being loaded, or alignment is garbage.
- DTW adds a second pass over the attention weights — expect ~10–30% more
  inference time per file.
- Requires the model to have alignment heads. All official `ggml-*` Whisper
  models do; custom fine-tunes may not.

Worth doing if word timing drives UX (karaoke highlighting, click-to-seek on
a word). Skippable if word timestamps are just metadata.
