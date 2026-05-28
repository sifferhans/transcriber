package whispercpp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"transcriber/internal/formats"
)

// goldenDir is shared with internal/formats; parsed whisper-cli output must round-trip byte-identical.
var goldenDir = filepath.Join("..", "..", "formats", "testdata", "golden")

func TestParseJSONMatchesGoldens(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "raw.json"))
	if err != nil {
		t.Fatalf("read raw: %v", err)
	}
	tr, err := parseJSON(raw, "")
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}

	cases := []struct {
		name   string
		format string
		golden string
	}{
		{"json", formats.JSON, "transcript.json"},
		{"srt", formats.SRT, "transcript.srt"},
		{"vtt", formats.VTT, "transcript.vtt"},
		{"txt", formats.TXT, "transcript.txt"},
		{"words.srt", formats.WordsSRT, "transcript.words.srt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			path, err := formats.Write(tc.format, tr, tmp, "transcript", formats.SubtitleOptions{})
			if err != nil {
				t.Fatalf("Write(%s): %v", tc.format, err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			want, err := os.ReadFile(filepath.Join(goldenDir, tc.golden))
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.format, got, want)
			}
		})
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

// Whisper.cpp BPE splits "å" into two byte-level tokens (0xC3 and 0xA5).
// Each byte is invalid UTF-8 on its own; rawString must preserve the raw
// bytes so concatenation reconstitutes the codepoint.
func TestParseJSONReconstructsSplitUTF8Tokens(t *testing.T) {
	raw := []byte("{" +
		"\"result\":{\"language\":\"no\"}," +
		"\"transcription\":[{" +
		"\"offsets\":{\"from\":0,\"to\":1000}," +
		"\"text\":\" Så\"," +
		"\"tokens\":[" +
		"{\"text\":\" S\",\"offsets\":{\"from\":0,\"to\":200},\"id\":1}," +
		"{\"text\":\"\xC3\",\"offsets\":{\"from\":200,\"to\":400},\"id\":2}," +
		"{\"text\":\"\xA5\",\"offsets\":{\"from\":400,\"to\":600},\"id\":3}" +
		"]" +
		"}]" +
		"}")

	tr, err := parseJSON(raw, "no")
	if err != nil {
		t.Fatalf("parseJSON: %v", err)
	}
	if len(tr.Segments) != 1 {
		t.Fatalf("segments = %d, want 1", len(tr.Segments))
	}
	if got := tr.Segments[0].Text; got != "Så" {
		t.Errorf("segment text = %q, want %q", got, "Så")
	}
	words := tr.Segments[0].Words
	if len(words) != 1 {
		t.Fatalf("words = %d (%+v), want 1", len(words), words)
	}
	if words[0].Text != "Så" {
		t.Errorf("word text = %q (bytes %x), want %q", words[0].Text, []byte(words[0].Text), "Så")
	}
}
