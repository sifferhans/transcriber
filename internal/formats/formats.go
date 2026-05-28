package formats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"transcriber/internal/transcriber"
)

// Format identifiers double as output-file extensions.
const (
	JSON     = "json"
	SRT      = "srt"
	VTT      = "vtt"
	TXT      = "txt"
	WordsSRT = "words.srt"
)

// All returns the formats produced when the request specifies "all".
func All() []string { return []string{JSON, SRT, VTT, WordsSRT, TXT} }

// Parse turns the `format` field into a list of formats; "" and "all" expand to All().
func Parse(spec string) []string {
	switch strings.ToLower(strings.TrimSpace(spec)) {
	case "", "all":
		return All()
	}
	parts := strings.Split(spec, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Write serializes the transcription to outDir/<basename>.<format>.
// SubtitleOptions tunes SRT/VTT cue splitting; zero values use broadcast defaults.
func Write(format string, t *transcriber.Transcription, outDir, basename string, opts SubtitleOptions) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	var (
		body []byte
		err  error
	)
	switch format {
	case JSON:
		body, err = json.MarshalIndent(t, "", "  ")
		if err == nil {
			body = append(body, '\n')
		}
	case SRT:
		body = []byte(toSRT(t, opts))
	case VTT:
		body = []byte(toVTT(t, opts))
	case TXT:
		body = []byte(strings.TrimSpace(t.Text) + "\n")
	case WordsSRT:
		body = []byte(toWordsSRT(t))
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return "", err
	}
	path := filepath.Join(outDir, basename+"."+format)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func toSRT(t *transcriber.Transcription, opts SubtitleOptions) string {
	var sb strings.Builder
	for i, cue := range BuildCues(t, opts) {
		fmt.Fprintf(&sb, "%d\n%s --> %s\n%s\n\n",
			i+1,
			srtTime(cue.Start),
			srtTime(cue.End),
			strings.Join(cue.Lines, "\n"),
		)
	}
	return sb.String()
}

func toWordsSRT(t *transcriber.Transcription) string {
	var sb strings.Builder
	n := 0
	for _, seg := range t.Segments {
		for _, w := range seg.Words {
			n++
			fmt.Fprintf(&sb, "%d\n%s --> %s\n%s\n\n",
				n,
				srtTime(w.Start),
				srtTime(w.End),
				strings.TrimSpace(w.Text),
			)
		}
	}
	return sb.String()
}

func toVTT(t *transcriber.Transcription, opts SubtitleOptions) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for _, cue := range BuildCues(t, opts) {
		fmt.Fprintf(&sb, "%s --> %s\n%s\n\n",
			vttTime(cue.Start),
			vttTime(cue.End),
			strings.Join(cue.Lines, "\n"),
		)
	}
	return sb.String()
}

func srtTime(s float64) string {
	if s < 0 {
		s = 0
	}
	d := time.Duration(s * float64(time.Second))
	h := int(d / time.Hour)
	d -= time.Duration(h) * time.Hour
	m := int(d / time.Minute)
	d -= time.Duration(m) * time.Minute
	sec := int(d / time.Second)
	d -= time.Duration(sec) * time.Second
	ms := int(d / time.Millisecond)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, sec, ms)
}

func vttTime(s float64) string {
	return strings.Replace(srtTime(s), ",", ".", 1)
}
