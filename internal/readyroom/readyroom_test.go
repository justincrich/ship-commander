package readyroom

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/commission"
)

func TestPlanSpawnsThreeSessionsWithCommissionContext(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
			RoleCommander: {
				1: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: false}},
				},
				2: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
			RoleDesignOfficer: {
				1: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: false}},
				},
				2: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
		},
	}

	room := newReadyRoomForTest(t, factory, 5)
	result, err := room.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if len(factory.spawnRequests) != 3 {
		t.Fatalf("spawn requests = %d, want 3", len(factory.spawnRequests))
	}
	for _, request := range factory.spawnRequests {
		if request.Commission.ID != "COMM-1" {
			t.Fatalf("commission id = %q, want COMM-1", request.Commission.ID)
		}
	}

	if !result.Consensus {
		t.Fatal("consensus = false, want true")
	}
	if result.Iterations != 2 {
		t.Fatalf("iterations = %d, want 2", result.Iterations)
	}

	for _, role := range requiredRoles {
		session := factory.sessionsByRole[role]
		if len(session.inputs) == 0 {
			t.Fatalf("session %s received no input", role)
		}
		for _, input := range session.inputs {
			if input.Commission.ID != "COMM-1" {
				t.Fatalf("session %s commission = %q, want COMM-1", role, input.Commission.ID)
			}
		}
	}
}

func TestPlanRoutesStructuredMessagesThroughOrchestrator(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {
					Messages: []ReadyRoomMessage{{
						To:      string(RoleCommander),
						Type:    "analysis",
						Domain:  "functional",
						Content: "captain->commander",
					}},
				},
				2: {
					Messages: []ReadyRoomMessage{{
						To:      "broadcast",
						Type:    "feedback",
						Domain:  "functional",
						Content: "captain-broadcast",
					}},
				},
			},
			RoleCommander: {
				2: {
					Messages: []ReadyRoomMessage{{
						To:      string(RoleDesignOfficer),
						Type:    "analysis",
						Domain:  "technical",
						Content: "commander->design",
					}},
				},
			},
			RoleDesignOfficer: {},
		},
	}

	room := newReadyRoomForTest(t, factory, 3)
	fixedTime := time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC)
	room.now = func() time.Time { return fixedTime }

	result, err := room.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if result.Consensus {
		t.Fatal("consensus = true, want false")
	}

	commanderInputs := factory.sessionsByRole[RoleCommander].inputs
	if len(commanderInputs) < 1 {
		t.Fatalf("commander inputs = %d, want at least 1", len(commanderInputs))
	}
	if got := len(commanderInputs[0].Inbox); got != 1 {
		t.Fatalf("commander inbox entries iteration 1 = %d, want 1", got)
	}
	if commanderInputs[0].Inbox[0].Content != "captain->commander" {
		t.Fatalf("commander inbox content = %q, want captain->commander", commanderInputs[0].Inbox[0].Content)
	}

	designInputs := factory.sessionsByRole[RoleDesignOfficer].inputs
	if len(designInputs) < 2 {
		t.Fatalf("design inputs = %d, want at least 2", len(designInputs))
	}
	if got := len(designInputs[1].Inbox); got == 0 {
		t.Fatal("design inbox at iteration 2 = empty, want routed messages")
	}

	for _, message := range result.Messages {
		if message.From == "" {
			t.Fatal("message from is empty")
		}
		if message.Timestamp.IsZero() {
			t.Fatal("message timestamp is zero")
		}
	}
}

func TestPlanStopsAtMaxIterationsWithoutConsensus(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: true}}},
				2: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: true}}},
			},
			RoleCommander: {
				1: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: false}}},
				2: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: false}}},
			},
			RoleDesignOfficer: {
				1: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: true}}},
				2: {Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1"}, SignOff: true}}},
			},
		},
	}

	room := newReadyRoomForTest(t, factory, 2)
	result, err := room.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	if result.Consensus {
		t.Fatal("consensus = true, want false")
	}
	if result.Iterations != 2 {
		t.Fatalf("iterations = %d, want 2", result.Iterations)
	}
}

func TestBuildUseCaseCoverageTracksCoveredPartialUncovered(t *testing.T) {
	t.Parallel()

	room := &ReadyRoom{
		commission: commission.Commission{
			ID: "COMM-1",
			UseCases: []commission.UseCase{
				{ID: "UC-1"},
				{ID: "UC-2"},
				{ID: "UC-3"},
			},
		},
		missionPlan: map[string]*MissionPlan{
			"M-COVERED": {
				ID:         "M-COVERED",
				UseCaseIDs: []string{"UC-1"},
				Signoffs: MissionSignoffs{
					Captain:       true,
					Commander:     true,
					DesignOfficer: true,
				},
			},
			"M-PARTIAL": {
				ID:         "M-PARTIAL",
				UseCaseIDs: []string{"UC-2"},
				Signoffs: MissionSignoffs{
					Captain:       true,
					Commander:     false,
					DesignOfficer: true,
				},
			},
		},
	}

	coverage := room.BuildUseCaseCoverage()

	if coverage["UC-1"] != CoverageCovered {
		t.Fatalf("UC-1 coverage = %q, want %q", coverage["UC-1"], CoverageCovered)
	}
	if coverage["UC-2"] != CoveragePartial {
		t.Fatalf("UC-2 coverage = %q, want %q", coverage["UC-2"], CoveragePartial)
	}
	if coverage["UC-3"] != CoverageUncovered {
		t.Fatalf("UC-3 coverage = %q, want %q", coverage["UC-3"], CoverageUncovered)
	}
}

func TestPlanReturnsErrorForUnknownMessageRecipient(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {
					Messages: []ReadyRoomMessage{{
						To:      "unknown-role",
						Type:    "analysis",
						Domain:  "functional",
						Content: "bad recipient",
					}},
				},
			},
			RoleCommander:     {},
			RoleDesignOfficer: {},
		},
	}

	room := newReadyRoomForTest(t, factory, 1)
	_, err := room.Plan(context.Background())
	if err == nil {
		t.Fatal("expected routing error")
	}
	if want := "unknown recipient"; !contains(err.Error(), want) {
		t.Fatalf("error %q missing %q", err.Error(), want)
	}
}

func TestNewValidatesInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		factory   SessionFactory
		comm      commission.Commission
		wantError bool
	}{
		{
			name:      "missing factory",
			factory:   nil,
			comm:      commission.Commission{ID: "COMM-1"},
			wantError: true,
		},
		{
			name:      "missing commission id",
			factory:   &fakeFactory{},
			comm:      commission.Commission{},
			wantError: true,
		},
		{
			name:      "valid",
			factory:   &fakeFactory{},
			comm:      commission.Commission{ID: "COMM-1"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			room, err := New(tt.factory, tt.comm, 0)
			if tt.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("new ready room: %v", err)
			}
			if room.maxIterations != DefaultMaxIterations {
				t.Fatalf("maxIterations = %d, want %d", room.maxIterations, DefaultMaxIterations)
			}
		})
	}
}

func newReadyRoomForTest(t *testing.T, factory *fakeFactory, maxIterations int) *ReadyRoom {
	t.Helper()

	room, err := New(
		factory,
		commission.Commission{
			ID:       "COMM-1",
			UseCases: []commission.UseCase{{ID: "UC-1"}, {ID: "UC-2"}},
		},
		maxIterations,
	)
	if err != nil {
		t.Fatalf("new ready room: %v", err)
	}
	return room
}

type fakeFactory struct {
	scripts        map[AgentRole]map[int]SessionOutput
	spawnRequests  []SpawnRequest
	sessionsByRole map[AgentRole]*fakeSession
	spawnErr       error
}

func (f *fakeFactory) Spawn(_ context.Context, request SpawnRequest) (Session, error) {
	if f.spawnErr != nil {
		return nil, f.spawnErr
	}

	if f.sessionsByRole == nil {
		f.sessionsByRole = make(map[AgentRole]*fakeSession)
	}

	f.spawnRequests = append(f.spawnRequests, request)
	session := &fakeSession{
		id:      fmt.Sprintf("session-%s", request.Role),
		scripts: f.scripts[request.Role],
	}
	f.sessionsByRole[request.Role] = session

	return session, nil
}

type fakeSession struct {
	id       string
	scripts  map[int]SessionOutput
	inputs   []SessionInput
	closeErr error
}

func (s *fakeSession) ID() string {
	return s.id
}

func (s *fakeSession) Execute(_ context.Context, input SessionInput) (SessionOutput, error) {
	s.inputs = append(s.inputs, input)
	if s.scripts == nil {
		return SessionOutput{}, nil
	}
	out, ok := s.scripts[input.Iteration]
	if !ok {
		return SessionOutput{}, nil
	}
	return out, nil
}

func (s *fakeSession) Close(_ context.Context) error {
	if s.closeErr != nil {
		return s.closeErr
	}
	return nil
}

func contains(value, substr string) bool {
	return strings.Contains(value, substr)
}
