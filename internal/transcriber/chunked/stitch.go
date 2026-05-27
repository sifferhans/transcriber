package chunked

import (
	"strings"

	"transcriber/internal/transcriber"
)

// stitch merges per-chunk Transcriptions back into one whose timestamps are
// in the original audio timeline. Each chunk owns the half-open interval
// `(prev_midpoint, next_midpoint]` of source seconds, where the midpoint is
// the middle of the overlap with the neighboring chunk; a segment is kept
// by the chunk whose owned interval contains its END timestamp. This makes
// every segment kept by exactly one chunk, including ones that straddle a
// midpoint.
//
// Word arrays inside kept segments are not re-trimmed; the segment is the
// atomic unit. This is good enough because chunks overlap by a few seconds
// and segments are sentence-sized — a sentence that straddles the midpoint
// stays in exactly one chunk.
func stitch(plan []Chunk, parts []*transcriber.Transcription) *transcriber.Transcription {
	if len(plan) == 0 || len(parts) == 0 {
		return &transcriber.Transcription{}
	}
	if len(plan) != len(parts) {
		panic("chunked.stitch: plan/parts length mismatch")
	}

	merged := &transcriber.Transcription{
		Language: parts[0].Language,
	}

	for i, part := range parts {
		if part == nil {
			continue
		}
		chunk := plan[i]
		var cutoff float64 // absolute time before which segments in this chunk are dropped
		var nextStart float64
		if i > 0 {
			// midpoint of overlap with previous chunk
			cutoff = (plan[i-1].End + chunk.Start) / 2
		}
		if i < len(plan)-1 {
			// midpoint of overlap with next chunk
			nextStart = (chunk.End + plan[i+1].Start) / 2
		} else {
			nextStart = chunk.End
		}

		for _, seg := range part.Segments {
			absStart := seg.Start + chunk.Start
			absEnd := seg.End + chunk.Start
			if absEnd <= cutoff {
				continue // belongs to previous chunk
			}
			if absEnd > nextStart {
				continue // belongs to next chunk
			}
			s := transcriber.Segment{
				ID:    len(merged.Segments),
				Text:  seg.Text,
				Start: absStart,
				End:   absEnd,
			}
			for _, w := range seg.Words {
				s.Words = append(s.Words, transcriber.Word{
					Text:  w.Text,
					Start: w.Start + chunk.Start,
					End:   w.End + chunk.Start,
				})
			}
			merged.Segments = append(merged.Segments, s)
		}
	}

	// Rebuild top-level Text from kept segments — same convention adapters use.
	var sb strings.Builder
	for i, s := range merged.Segments {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(strings.TrimSpace(s.Text))
	}
	merged.Text = sb.String()

	return merged
}
