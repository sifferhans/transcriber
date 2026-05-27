package chunked

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"transcriber/internal/transcriber"
)

// fakeInner is an inner Transcriber that records the inputs it sees and
// returns a single-segment Transcription whose text encodes the chunk's
// duration. We use that to assert the stitcher ran on the right inputs.
type fakeInner struct {
	calls atomic.Int32
}

func (f *fakeInner) ID() string   { return "fake" }
func (f *fakeInner) Name() string { return "fake" }

func (f *fakeInner) Transcribe(ctx context.Context, req transcriber.Request, onProgress transcriber.ProgressFunc) (*transcriber.Result, error) {
	idx := f.calls.Add(1) - 1
	// Probe the wav we just received to make sure ExtractChunk wrote something
	// real (16kHz mono pcm, decodable by ffprobe).
	d, err := ProbeDuration(ctx, "", req.InputPath)
	if err != nil {
		return nil, fmt.Errorf("fakeInner probe: %w", err)
	}
	if onProgress != nil {
		onProgress(1.0)
	}
	// Produce a short segment centered in the chunk so it doesn't span the
	// overlap region on either side — real ASR emits sentence-sized
	// segments, not one chunk-long segment.
	mid := d / 2
	half := d / 8
	return &transcriber.Result{
		Transcription: &transcriber.Transcription{
			Language: "en",
			Text:     fmt.Sprintf("chunk-%d", idx),
			Segments: []transcriber.Segment{
				{Text: fmt.Sprintf("chunk-%d", idx), Start: mid - half, End: mid + half},
			},
		},
	}, nil
}

func TestAdapterEndToEnd(t *testing.T) {
	requireBins(t, "ffmpeg", "ffprobe")

	tmp := t.TempDir()
	// 6-second silent wav. Long enough that Plan with chunkLen=2,
	// overlap=0.5 produces 4 chunks.
	src := filepath.Join(tmp, "src.wav")
	gen := exec.Command("ffmpeg", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "anullsrc=channel_layout=mono:sample_rate=16000",
		"-t", "6", src,
	)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg generate: %v: %s", err, out)
	}

	outDir := filepath.Join(tmp, "out")
	inner := &fakeInner{}
	a := New(inner, Config{ChunkLengthSec: 2, OverlapSec: 0.5})

	var progressLast float64
	res, err := a.Transcribe(context.Background(), transcriber.Request{
		InputPath: src,
		OutputDir: outDir,
	}, func(p float64) {
		progressLast = p
	})
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if got := inner.calls.Load(); got < 3 {
		t.Errorf("inner calls = %d, want >= 3", got)
	}
	tr := res.Transcription
	if tr.Language != "en" {
		t.Errorf("language = %q", tr.Language)
	}
	if len(tr.Segments) < 3 {
		t.Fatalf("segments = %d, want >= 3", len(tr.Segments))
	}
	// Segments span the source timeline in order.
	for i := 1; i < len(tr.Segments); i++ {
		prev, cur := tr.Segments[i-1], tr.Segments[i]
		if cur.Start < prev.End-1e-6 {
			t.Errorf("segment %d starts at %.3f, before prev ends at %.3f", i, cur.Start, prev.End)
		}
	}
	// First segment starts near the beginning, last ends within the audio.
	if tr.Segments[0].Start < 0 {
		t.Errorf("first segment Start = %.2f, want >= 0", tr.Segments[0].Start)
	}
	if tr.Segments[len(tr.Segments)-1].End > 6.1 {
		t.Errorf("last segment End = %.2f, want <= 6", tr.Segments[len(tr.Segments)-1].End)
	}
	// Segment IDs renumbered.
	for i, s := range tr.Segments {
		if s.ID != i {
			t.Errorf("segment[%d].ID = %d, want %d", i, s.ID, i)
		}
	}
	// Progress should reach ~1.0 by completion.
	if progressLast < 0.99 {
		t.Errorf("final progress = %.3f, want >= 0.99", progressLast)
	}
}

func TestAdapterShortFileBypassesChunking(t *testing.T) {
	requireBins(t, "ffmpeg", "ffprobe")

	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.wav")
	gen := exec.Command("ffmpeg", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "anullsrc=channel_layout=mono:sample_rate=16000",
		"-t", "1", src,
	)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg generate: %v: %s", err, out)
	}

	inner := &fakeInner{}
	a := New(inner, Config{ChunkLengthSec: 300, OverlapSec: 3})
	_, err := a.Transcribe(context.Background(), transcriber.Request{
		InputPath: src,
		OutputDir: filepath.Join(tmp, "out"),
	}, nil)
	if err != nil {
		t.Fatalf("Transcribe: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Errorf("inner calls = %d, want 1 (short-file bypass)", got)
	}
}

func requireBins(t *testing.T, names ...string) {
	t.Helper()
	for _, n := range names {
		if _, err := exec.LookPath(n); err != nil {
			t.Skipf("%s not on PATH: %v", n, err)
		}
	}
}
