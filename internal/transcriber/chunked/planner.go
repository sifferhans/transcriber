package chunked

// Chunk is one slice of source audio in original-timeline seconds; adjacent chunks overlap.
type Chunk struct {
	Index int
	Start float64
	End   float64
}

func (c Chunk) Duration() float64 { return c.End - c.Start }

// Plan splits duration into overlapping chunks; returns a single chunk if duration <= chunkLen.
func Plan(duration, chunkLen, overlapSec float64) []Chunk {
	if duration <= 0 {
		return nil
	}
	if chunkLen <= 0 {
		chunkLen = 300
	}
	if overlapSec < 0 {
		overlapSec = 0
	}
	if overlapSec >= chunkLen {
		overlapSec = chunkLen / 10
	}
	if duration <= chunkLen {
		return []Chunk{{Index: 0, Start: 0, End: duration}}
	}

	stride := chunkLen - overlapSec
	var out []Chunk
	for i, start := 0, 0.0; start < duration; i, start = i+1, start+stride {
		end := start + chunkLen
		if end >= duration {
			end = duration
			out = append(out, Chunk{Index: i, Start: start, End: end})
			break
		}
		out = append(out, Chunk{Index: i, Start: start, End: end})
	}
	return out
}
