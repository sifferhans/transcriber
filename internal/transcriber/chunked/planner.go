package chunked

// Chunk describes one slice of the input audio in original-timeline seconds.
// Adjacent chunks overlap by Config.OverlapSec so word-boundary cuts at chunk
// joins are recoverable during stitching.
type Chunk struct {
	Index int
	Start float64 // seconds from the start of the source audio (inclusive)
	End   float64 // seconds from the start of the source audio (exclusive)
}

// Duration returns End-Start in seconds.
func (c Chunk) Duration() float64 { return c.End - c.Start }

// Plan splits a total audio duration into overlapping chunks of approximately
// chunkLen seconds each, with overlapSec of overlap between adjacent chunks.
// The last chunk is truncated to the audio end. If duration <= chunkLen,
// Plan returns a single chunk spanning [0, duration).
//
// Inputs are clamped to sane minimums (chunkLen > overlapSec > 0).
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
