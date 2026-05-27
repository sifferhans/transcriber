package chunked

import "testing"

func TestPlanShortFile(t *testing.T) {
	// Anything shorter than chunkLen → one chunk spanning the whole audio.
	plan := Plan(120, 300, 3)
	if len(plan) != 1 {
		t.Fatalf("len(plan) = %d, want 1", len(plan))
	}
	if plan[0].Start != 0 || plan[0].End != 120 {
		t.Errorf("plan[0] = %+v, want [0, 120]", plan[0])
	}
}

func TestPlanExactlyChunkLen(t *testing.T) {
	plan := Plan(300, 300, 3)
	if len(plan) != 1 {
		t.Fatalf("len(plan) = %d, want 1", len(plan))
	}
}

func TestPlanLongFile(t *testing.T) {
	// 30 min audio, 5 min chunks, 3s overlap → stride 297s.
	plan := Plan(1800, 300, 3)
	if len(plan) < 6 {
		t.Fatalf("len(plan) = %d, want >= 6", len(plan))
	}
	// Adjacent chunks overlap by ~3s.
	for i := 1; i < len(plan); i++ {
		got := plan[i-1].End - plan[i].Start
		if got < 2.9 || got > 3.1 {
			// Last chunk may have different overlap if it's truncated.
			if i == len(plan)-1 {
				continue
			}
			t.Errorf("overlap between chunk %d/%d = %.2f, want ~3", i-1, i, got)
		}
	}
	// Last chunk ends at the audio duration.
	last := plan[len(plan)-1]
	if last.End != 1800 {
		t.Errorf("last chunk End = %.2f, want 1800", last.End)
	}
	// First chunk starts at 0.
	if plan[0].Start != 0 {
		t.Errorf("first chunk Start = %.2f, want 0", plan[0].Start)
	}
	// Coverage: every second in [0, 1800) is inside at least one chunk.
	for t1 := 0.0; t1 < 1800; t1 += 30 {
		covered := false
		for _, c := range plan {
			if t1 >= c.Start && t1 < c.End {
				covered = true
				break
			}
		}
		if !covered {
			t.Errorf("t=%.0f not covered by any chunk", t1)
		}
	}
}

func TestPlanZeroDuration(t *testing.T) {
	if plan := Plan(0, 300, 3); plan != nil {
		t.Errorf("Plan(0) = %v, want nil", plan)
	}
}

func TestPlanIndexesContiguous(t *testing.T) {
	plan := Plan(900, 300, 5)
	for i, c := range plan {
		if c.Index != i {
			t.Errorf("plan[%d].Index = %d, want %d", i, c.Index, i)
		}
	}
}
