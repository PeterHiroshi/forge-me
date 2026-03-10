package dashboard

import (
	"fmt"
	"strings"
	"testing"
)

func TestRenderGauge(t *testing.T) {
	tests := []struct {
		name     string
		score    int
		width    int
		wantFill int
	}{
		{"zero score", 0, 20, 0},
		{"full score", 100, 20, 20},
		{"half score", 50, 20, 10},
		{"quarter score", 25, 40, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderGauge(tt.score, tt.width)
			if got == "" {
				t.Error("expected non-empty gauge string")
			}
			if !strings.Contains(got, fmt.Sprintf("%d", tt.score)) {
				t.Errorf("gauge should contain score %d, got %q", tt.score, got)
			}
		})
	}
}

func TestRenderGaugeClampsValues(t *testing.T) {
	got := RenderGauge(150, 20)
	if got == "" {
		t.Error("expected non-empty gauge for clamped value")
	}

	got = RenderGauge(-5, 20)
	if got == "" {
		t.Error("expected non-empty gauge for negative value")
	}
}

func TestGaugeColor(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{95, "green"},
		{80, "green"},
		{60, "yellow"},
		{30, "red"},
		{10, "red"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("score_%d", tt.score), func(t *testing.T) {
			got := gaugeColor(tt.score)
			if got != tt.want {
				t.Errorf("gaugeColor(%d) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}
