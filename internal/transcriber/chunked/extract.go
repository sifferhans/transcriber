package chunked

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"transcriber/internal/procutil"
)

// ProbeDuration returns the audio duration in seconds via ffprobe.
func ProbeDuration(ctx context.Context, ffprobeBin, path string) (float64, error) {
	if ffprobeBin == "" {
		ffprobeBin = "ffprobe"
	}
	cmd := exec.CommandContext(ctx, ffprobeBin,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	procutil.KillGroupOnCancel(cmd)
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return 0, fmt.Errorf("ffprobe %q: %w", path, err)
		}
		return 0, fmt.Errorf("ffprobe %q: %w: %s", path, err, msg)
	}
	s := strings.TrimSpace(string(out))
	d, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration parse %q: %w", s, err)
	}
	return d, nil
}

// ExtractChunk writes a 16kHz mono PCM wav of the range to outPath.
// `-ss` is placed after `-i` for accurate seek (fast-seek skews timing by up to hundreds of ms).
func ExtractChunk(ctx context.Context, ffmpegBin, inputPath, outPath string, startSec, durationSec float64) error {
	if ffmpegBin == "" {
		ffmpegBin = "ffmpeg"
	}
	cmd := exec.CommandContext(ctx, ffmpegBin,
		"-nostdin", "-loglevel", "error", "-y",
		"-i", inputPath,
		"-ss", formatSec(startSec),
		"-t", formatSec(durationSec),
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		outPath,
	)
	procutil.KillGroupOnCancel(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg extract: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func formatSec(s float64) string {
	return strconv.FormatFloat(s, 'f', 3, 64)
}
