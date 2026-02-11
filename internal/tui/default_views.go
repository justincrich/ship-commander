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
		ViewPlanReview: {
			FocusOrder: []string{"manifest_panel", "analysis_panel", "toolbar"},
			Render: func(model AppModel) string {
				width, _ := model.Dimensions()
				if width == 0 {
					width = StandardLayoutMinWidth
				}

				return views.RenderPlanReview(views.PlanReviewConfig{
					Width:          width,
					ShipName:       "USS Enterprise",
					DirectiveTitle: "Demonstrate plan review",
					Missions: []views.PlanReviewMission{
						{
							ID:             "M-001",
							Title:          "Prepare mission manifest",
							Classification: "STANDARD_OPS",
							Wave:           1,
							UseCaseRefs:    []string{"UC-TUI-01", "UC-TUI-03"},
							ACTotal:        3,
							SurfaceArea:    "internal/tui/views",
						},
						{
							ID:             "M-002",
							Title:          "Validate dependency graph",
							Classification: "RED_ALERT",
							Wave:           2,
							UseCaseRefs:    []string{"UC-TUI-15"},
							ACTotal:        2,
							SurfaceArea:    "internal/commander",
						},
					},
					Coverage: []views.PlanReviewCoverageRow{
						{UseCaseID: "UC-TUI-01", MissionIDs: []string{"M-001"}, Status: views.PlanReviewCoverageCovered},
						{UseCaseID: "UC-TUI-03", MissionIDs: []string{"M-001"}, Status: views.PlanReviewCoveragePartial},
						{UseCaseID: "UC-TUI-15", MissionIDs: nil, Status: views.PlanReviewCoverageUncovered},
					},
					Dependencies: []views.PlanReviewDependencyWave{
						{
							Wave: 1,
							Missions: []views.PlanReviewDependencyMission{
								{ID: "M-001", Title: "Prepare mission manifest", Status: "done"},
							},
						},
						{
							Wave: 2,
							Missions: []views.PlanReviewDependencyMission{
								{ID: "M-002", Title: "Validate dependency graph", Status: "waiting", Dependencies: []string{"M-001"}},
							},
						},
					},
					SignoffsDone:       2,
					SignoffsTotal:      3,
					AnalysisTab:        views.PlanReviewAnalysisCoverage,
					ToolbarHighlighted: 0,
				})
			},
		},
	}
}

// NewDefaultAppModel constructs an AppModel with Fleet Overview as the initial screen.
func NewDefaultAppModel() *AppModel {
	return NewAppModel(ViewFleetOverview, DefaultViewDefinitions())
}
