package tui

import "github.com/ship-commander/sc3/internal/tui/views"

// DefaultViewDefinitions returns the baseline AppShell view map with Fleet Overview as entry.
func DefaultViewDefinitions() map[ViewID]ViewDefinition {
	return map[ViewID]ViewDefinition{
		ViewFleetOverview: {
			FocusOrder:  []string{"ship_list_panel", "preview_panel", "toolbar"},
			EnterTarget: "ship_bridge",
			Render: func(model AppModel) string {
				width, _ := model.Dimensions()
				if width == 0 {
					width = StandardLayoutMinWidth
				}
				return views.RenderFleetOverview(views.FleetOverviewConfig{
					Ships:              nil,
					SelectedIndex:      0,
					ToolbarHighlighted: 0,
					Width:              width,
					FleetHealthLabel:   "Optimal",
					PendingMessages:    0,
				})
			},
		},
	}
}

// NewDefaultAppModel constructs an AppModel with Fleet Overview as the initial screen.
func NewDefaultAppModel() *AppModel {
	return NewAppModel(ViewFleetOverview, DefaultViewDefinitions())
}
