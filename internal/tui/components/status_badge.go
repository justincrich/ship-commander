package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ship-commander/sc3/internal/tui/theme"
)

// BadgeOpt configures optional rendering behavior for StatusBadge.
type BadgeOpt func(*badgeOptions)

type badgeOptions struct {
	showIcon bool
	bold     bool
}

type badgeVariant struct {
	icon  string
	label string
	color lipgloss.TerminalColor
}

var statusBadgeVariants = map[string]badgeVariant{
	"running": {
		icon:  theme.IconWorking,
		label: "RUNNING",
		color: theme.ButterscotchColor,
	},
	"done": {
		icon:  theme.IconDone,
		label: "DONE",
		color: theme.GreenOkColor,
	},
	"waiting": {
		icon:  theme.IconWaiting,
		label: "WAITING",
		color: theme.BlueColor,
	},
	"skipped": {
		icon:  theme.IconSkipped,
		label: "SKIPPED",
		color: theme.GalaxyGrayColor,
	},
	"failed": {
		icon:  theme.IconFailed,
		label: "FAILED",
		color: theme.RedAlertColor,
	},
	"stuck": {
		icon:  theme.IconWorking,
		label: "STUCK",
		color: theme.YellowCautionColor,
	},
	"halted": {
		icon:  theme.IconFailed,
		label: "HALTED",
		color: theme.RedAlertColor,
	},
	"planning": {
		icon:  theme.IconWorking,
		label: "PLANNING",
		color: theme.PurpleColor,
	},
	"review": {
		icon:  theme.IconWorking,
		label: "REVIEW",
		color: theme.PurpleColor,
	},
	"approved": {
		icon:  theme.IconDone,
		label: "APPROVED",
		color: theme.GreenOkColor,
	},
}

// WithBadgeIcon controls whether the icon is shown (default: true).
func WithBadgeIcon(show bool) BadgeOpt {
	return func(options *badgeOptions) {
		options.showIcon = show
	}
}

// WithBadgeBold controls whether the badge text is bold (default: false).
func WithBadgeBold(bold bool) BadgeOpt {
	return func(options *badgeOptions) {
		options.bold = bold
	}
}

// RenderStatusBadge renders `[icon] LABEL` with semantic LCARS color styling.
func RenderStatusBadge(status string, opts ...BadgeOpt) string {
	options := badgeOptions{
		showIcon: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	variant, ok := statusBadgeVariants[normalizedStatus]
	if !ok {
		variant = badgeVariant{
			icon:  theme.IconAlert,
			label: strings.ToUpper(strings.TrimSpace(status)),
			color: theme.GalaxyGrayColor,
		}
		if variant.label == "" {
			variant.label = "UNKNOWN"
		}
	}

	content := variant.label
	if options.showIcon {
		content = variant.icon + " " + variant.label
	}

	return lipgloss.NewStyle().
		Foreground(variant.color).
		Bold(options.bold).
		Render(content)
}
