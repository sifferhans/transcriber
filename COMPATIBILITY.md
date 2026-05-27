# Drop-in compatibility with the legacy Whisper API

This service is intended as a drop-in replacement for the old Python-backed
Whisper API. The endpoint paths, HTTP verbs, and JSON field names match.
What follows is the delta between the two — things a legacy caller would
hit if it pointed at this server unchanged.

## Other behavioral differences (not yet addressed)

Verified against the current caller — none of the items below produce a
real failure. They're documented for context if a future caller appears.

- **`result` field semantics.** Old: accumulated stdout from the python child process (mostly debug noise). New: filesystem path to `transcript.json`. The caller never reads `result`, so this is invisible.
- **Callback firing on failure.** Old fired `doCallback` only on `COMPLETED`. New fires for `FAILED` and `CANCELED` as well. The caller doesn't register a callback (it polls), so this is invisible.
- **`output_path` is required.** Old accepted empty; new returns `400`. The caller always sets it.
- **`format` field honoring.** Old python ignored `format` entirely and always wrote all five files. New honors `format` strictly. The caller always sends `format: "all"`, which writes all five (now including `words.srt`).
- **`list` cap.** Old kept only the last 10 processed jobs in memory. New keeps every job for the process lifetime. Differs in size only.
- **Status set additive.** New emits `CANCELED` (in addition to the old `QUEUED|RUNNING|COMPLETED|FAILED`); the caller's status switch falls through unknowns and keeps polling.

## Missing endpoints

- `GET /smi` — old shelled out to `nvidia-smi` and returned text. Only meaningful on GPU boxes; skip unless ops asks.
