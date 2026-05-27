package chunked

import (
	"strings"

	"transcriber/internal/transcriber"
)

// stitch merges per-chunk Transcriptions onto the source timeline.
// Each chunk owns segments whose END falls in (prev_midpoint, next_midpoint].
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
		var cutoff, nextStart float64
		if i > 0 {
			cutoff = (plan[i-1].End + chunk.Start) / 2
		}
		if i < len(plan)-1 {
			nextStart = (chunk.End + plan[i+1].Start) / 2
		} else {
			nextStart = chunk.End
		}

		for _, seg := range part.Segments {
			absStart := seg.Start + chunk.Start
			absEnd := seg.End + chunk.Start
			if absEnd <= cutoff || absEnd > nextStart {
				continue
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
