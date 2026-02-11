package readyroom

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/commission"
	"github.com/ship-commander/sc3/internal/events"
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

//nolint:gocyclo // Comprehensive end-to-end behavior test for question suspension and answer routing.
func TestPlanSuspendsOnQuestionPublishesEventAndRoutesAnswer(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {
					Questions: []admiral.AdmiralQuestion{{
						QuestionID:    "Q-1",
						Domain:        "functional",
						QuestionText:  "Should this mission proceed?",
						Options:       []string{"Proceed", "Hold"},
						AllowFreeText: true,
					}},
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
			RoleCommander: {
				1: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
			RoleDesignOfficer: {
				1: {
					Missions: []MissionContribution{{MissionID: "M-1", UseCaseIDs: []string{"UC-1", "UC-2"}, SignOff: true}},
				},
			},
		},
	}

	room := newReadyRoomForTest(t, factory, 2)
	eventBus := &captureBus{}
	if err := room.SetEventBus(eventBus); err != nil {
		t.Fatalf("set event bus: %v", err)
	}

	answerRelease := make(chan struct{})
	questionSeen := make(chan admiral.AdmiralQuestion, 1)
	go func() {
		question := <-room.QuestionGate().Questions()
		questionSeen <- question
		<-answerRelease

		if err := room.QuestionGate().SubmitAnswer(admiral.AdmiralAnswer{
			QuestionID:     question.QuestionID,
			SelectedOption: "Proceed",
		}); err != nil {
			panic(err)
		}
	}()

	resultCh := make(chan PlanResult, 1)
	errCh := make(chan error, 1)
	go func() {
		result, err := room.Plan(context.Background())
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- result
	}()

	var asked admiral.AdmiralQuestion
	select {
	case asked = <-questionSeen:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for surfaced admiral question")
	}

	if asked.QuestionID != "Q-1" {
		t.Fatalf("question id = %q, want Q-1", asked.QuestionID)
	}
	if got := len(factory.sessionsByRole[RoleCommander].inputs); got != 0 {
		t.Fatalf("commander inputs before answer = %d, want 0 (planning should be suspended)", got)
	}

	close(answerRelease)

	var result PlanResult
	select {
	case err := <-errCh:
		t.Fatalf("plan: %v", err)
	case result = <-resultCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for plan result")
	}

	if !result.Consensus {
		t.Fatal("consensus = false, want true")
	}

	foundQuestionEvent := false
	for _, event := range eventBus.snapshot() {
		if event.Type == events.EventTypeAdmiralQuestion && event.EntityID == "Q-1" {
			foundQuestionEvent = true
			break
		}
	}
	if !foundQuestionEvent {
		t.Fatal("expected AdmiralQuestion event to be published")
	}

	answerRouted := false
	for _, message := range result.Messages {
		if message.From == "admiral" && message.To == string(RoleCaptain) && strings.Contains(message.Content, "question_id=Q-1") {
			answerRouted = true
			break
		}
	}
	if !answerRouted {
		t.Fatal("expected admiral answer message routed back to captain")
	}

	if len(result.QuestionLog) != 1 {
		t.Fatalf("question log entries = %d, want 1", len(result.QuestionLog))
	}
	if result.QuestionLog[0].QuestionID != "Q-1" {
		t.Fatalf("question log id = %q, want Q-1", result.QuestionLog[0].QuestionID)
	}
	if result.QuestionLog[0].Answer.SelectedOption != "Proceed" {
		t.Fatalf(
			"question log selected option = %q, want Proceed",
			result.QuestionLog[0].Answer.SelectedOption,
		)
	}
}

func TestPlanBroadcastsAdmiralAnswerWhenRequested(t *testing.T) {
	t.Parallel()

	factory := &fakeFactory{
		scripts: map[AgentRole]map[int]SessionOutput{
			RoleCaptain: {
				1: {
					Questions: []admiral.AdmiralQuestion{{
						QuestionID:     "Q-2",
						Domain:         "technical",
						QuestionText:   "Should this answer be broadcast?",
						Options:        []string{"Yes", "No"},
						AllowBroadcast: true,
						AllowFreeText:  true,
					}},
				},
			},
			RoleCommander:     {},
			RoleDesignOfficer: {},
		},
	}

	room := newReadyRoomForTest(t, factory, 1)
	answerErrCh := make(chan error, 1)

	go func() {
		question := <-room.QuestionGate().Questions()
		answerErrCh <- room.QuestionGate().SubmitAnswer(admiral.AdmiralAnswer{
			QuestionID: question.QuestionID,
			SkipFlag:   true,
			Broadcast:  true,
		})
	}()

	result, err := room.Plan(context.Background())
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	commanderInput := factory.sessionsByRole[RoleCommander].inputs
	if len(commanderInput) != 1 {
		t.Fatalf("commander inputs = %d, want 1", len(commanderInput))
	}
	if len(commanderInput[0].Inbox) != 1 {
		t.Fatalf("commander inbox entries = %d, want 1", len(commanderInput[0].Inbox))
	}
	if commanderInput[0].Inbox[0].From != "admiral" {
		t.Fatalf("commander inbox sender = %q, want admiral", commanderInput[0].Inbox[0].From)
	}

	designInput := factory.sessionsByRole[RoleDesignOfficer].inputs
	if len(designInput) != 1 {
		t.Fatalf("design inputs = %d, want 1", len(designInput))
	}
	if len(designInput[0].Inbox) != 1 {
		t.Fatalf("design inbox entries = %d, want 1", len(designInput[0].Inbox))
	}
	if designInput[0].Inbox[0].From != "admiral" {
		t.Fatalf("design inbox sender = %q, want admiral", designInput[0].Inbox[0].From)
	}

	if len(result.QuestionLog) != 1 {
		t.Fatalf("question log entries = %d, want 1", len(result.QuestionLog))
	}
	if !result.QuestionLog[0].Answer.SkipFlag {
		t.Fatal("expected skip flag to be preserved")
	}
	if !result.QuestionLog[0].Answer.Broadcast {
		t.Fatal("expected broadcast flag to be preserved")
	}
	select {
	case submitErr := <-answerErrCh:
		if submitErr != nil {
			t.Fatalf("submit answer: %v", submitErr)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for answer submission")
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

type captureBus struct {
	mu     sync.Mutex
	events []events.Event
}

func (b *captureBus) Subscribe(_ string, _ events.Handler) {}

func (b *captureBus) SubscribeAll(_ events.Handler) {}

func (b *captureBus) Publish(event events.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
}

func (b *captureBus) snapshot() []events.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]events.Event, len(b.events))
	copy(out, b.events)
	return out
}
