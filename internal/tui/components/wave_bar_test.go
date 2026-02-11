package components

import (
	"regexp"
	"strings"
	"testing"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestRenderWaveProgressBarFormat(t *testing.T) {
	t.Parallel()

	rendered := stripANSI(RenderWaveProgressBar(WaveProgressBarConfig{
		WaveNumber: 1,
		Completed:  2,
		Total:      4,
		Width:      20,
		Variant:    WaveProgressActive,
	}))

	if !strings.Contains(rendered, "Wave 1: [") {
		t.Fatalf("expected wave prefix in output, got %q", rendered)
	}
	if !strings.Contains(rendered, "] 2/4 done") {
		t.Fatalf("expected completion summary in output, got %q", rendered)
	}
}

func TestRenderWaveProgressBarClampsValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		config   WaveProgressBarConfig
		contains string
	}{
		{
			name: "completed over total is clamped",
			config: WaveProgressBarConfig{
				WaveNumber: 2,
				Completed:  7,
				Total:      5,
			},
			contains: "5/5 done",
		},
		{
			name: "negative completed is clamped to zero",
			config: WaveProgressBarConfig{
				WaveNumber: 2,
				Completed:  -2,
				Total:      5,
			},
			contains: "0/5 done",
		},
		{
			name: "negative total is clamped to zero",
			config: WaveProgressBarConfig{
				WaveNumber: 2,
				Completed:  3,
				Total:      -1,
			},
			contains: "0/0 done",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rendered := stripANSI(RenderWaveProgressBar(testCase.config))
			if !strings.Contains(rendered, testCase.contains) {
				t.Fatalf("expected output %q to contain %q", rendered, testCase.contains)
			}
		})
	}
}

func TestRenderWaveProgressBarDefaults(t *testing.T) {
	t.Parallel()

	rendered := stripANSI(RenderWaveProgressBar(WaveProgressBarConfig{}))
	if !strings.Contains(rendered, "Wave 1: [") {
		t.Fatalf("expected default wave number 1, got %q", rendered)
	}
	if !strings.Contains(rendered, "] 0/0 done") {
		t.Fatalf("expected default counts 0/0, got %q", rendered)
	}
}

func TestResolveWaveProgressVariant(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		variant   WaveProgressVariant
		completed int
		total     int
		want      WaveProgressVariant
	}{
		{
			name:      "explicit active",
			variant:   WaveProgressActive,
			completed: 1,
			total:     5,
			want:      WaveProgressActive,
		},
		{
			name:      "infer pending when zero total",
			completed: 0,
			total:     0,
			want:      WaveProgressPending,
		},
		{
			name:      "infer pending when nothing complete",
			completed: 0,
			total:     5,
			want:      WaveProgressPending,
		},
		{
			name:      "infer complete when complete",
			completed: 5,
			total:     5,
			want:      WaveProgressComplete,
		},
		{
			name:      "infer active while in progress",
			completed: 2,
			total:     5,
			want:      WaveProgressActive,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := resolveWaveProgressVariant(testCase.variant, testCase.completed, testCase.total); got != testCase.want {
				t.Fatalf("resolveWaveProgressVariant() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func stripANSI(input string) string {
	return ansiPattern.ReplaceAllString(input, "")
}
