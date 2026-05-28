package formats

import (
	"math"
	"strings"

	"transcriber/internal/transcriber"
)

// SubtitleOptions controls how segments are split into broadcast-grade cues.
// Zero values fall back to broadcast defaults (≤42 chars, ≤2 lines).
type SubtitleOptions struct {
	MaxCharsPerLine int `json:"max_chars_per_line,omitempty"`
	MaxLinesPerCue  int `json:"max_lines_per_cue,omitempty"`
}

func (o SubtitleOptions) normalized() SubtitleOptions {
	if o.MaxCharsPerLine <= 0 {
		o.MaxCharsPerLine = 42
	}
	if o.MaxLinesPerCue <= 0 {
		o.MaxLinesPerCue = 2
	}
	return o
}

// Cue is a single subtitle event ready for an SRT/VTT writer.
type Cue struct {
	Start float64
	End   float64
	Lines []string
}

type cueWord struct {
	text       string
	start, end float64
}

// BuildCues splits each segment into cues using word timestamps. Segments
// without words fall back to a single cue per segment.
func BuildCues(t *transcriber.Transcription, opts SubtitleOptions) []Cue {
	opts = opts.normalized()

	var cues []Cue
	for _, seg := range t.Segments {
		if len(seg.Words) == 0 {
			cues = append(cues, Cue{
				Start: seg.Start,
				End:   seg.End,
				Lines: []string{strings.TrimSpace(seg.Text)},
			})
			continue
		}
		cues = append(cues, splitSegment(seg, opts)...)
	}
	return cues
}

func splitSegment(seg transcriber.Segment, opts SubtitleOptions) []Cue {
	ws := make([]cueWord, 0, len(seg.Words))
	for _, w := range seg.Words {
		text := strings.TrimSpace(w.Text)
		if text == "" {
			continue
		}
		ws = append(ws, cueWord{text: text, start: w.Start, end: w.End})
	}
	if len(ws) == 0 {
		return nil
	}

	var cues []Cue
	i := 0
	for i < len(ws) {
		// Greedy: find the largest j such that ws[i:j+1] still fits the cue.
		maxJ := i
		for j := i; j < len(ws); j++ {
			if tryWrap(ws[i:j+1], opts.MaxLinesPerCue, opts.MaxCharsPerLine) != nil {
				maxJ = j
			} else {
				break
			}
		}
		end := maxJ

		// If the remaining tail would be an orphan fragment, back off to the
		// last natural break (sentence end, then clause break) within the cue.
		// Skip if doing so leaves the current cue too small.
		if maxJ+1 < len(ws) && isOrphan(ws[maxJ+1:]) {
			if k := lastTerminalIdx(ws[i : maxJ+1]); k > 0 && cueSubstantial(ws[i:i+k+1]) {
				end = i + k
			} else if k := lastClauseIdx(ws[i : maxJ+1]); k > 0 && cueSubstantial(ws[i:i+k+1]) {
				end = i + k
			}
		}

		cueWords := ws[i : end+1]
		cues = append(cues, Cue{
			Start: cueWords[0].start,
			End:   cueWords[len(cueWords)-1].end,
			Lines: wrapCue(cueWords, opts.MaxLinesPerCue, opts.MaxCharsPerLine),
		})
		i = end + 1
	}
	return cues
}

// tryWrap returns lines if the words fit within maxLines × maxChars; otherwise
// nil. Lines containing a single word are allowed to overflow maxChars.
func tryWrap(words []cueWord, maxLines, maxChars int) []string {
	if len(words) == 0 {
		return nil
	}
	text := joinWords(words)
	if len(text) <= maxChars || len(words) == 1 {
		return []string{text}
	}
	if maxLines >= 2 {
		if lines := tryWrapTwoLines(words, maxChars); lines != nil {
			return lines
		}
	}
	if maxLines > 2 {
		lines := greedyWrap(words, maxChars)
		if len(lines) <= maxLines {
			return lines
		}
	}
	return nil
}

// tryWrapTwoLines finds the most balanced 2-line split where both lines fit
// in maxChars (single-word lines may overflow). Breaks after sentence-ending
// or clause-ending punctuation are preferred. Returns nil if no split fits.
func tryWrapTwoLines(words []cueWord, maxChars int) []string {
	bestI := -1
	bestScore := math.MaxInt
	for i := 1; i < len(words); i++ {
		l1 := joinWords(words[:i])
		l2 := joinWords(words[i:])
		if i > 1 && len(l1) > maxChars {
			continue
		}
		if len(words)-i > 1 && len(l2) > maxChars {
			continue
		}
		score := abs(len(l1) - len(l2))
		switch lastRune(words[i-1].text) {
		case '.', '!', '?':
			score -= 200
		case ',', ';', ':':
			score -= 100
		}
		if score < bestScore {
			bestScore = score
			bestI = i
		}
	}
	if bestI < 0 {
		return nil
	}
	return []string{joinWords(words[:bestI]), joinWords(words[bestI:])}
}

// wrapCue produces the final lines for a cue. Falls back to greedy wrap if a
// balanced layout isn't found.
func wrapCue(words []cueWord, maxLines, maxChars int) []string {
	if lines := tryWrap(words, maxLines, maxChars); lines != nil {
		return lines
	}
	return greedyWrap(words, maxChars)
}

// greedyWrap fills each line up to maxChars left-to-right.
func greedyWrap(words []cueWord, maxChars int) []string {
	var lines []string
	var cur string
	for _, w := range words {
		switch {
		case cur == "":
			cur = w.text
		case len(cur)+1+len(w.text) > maxChars:
			lines = append(lines, cur)
			cur = w.text
		default:
			cur += " " + w.text
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

func isOrphan(tail []cueWord) bool {
	if len(tail) > 2 {
		return false
	}
	total := 0
	for _, w := range tail {
		total += len(w.text)
	}
	if len(tail) > 1 {
		total += len(tail) - 1
	}
	return total <= 20
}

func cueSubstantial(words []cueWord) bool {
	if len(words) < 2 {
		return false
	}
	total := len(words) - 1
	for _, w := range words {
		total += len(w.text)
	}
	return total >= 12
}

func lastTerminalIdx(words []cueWord) int {
	for i := len(words) - 2; i >= 0; i-- {
		switch lastRune(words[i].text) {
		case '.', '!', '?':
			return i
		}
	}
	return -1
}

func lastClauseIdx(words []cueWord) int {
	for i := len(words) - 2; i >= 0; i-- {
		switch lastRune(words[i].text) {
		case ',', ';', ':':
			return i
		}
	}
	return -1
}

func lastRune(s string) rune {
	if s == "" {
		return 0
	}
	rs := []rune(s)
	return rs[len(rs)-1]
}

func joinWords(words []cueWord) string {
	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = w.text
	}
	return strings.Join(parts, " ")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
