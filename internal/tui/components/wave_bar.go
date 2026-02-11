package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

const defaultWaveBarWidth = 20

// WaveProgressVariant controls the visual state of the rendered wave bar.
type WaveProgressVariant string

const (
	// WaveProgressActive is the active wave variant (butterscotch -> gold gradient).
	WaveProgressActive WaveProgressVariant = "active"
	// WaveProgressComplete is the complete wave variant (solid green).
	WaveProgressComplete WaveProgressVariant = "complete"
	// WaveProgressPending is the pending wave variant (dim gray).
	WaveProgressPending WaveProgressVariant = "pending"
)

// WaveProgressBarConfig contains all rendering inputs for a wave progress bar.
type WaveProgressBarConfig struct {
	WaveNumber int
	Completed  int
	Total      int
	Width      int
	Variant    WaveProgressVariant
}

// RenderWaveProgressBar renders "Wave N: [bar] X/Y done" using bubbles/progress.
func RenderWaveProgressBar(config WaveProgressBarConfig) string {
	waveNumber, completed, total, width := normalizeWaveProgressConfig(config)
	variant := resolveWaveProgressVariant(config.Variant, completed, total)

	progressFraction := 0.0
	if total > 0 {
		progressFraction = float64(completed) / float64(total)
	}

	bar := newWaveProgressModel(width, variant).ViewAs(progressFraction)
	if variant == WaveProgressPending {
		bar = lipgloss.NewStyle().Faint(true).Render(bar)
	}

	return fmt.Sprintf("Wave %d: [%s] %d/%d done", waveNumber, bar, completed, total)
}

func normalizeWaveProgressConfig(config WaveProgressBarConfig) (int, int, int, int) {
	waveNumber := config.WaveNumber
	if waveNumber <= 0 {
		waveNumber = 1
	}

	total := config.Total
	if total < 0 {
		total = 0
	}

	completed := config.Completed
	if completed < 0 {
		completed = 0
	}
	if completed > total {
		completed = total
	}

	width := config.Width
	if width <= 0 {
		width = defaultWaveBarWidth
	}

	return waveNumber, completed, total, width
}

func resolveWaveProgressVariant(variant WaveProgressVariant, completed int, total int) WaveProgressVariant {
	switch variant {
	case WaveProgressActive, WaveProgressComplete, WaveProgressPending:
		return variant
	}

	if total == 0 || completed == 0 {
		return WaveProgressPending
	}
	if completed >= total {
		return WaveProgressComplete
	}
	return WaveProgressActive
}

func newWaveProgressModel(width int, variant WaveProgressVariant) progress.Model {
	options := []progress.Option{
		progress.WithWidth(width),
		progress.WithoutPercentage(),
		progress.WithFillCharacters('#', '.'),
	}

	switch variant {
	case WaveProgressComplete:
		options = append(options, progress.WithSolidFill(theme.GreenOk))
	case WaveProgressPending:
		options = append(options, progress.WithSolidFill(theme.GalaxyGray))
	default:
		options = append(options, progress.WithScaledGradient(theme.Butterscotch, theme.Gold))
	}

	return progress.New(options...)
}
