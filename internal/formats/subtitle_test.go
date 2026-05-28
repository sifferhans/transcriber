package formats

import (
	"reflect"
	"testing"

	"transcriber/internal/transcriber"
)

func words(specs ...wordSpec) []transcriber.Word {
	out := make([]transcriber.Word, 0, len(specs))
	for _, s := range specs {
		out = append(out, transcriber.Word{Text: s.text, Start: s.start, End: s.end})
	}
	return out
}

type wordSpec struct {
	text       string
	start, end float64
}

func TestBuildCues_LineWrapWithinCue(t *testing.T) {
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 0, End: 5,
		Words: words(
			wordSpec{"This", 0, 0.3},
			wordSpec{"is", 0.3, 0.5},
			wordSpec{"a", 0.5, 0.6},
			wordSpec{"test", 0.6, 1.0},
			wordSpec{"of", 1.0, 1.2},
			wordSpec{"the", 1.2, 1.4},
			wordSpec{"transcription", 1.4, 2.8},
			wordSpec{"system.", 2.8, 5.0},
		),
	}}}

	cues := BuildCues(tr, SubtitleOptions{})
	want := []Cue{{
		Start: 0, End: 5.0,
		Lines: []string{"This is a test of the", "transcription system."},
	}}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}

func TestBuildCues_FlushOnMaxLines(t *testing.T) {
	// Three lines forces a second cue at maxLines=2.
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 0, End: 9,
		Words: words(
			wordSpec{"aaaaaaaa", 0, 1},
			wordSpec{"bbbbbbbb", 1, 2},
			wordSpec{"cccccccc", 2, 3},
			wordSpec{"dddddddd", 3, 4},
			wordSpec{"eeeeeeee", 4, 5},
			wordSpec{"ffffffff", 5, 6},
		),
	}}}

	cues := BuildCues(tr, SubtitleOptions{MaxCharsPerLine: 8, MaxLinesPerCue: 2})
	want := []Cue{
		{Start: 0, End: 2, Lines: []string{"aaaaaaaa", "bbbbbbbb"}},
		{Start: 2, End: 4, Lines: []string{"cccccccc", "dddddddd"}},
		{Start: 4, End: 6, Lines: []string{"eeeeeeee", "ffffffff"}},
	}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}

func TestBuildCues_NoWordsFallsBackToSegment(t *testing.T) {
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 1.0, End: 4.5,
		Text:  "Hello there friend.",
	}}}

	cues := BuildCues(tr, SubtitleOptions{})
	want := []Cue{{Start: 1.0, End: 4.5, Lines: []string{"Hello there friend."}}}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}

func TestBuildCues_AvoidsOrphanWithClauseBreak(t *testing.T) {
	// "aaa bbb ccc, ddd eee fff ggg hhh" — fits in 2 lines × 13 chars only if we
	// break after the comma. The tail "hhh" alone would otherwise be an orphan.
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 0, End: 8,
		Words: words(
			wordSpec{"aaa", 0, 1},
			wordSpec{"bbb", 1, 2},
			wordSpec{"ccc,", 2, 3},
			wordSpec{"ddd", 3, 4},
			wordSpec{"eee", 4, 5},
			wordSpec{"fff", 5, 6},
			wordSpec{"ggg", 6, 7},
			wordSpec{"hhh", 7, 8},
		),
	}}}

	cues := BuildCues(tr, SubtitleOptions{MaxCharsPerLine: 13, MaxLinesPerCue: 2})
	want := []Cue{
		{Start: 0, End: 3, Lines: []string{"aaa bbb ccc,"}},
		{Start: 3, End: 8, Lines: []string{"ddd eee", "fff ggg hhh"}},
	}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}

func TestBuildCues_BalancedTwoLineSplit(t *testing.T) {
	// Greedy fill would put "Goodbye." on its own short line; balanced wrap
	// distributes words more evenly across the two lines.
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 0, End: 6,
		Words: words(
			wordSpec{"Hello", 0, 1},
			wordSpec{"world", 1, 2},
			wordSpec{"this", 2, 3},
			wordSpec{"is", 3, 4},
			wordSpec{"a", 4, 4.3},
			wordSpec{"farewell.", 4.3, 6},
		),
	}}}

	cues := BuildCues(tr, SubtitleOptions{MaxCharsPerLine: 20, MaxLinesPerCue: 2})
	want := []Cue{{
		Start: 0, End: 6,
		Lines: []string{"Hello world this", "is a farewell."},
	}}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}

func TestBuildCues_LongWordFitsOnOwnLine(t *testing.T) {
	tr := &transcriber.Transcription{Segments: []transcriber.Segment{{
		Start: 0, End: 2,
		Words: words(
			wordSpec{"supercalifragilisticexpialidocious", 0, 1.5},
			wordSpec{"ok.", 1.5, 2.0},
		),
	}}}

	cues := BuildCues(tr, SubtitleOptions{MaxCharsPerLine: 10, MaxLinesPerCue: 2})
	want := []Cue{{
		Start: 0, End: 2.0,
		Lines: []string{"supercalifragilisticexpialidocious", "ok."},
	}}
	if !reflect.DeepEqual(cues, want) {
		t.Errorf("got %+v, want %+v", cues, want)
	}
}
