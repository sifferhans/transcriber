# Drop-in compatibility with `bcc-code/ai-api`

This service is intended as a drop-in replacement for the old Python-backed
[`bcc-code/ai-api`](https://github.com/bcc-code/ai-api) Whisper API. The
endpoint paths, HTTP verbs, and JSON field names match. What follows is the
delta between the two — things a legacy caller would hit if it pointed at
this server unchanged.

## Other behavioral differences (not yet addressed)

- **`result` field semantics.** Old: accumulated stdout text from the python child process (mostly debug noise; the README example shows a `"TEXT"` placeholder). New: filesystem path to `transcript.json`. The path form is more useful; if downstream callers parsed `result` as text they will need updating. Recommend keeping the new behavior and updating callers rather than reverting.
- **Callback firing on failure.** Old fired `doCallback` only on `COMPLETED`. New fires for `FAILED` and `CANCELED` as well. Receivers that don't filter on `status` may misbehave. Easy to gate behind config if it bites.
- **`output_path` is required.** Old accepted empty; new returns `400`. Probably an improvement; mention it in any migration note.
- **`format` field honoring.** Old python ignored `format` entirely and always wrote all five files. New honors `format` strictly (e.g. `format: "txt"` writes only `transcript.txt`). For full file-on-disk parity, callers that previously passed `format: "txt"` and then read the `.vtt` should switch to `format: "all"` (which now includes `words.srt`).
- **`list` cap.** Old kept only the last 10 processed jobs in memory. New keeps every job for the process lifetime. Differs in size only.
- **Status set additive.** New emits `CANCELED` (in addition to the old `QUEUED|RUNNING|COMPLETED|FAILED`); legacy receivers should treat unknown statuses defensively.

## Missing endpoints

- `GET /smi` — old shelled out to `nvidia-smi` and returned text. Only meaningful on GPU boxes; skip unless ops asks.
