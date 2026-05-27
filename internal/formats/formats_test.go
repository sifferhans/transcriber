package formats_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"transcriber/internal/formats"
	"transcriber/internal/transcriber"
)

// fixture is the Transcription that must serialize byte-for-byte to the
// golden files in testdata/golden/. Update both this fixture and the
// goldens together when changing the output shape.
func fixture() *transcriber.Transcription {
	return &transcriber.Transcription{
		Text:     "Hello world. This is a test of the transcription system. Goodbye.",
		Language: "en",
		Segments: []transcriber.Segment{
			{
				ID:    0,
				Text:  "Hello world.",
				Start: 0.04, End: 1.24,
				Words: []transcriber.Word{
					{Text: "Hello", Start: 0.04, End: 0.62},
					{Text: "world.", Start: 0.62, End: 1.24},
				},
			},
			{
				ID:    1,
				Text:  "This is a test of the transcription system.",
				Start: 1.24, End: 5.1,
				Words: []transcriber.Word{
					{Text: "This", Start: 1.24, End: 1.58},
					{Text: "is", Start: 1.58, End: 1.8},
					{Text: "a", Start: 1.8, End: 1.94},
					{Text: "test", Start: 1.94, End: 2.36},
					{Text: "of", Start: 2.36, End: 2.58},
					{Text: "the", Start: 2.58, End: 2.82},
					{Text: "transcription", Start: 2.82, End: 4.2},
					{Text: "system.", Start: 4.2, End: 5.1},
				},
			},
			{
				ID:    2,
				Text:  "Goodbye.",
				Start: 5.1, End: 6.45,
				Words: []transcriber.Word{
					{Text: "Goodbye.", Start: 5.1, End: 6.45},
				},
			},
		},
	}
}

func TestWriteMatchesGolden(t *testing.T) {
	cases := []struct {
		name   string
		format string
		golden string
	}{
		{"json", formats.JSON, "transcript.json"},
		{"srt", formats.SRT, "transcript.srt"},
		{"vtt", formats.VTT, "transcript.vtt"},
		{"txt", formats.TXT, "transcript.txt"},
	}

	tr := fixture()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			path, err := formats.Write(tc.format, tr, tmp)
			if err != nil {
				t.Fatalf("Write(%s): %v", tc.format, err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read written file: %v", err)
			}
			goldenPath := filepath.Join("testdata", "golden", tc.golden)
			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v", goldenPath, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("output mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.format, got, want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", formats.All()},
		{"all", formats.All()},
		{"json", []string{"json"}},
		{"json,srt", []string{"json", "srt"}},
		{" JSON , SRT ", []string{"json", "srt"}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := formats.Parse(tc.in)
			if !equalSlices(got, tc.want) {
				t.Errorf("Parse(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
