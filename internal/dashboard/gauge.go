package dashboard

import (
	"fmt"
	"strings"
)

const (
	gaugeFillChar  = '█'
	gaugeEmptyChar = '░'
)

// RenderGauge renders an ASCII health gauge bar with score label.
func RenderGauge(score, width int) string {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	if width < 1 {
		width = 1
	}

	filled := score * width / 100
	empty := width - filled

	bar := strings.Repeat(string(gaugeFillChar), filled) + strings.Repeat(string(gaugeEmptyChar), empty)

	return fmt.Sprintf("%s %d/100", bar, score)
}

// gaugeColor returns a color name based on the health score.
func gaugeColor(score int) string {
	switch {
	case score >= 75:
		return "green"
	case score >= 50:
		return "yellow"
	default:
		return "red"
	}
}
