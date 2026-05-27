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

// Format identifiers used in the API's `format` field. Each identifier is
// also used directly as the output-file extension: e.g. `format: "words.srt"`
// writes `<basename>.words.srt`.
const (
	JSON     = "json"
	SRT      = "srt"
	VTT      = "vtt"
	TXT      = "txt"
	WordsSRT = "words.srt"
)

// All returns the formats produced when the request specifies "all". Mirrors
// the five files the legacy Python API always wrote regardless of `format`.
func All() []string { return []string{JSON, SRT, VTT, WordsSRT, TXT} }

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

// Write serializes the transcription to outDir/<basename>.<format>. Returns
// the path of the written file. The basename comes from the input file
// (e.g. "track_109473_media_en.mp3"), matching the legacy Python API's
// `<basename>.<ext>` layout. Caller is responsible for ensuring outDir
// exists; Write will create it if missing.
func Write(format string, t *transcriber.Transcription, outDir, basename string) (string, error) {
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
		body = []byte(toSRT(t))
	case VTT:
		body = []byte(toVTT(t))
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

// toWordsSRT emits one SRT cue per word using word-level timestamps,
// numbered monotonically 1..N. Drop-in counterpart to the legacy Python
// API's `<basename>.words.srt` output.
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
