# Potential improvements

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

## Per-word confidence field

The whisper.cpp adapter now emits a `confidence` value per word (mean of the
contributing tokens' `p`). The reference format in `test/formats/transcript.json`
doesn't include this field — it's currently kept because it's free information
that downstream consumers may find useful, and the field is `omitempty` so
zero-confidence words drop it cleanly. If it turns out to be noise, remove the
`Confidence` field from `transcriber.Word` and stop populating it in
`tokensToWords`.
