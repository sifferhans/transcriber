package fasterwhisper

import (
	"testing"

	"transcriber/internal/transcriber"
)

// rawJSON mimics whisper-ctranslate2's JSON output: top-level text +
// language, segments with the OpenAI-style fields plus a word array that
// uses "word"/"probability" (which we map onto Word.Text and discard).
const rawJSON = `{
  "text": " Hello world. Goodbye.",
  "language": "en",
  "segments": [
    {
      "id": 0,
      "start": 0.04,
      "end": 1.24,
      "text": " Hello world.",
      "tokens": [50364, 2425, 1002, 13],
      "temperature": 0.0,
      "avg_logprob": -0.3,
      "compression_ratio": 1.5,
      "no_speech_prob": 0.01,
      "words": [
        { "word": " Hello",  "start": 0.04, "end": 0.62, "probability": 0.95 },
        { "word": " world.", "start": 0.62, "end": 1.24, "probability": 0.88 }
      ]
    },
    {
      "id": 1,
      "start": 1.24,
      "end": 2.5,
      "text": " Goodbye.",
      "tokens": [50364, 1240, 13],
      "temperature": 0.0,
      "avg_logprob": -0.2,
      "compression_ratio": 1.1,
      "no_speech_prob": 0.02,
      "words": [
        { "word": " Goodbye.", "start": 1.24, "end": 2.5, "probability": 0.97 }
      ]
    }
  ]
}`

func TestParseJSON(t *testing.T) {
	tr, err := parseJSON([]byte(rawJSON))
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}

	if tr.Language != "en" {
		t.Errorf("language = %q, want %q", tr.Language, "en")
	}
	if want := "Hello world. Goodbye."; tr.Text != want {
		t.Errorf("text = %q, want %q", tr.Text, want)
	}
	if len(tr.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(tr.Segments))
	}

	// Spot-check the first segment: metadata fields preserved, words mapped
	// from "word"/"probability" to Word.Text (probability dropped).
	s0 := tr.Segments[0]
	if s0.ID != 0 || s0.Text != " Hello world." || s0.Start != 0.04 || s0.End != 1.24 {
		t.Errorf("segment[0] = %+v", s0)
	}
	if s0.Temperature != 0.0 || s0.AvgLogprob != -0.3 || s0.CompressionRatio != 1.5 || s0.NoSpeechProb != 0.01 {
		t.Errorf("segment[0] metadata not preserved: %+v", s0)
	}
	if len(s0.Tokens) != 4 || s0.Tokens[0] != 50364 {
		t.Errorf("segment[0] tokens = %v", s0.Tokens)
	}

	wantWords := []transcriber.Word{
		{Text: " Hello", Start: 0.04, End: 0.62},
		{Text: " world.", Start: 0.62, End: 1.24},
	}
	if len(s0.Words) != len(wantWords) {
		t.Fatalf("segment[0] words = %d, want %d", len(s0.Words), len(wantWords))
	}
	for i, w := range wantWords {
		if s0.Words[i] != w {
			t.Errorf("segment[0].words[%d] = %+v, want %+v", i, s0.Words[i], w)
		}
	}

	// Second segment: single word, sanity check the mapping holds.
	s1 := tr.Segments[1]
	if len(s1.Words) != 1 || s1.Words[0] != (transcriber.Word{Text: " Goodbye.", Start: 1.24, End: 2.5}) {
		t.Errorf("segment[1].words = %+v", s1.Words)
	}
}
