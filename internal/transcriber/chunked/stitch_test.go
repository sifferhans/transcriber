package chunked

import (
	"testing"

	"transcriber/internal/transcriber"
)

// TestStitchOffsetsAndDedups builds three chunks with deliberately
// overlapping segments around the midpoint and verifies:
//   - timestamps are shifted into the original timeline
//   - segments inside the overlap are claimed by exactly one chunk
//   - segment IDs are renumbered contiguously
//   - top-level text is rebuilt from the kept segments
func TestStitchOffsetsAndDedups(t *testing.T) {
	plan := []Chunk{
		{Index: 0, Start: 0, End: 10},
		{Index: 1, Start: 7, End: 17},   // overlaps prev by 3s, midpoint=8.5
		{Index: 2, Start: 14, End: 20},  // overlaps prev by 3s, midpoint=15.5
	}
	parts := []*transcriber.Transcription{
		{
			Language: "en",
			Segments: []transcriber.Segment{
				// Local times within chunk 0 ([0, 10] of source).
				seg("alpha", 0, 4),
				seg("beta", 4, 8), // abs ends at 8, before midpoint 8.5 → kept by chunk 0
				seg("gamma", 8, 10), // abs starts at 8, ends at 10 — straddles midpoint 8.5 → dropped here (chunk 1 keeps it)
			},
		},
		{
			Language: "en",
			Segments: []transcriber.Segment{
				// Local times within chunk 1 ([7, 17] of source).
				seg("gamma", 1, 3),   // abs 8–10, after midpoint 8.5 → kept here
				seg("delta", 3, 7),   // abs 10–14 → kept (clearly in chunk 1's zone)
				seg("epsilon", 7, 10), // abs 14–17, straddles midpoint 15.5 → dropped (chunk 2 keeps)
			},
		},
		{
			Language: "en",
			Segments: []transcriber.Segment{
				// Local times within chunk 2 ([14, 20] of source).
				seg("epsilon", 1, 3),  // abs 15–17, after midpoint 15.5 → kept
				seg("zeta", 3, 6),     // abs 17–20 → kept
			},
		},
	}

	got := stitch(plan, parts)
	if got.Language != "en" {
		t.Errorf("language = %q", got.Language)
	}
	wantTexts := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"}
	if len(got.Segments) != len(wantTexts) {
		t.Fatalf("segments = %d, want %d (texts: %+v)", len(got.Segments), len(wantTexts), segTexts(got.Segments))
	}
	for i, want := range wantTexts {
		if got.Segments[i].Text != want {
			t.Errorf("segment[%d] = %q, want %q", i, got.Segments[i].Text, want)
		}
		if got.Segments[i].ID != i {
			t.Errorf("segment[%d].ID = %d, want %d (renumbering)", i, got.Segments[i].ID, i)
		}
	}
	// Top-level text rebuilt as space-joined segments.
	if want := "alpha beta gamma delta epsilon zeta"; got.Text != want {
		t.Errorf("text = %q, want %q", got.Text, want)
	}
	// Timestamps shifted into source timeline.
	if got.Segments[0].Start != 0 || got.Segments[0].End != 4 {
		t.Errorf("alpha = %.2f–%.2f, want 0–4", got.Segments[0].Start, got.Segments[0].End)
	}
	if got.Segments[5].Start != 17 || got.Segments[5].End != 20 {
		t.Errorf("zeta = %.2f–%.2f, want 17–20", got.Segments[5].Start, got.Segments[5].End)
	}
}

func TestStitchWordsOffset(t *testing.T) {
	plan := []Chunk{{Index: 0, Start: 100, End: 110}}
	parts := []*transcriber.Transcription{
		{
			Segments: []transcriber.Segment{
				{
					Text:  "hi there",
					Start: 0, End: 2,
					Words: []transcriber.Word{
						{Text: "hi", Start: 0, End: 1},
						{Text: "there", Start: 1, End: 2},
					},
				},
			},
		},
	}
	got := stitch(plan, parts)
	if len(got.Segments) != 1 || len(got.Segments[0].Words) != 2 {
		t.Fatalf("unexpected shape: %+v", got)
	}
	w := got.Segments[0].Words
	if w[0].Start != 100 || w[0].End != 101 {
		t.Errorf("word[0] = %.2f–%.2f, want 100–101", w[0].Start, w[0].End)
	}
	if w[1].Start != 101 || w[1].End != 102 {
		t.Errorf("word[1] = %.2f–%.2f, want 101–102", w[1].Start, w[1].End)
	}
}

func TestStitchEmpty(t *testing.T) {
	got := stitch(nil, nil)
	if got == nil || got.Text != "" || len(got.Segments) != 0 {
		t.Errorf("stitch(nil, nil) = %+v, want empty", got)
	}
}

func seg(text string, start, end float64) transcriber.Segment {
	return transcriber.Segment{Text: text, Start: start, End: end}
}

func segTexts(segs []transcriber.Segment) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = s.Text
	}
	return out
}
