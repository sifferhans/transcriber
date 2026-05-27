package whispercpp

import (
	"testing"

	"transcriber/internal/transcriber"
)

// rawJSON mimics what `whisper-cli --output-json-full` writes: two segments,
// the second one containing sub-word pieces, attached punctuation, and a
// trailing special token (id >= firstSpecialTokenID) that must be skipped.
const rawJSON = `{
  "result": { "language": "en" },
  "transcription": [
    {
      "offsets": { "from": 0, "to": 1240 },
      "text": " Hello world.",
      "tokens": [
        { "text": " Hello", "offsets": { "from": 40,  "to": 620  }, "id": 12842 },
        { "text": " world", "offsets": { "from": 620, "to": 1100 }, "id": 1002  },
        { "text": ".",      "offsets": { "from": 1100,"to": 1240 }, "id": 13    }
      ]
    },
    {
      "offsets": { "from": 1240, "to": 2500 },
      "text": " Transcribing test.",
      "tokens": [
        { "text": " Trans",   "offsets": { "from": 1240, "to": 1480 }, "id": 8757  },
        { "text": "cribing",  "offsets": { "from": 1480, "to": 1940 }, "id": 9123  },
        { "text": " test",    "offsets": { "from": 1940, "to": 2360 }, "id": 1332  },
        { "text": ".",        "offsets": { "from": 2360, "to": 2500 }, "id": 13    },
        { "text": "[_EOT_]",  "offsets": { "from": 2500, "to": 2500 }, "id": 50257 }
      ]
    }
  ]
}`

func TestParseJSON(t *testing.T) {
	tr, err := parseJSON([]byte(rawJSON), "")
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}

	if tr.Language != "en" {
		t.Errorf("language = %q, want %q", tr.Language, "en")
	}
	if want := "Hello world. Transcribing test."; tr.Text != want {
		t.Errorf("text = %q, want %q", tr.Text, want)
	}
	if len(tr.Segments) != 2 {
		t.Fatalf("segments = %d, want 2", len(tr.Segments))
	}

	want := []transcriber.Segment{
		{
			ID:    0,
			Text:  "Hello world.",
			Start: 0.0, End: 1.24,
			Words: []transcriber.Word{
				{Text: "Hello", Start: 0.04, End: 0.62},
				{Text: "world.", Start: 0.62, End: 1.24},
			},
		},
		{
			ID:    1,
			Text:  "Transcribing test.",
			Start: 1.24, End: 2.5,
			Words: []transcriber.Word{
				// Sub-word pieces "Trans" + "cribing" merge into one word;
				// end time stretches to cover both tokens.
				{Text: "Transcribing", Start: 1.24, End: 1.94},
				// Trailing "." attaches to "test"; [_EOT_] is dropped.
				{Text: "test.", Start: 1.94, End: 2.5},
			},
		},
	}
	for i, w := range want {
		got := tr.Segments[i]
		if got.ID != w.ID || got.Text != w.Text || got.Start != w.Start || got.End != w.End {
			t.Errorf("segment[%d] = {id:%d text:%q start:%v end:%v}, want {id:%d text:%q start:%v end:%v}",
				i, got.ID, got.Text, got.Start, got.End, w.ID, w.Text, w.Start, w.End)
		}
		if len(got.Words) != len(w.Words) {
			t.Errorf("segment[%d] words len = %d, want %d", i, len(got.Words), len(w.Words))
			continue
		}
		for j, wantWord := range w.Words {
			if got.Words[j] != wantWord {
				t.Errorf("segment[%d].words[%d] = %+v, want %+v", i, j, got.Words[j], wantWord)
			}
		}
	}
}

func TestParseJSONLanguageFallback(t *testing.T) {
	raw := `{"result": {"language": ""}, "transcription": []}`
	tr, err := parseJSON([]byte(raw), "no")
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}
	if tr.Language != "no" {
		t.Errorf("language = %q, want fallback %q", tr.Language, "no")
	}
}

func TestTokensToWordsEmpty(t *testing.T) {
	if got := tokensToWords(nil); got != nil {
		t.Errorf("tokensToWords(nil) = %v, want nil", got)
	}
}
