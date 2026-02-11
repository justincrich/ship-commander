package tui

import "github.com/ship-commander/sc3/internal/tui/views"

// DefaultViewDefinitions returns the baseline AppShell view map with Fleet Overview as entry.
func DefaultViewDefinitions() map[ViewID]ViewDefinition {
	return map[ViewID]ViewDefinition{
		ViewFleetOverview: {
			FocusOrder:  []string{"ship_list_panel", "preview_panel", "toolbar"},
			EnterTarget: ViewShipBridge,
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
		ViewShipBridge: {
			FocusOrder: []string{"crew_panel", "mission_board", "event_log", "toolbar"},
			Render: func(model AppModel) string {
				width, _ := model.Dimensions()
				if width == 0 {
					width = StandardLayoutMinWidth
				}

				return views.RenderShipBridge(views.ShipBridgeConfig{
					Width:            width,
					ShipName:         "USS Enterprise",
					ShipClass:        "Galaxy-class",
					DirectiveTitle:   "Demonstrate ship bridge",
					Status:           views.ShipBridgeStatusDocked,
					FleetHealthLabel: "Optimal",
					WaveCurrent:      1,
					WaveTotal:        1,
					MissionsDone:     0,
					MissionsTotal:    1,
					Crew: []views.ShipBridgeCrewMember{
						{Name: "Riker", Role: "Captain", MissionID: "M-001", Phase: "PLANNING", Elapsed: "00:42", Status: "waiting"},
						{Name: "Data", Role: "Commander", MissionID: "", Phase: "IDLE", Elapsed: "00:00", Status: "waiting"},
					},
					Missions: []views.ShipBridgeMission{
						{ID: "M-001", Title: "Prepare launch checklist", Column: "backlog", Classification: "STANDARD_OPS", AssignedAgent: "Riker", Phase: "PLANNING", ACCompleted: 0, ACTotal: 3},
					},
					Events: []views.ShipBridgeEvent{
						{Timestamp: "09:00:00", Severity: "info", Actor: "system", Message: "Ship bridge ready"},
					},
				})
			},
		},
	}
}

// NewDefaultAppModel constructs an AppModel with Fleet Overview as the initial screen.
func NewDefaultAppModel() *AppModel {
	return NewAppModel(ViewFleetOverview, DefaultViewDefinitions())
}
