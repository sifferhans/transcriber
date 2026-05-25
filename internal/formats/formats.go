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

// Format identifiers used in the API's `format` field.
const (
	JSON = "json"
	SRT  = "srt"
	VTT  = "vtt"
	TXT  = "txt"
)

// All returns the formats produced when the request specifies "all".
func All() []string { return []string{JSON, SRT, VTT, TXT} }

// Parse turns the API `format` field into a list of concrete formats.
// "" and "all" both expand to All().
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

// Write serializes the transcription to outDir/transcript.<format>. Returns
// the path of the written file. Caller is responsible for ensuring outDir
// exists; Write will create it if missing.
func Write(format string, t *transcriber.Transcription, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	var (
		filename string
		body     []byte
		err      error
	)
	switch format {
	case JSON:
		filename = "transcript.json"
		body, err = json.MarshalIndent(t, "", "  ")
	case SRT:
		filename = "transcript.srt"
		body = []byte(toSRT(t))
	case VTT:
		filename = "transcript.vtt"
		body = []byte(toVTT(t))
	case TXT:
		filename = "transcript.txt"
		body = []byte(strings.TrimSpace(t.Text) + "\n")
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return "", err
	}
	path := filepath.Join(outDir, filename)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func toSRT(t *transcriber.Transcription) string {
	var sb strings.Builder
	for i, seg := range t.Segments {
		fmt.Fprintf(&sb, "%d\n%s --> %s\n%s\n\n",
			i+1,
			srtTime(seg.Start),
			srtTime(seg.End),
			strings.TrimSpace(seg.Text),
		)
	}
	return sb.String()
}

func toVTT(t *transcriber.Transcription) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")
	for _, seg := range t.Segments {
		fmt.Fprintf(&sb, "%s --> %s\n%s\n\n",
			vttTime(seg.Start),
			vttTime(seg.End),
			strings.TrimSpace(seg.Text),
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
