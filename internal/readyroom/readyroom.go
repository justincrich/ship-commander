package readyroom

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ship-commander/sc3/internal/admiral"
	"github.com/ship-commander/sc3/internal/commission"
	"github.com/ship-commander/sc3/internal/events"
)

const (
	// DefaultMaxIterations bounds the planning loop when no override is provided.
	DefaultMaxIterations = 5
)

// AgentRole identifies one planning specialist in the Ready Room.
type AgentRole string

const (
	// RoleCaptain is responsible for functional decomposition.
	RoleCaptain AgentRole = "captain"
	// RoleCommander is responsible for technical decomposition.
	RoleCommander AgentRole = "commander"
	// RoleDesignOfficer is responsible for design decomposition.
	RoleDesignOfficer AgentRole = "designOfficer"
)

var requiredRoles = []AgentRole{RoleCaptain, RoleCommander, RoleDesignOfficer}

// CoverageState describes use-case coverage status in planning output.
type CoverageState string

const (
	// CoverageCovered indicates at least one fully signed mission references a use case.
	CoverageCovered CoverageState = "covered"
	// CoveragePartial indicates a use case is referenced but only by partially signed missions.
	CoveragePartial CoverageState = "partial"
	// CoverageUncovered indicates a use case is not referenced by any mission.
	CoverageUncovered CoverageState = "uncovered"
)

// ReadyRoomMessage is the structured routing envelope between planning sessions.
//
//nolint:revive // Required by issue contract and upstream planning schema.
type ReadyRoomMessage struct {
	From      string
	To        string
	Type      string
	Domain    string
	Content   string
	Timestamp time.Time
}

// MissionSignoffs tracks deterministic three-way mission approval state.
type MissionSignoffs struct {
	Captain       bool
	Commander     bool
	DesignOfficer bool
}

// MissionPlan is one planned mission candidate produced by Ready Room sessions.
type MissionPlan struct {
	ID         string
	UseCaseIDs []string
	Signoffs   MissionSignoffs
}

// MissionContribution captures a single session's mission-level output for one iteration.
type MissionContribution struct {
	MissionID  string
	UseCaseIDs []string
	SignOff    bool
}

// SessionInput is the isolated context each session receives on each loop iteration.
type SessionInput struct {
	Iteration  int
	Commission commission.Commission
	Inbox      []ReadyRoomMessage
}

// SessionOutput is what one session returns for a single planning iteration.
type SessionOutput struct {
	Messages  []ReadyRoomMessage
	Missions  []MissionContribution
	Questions []admiral.AdmiralQuestion
}

// SpawnRequest contains context required to initialize one session.
type SpawnRequest struct {
	Role       AgentRole
	Commission commission.Commission
}

// SessionFactory spawns isolated planning sessions.
type SessionFactory interface {
	Spawn(ctx context.Context, request SpawnRequest) (Session, error)
}

// Session defines the minimal deterministic contract for a planning specialist session.
type Session interface {
	ID() string
	Execute(ctx context.Context, input SessionInput) (SessionOutput, error)
	Close(ctx context.Context) error
}

// PlanResult is the deterministic Ready Room output snapshot.
type PlanResult struct {
	Missions    []MissionPlan
	Coverage    map[string]CoverageState
	Messages    []ReadyRoomMessage
	QuestionLog []admiral.QuestionRecord
	Iterations  int
	Consensus   bool
}

// ReadyRoom coordinates planning across captain, commander, and design officer sessions.
type ReadyRoom struct {
	factory       SessionFactory
	commission    commission.Commission
	maxIterations int
	now           func() time.Time

	sessions     map[AgentRole]Session
	mailboxes    map[AgentRole][]ReadyRoomMessage
	messages     []ReadyRoomMessage
	missionPlan  map[string]*MissionPlan
	eventBus     events.Bus
	questionGate *admiral.QuestionGate
}

// New builds a ReadyRoom planning coordinator.
func New(factory SessionFactory, comm commission.Commission, maxIterations int) (*ReadyRoom, error) {
	if factory == nil {
		return nil, errors.New("session factory is required")
	}
	if strings.TrimSpace(comm.ID) == "" {
		return nil, errors.New("commission id is required")
	}
	if maxIterations <= 0 {
		maxIterations = DefaultMaxIterations
	}

	return &ReadyRoom{
		factory:       factory,
		commission:    comm,
		maxIterations: maxIterations,
		now:           time.Now,
		sessions:      make(map[AgentRole]Session, len(requiredRoles)),
		mailboxes:     make(map[AgentRole][]ReadyRoomMessage, len(requiredRoles)),
		messages:      make([]ReadyRoomMessage, 0),
		missionPlan:   make(map[string]*MissionPlan),
		eventBus:      events.New(),
		questionGate:  admiral.NewQuestionGate(1),
	}, nil
}

// QuestionGate returns the blocking Admiral question gate used by the planning loop.
func (r *ReadyRoom) QuestionGate() *admiral.QuestionGate {
	if r == nil {
		return nil
	}
	return r.questionGate
}

// SetEventBus overrides the default event bus.
func (r *ReadyRoom) SetEventBus(bus events.Bus) error {
	if r == nil {
		return errors.New("ready room is nil")
	}
	if bus == nil {
		return errors.New("event bus is required")
	}
	r.eventBus = bus
	return nil
}

// Plan executes the deterministic planning loop until consensus or max iterations.
func (r *ReadyRoom) Plan(ctx context.Context) (result PlanResult, err error) {
	if r == nil {
		return PlanResult{}, errors.New("ready room is nil")
	}

	if err := r.spawnSessions(ctx); err != nil {
		return PlanResult{}, err
	}
	defer func() {
		closeErr := r.closeSessions(context.Background())
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		err = errors.Join(err, closeErr)
	}()

	for iteration := 1; iteration <= r.maxIterations; iteration++ {
		for _, role := range requiredRoles {
			session, ok := r.sessions[role]
			if !ok {
				return PlanResult{}, fmt.Errorf("session for role %q not found", role)
			}

			input := SessionInput{
				Iteration:  iteration,
				Commission: r.commission,
				Inbox:      append([]ReadyRoomMessage(nil), r.mailboxes[role]...),
			}
			r.mailboxes[role] = nil

			output, err := session.Execute(ctx, input)
			if err != nil {
				return PlanResult{}, fmt.Errorf("execute session role=%s id=%s: %w", role, session.ID(), err)
			}

			if err := r.handleQuestions(ctx, role, output.Questions); err != nil {
				return PlanResult{}, err
			}
			r.mergeMissionContributions(role, output.Missions)
			if err := r.routeMessages(role, output.Messages); err != nil {
				return PlanResult{}, err
			}
		}

		consensus, coverage := r.ValidateConsensus()
		if consensus {
			return r.buildResult(iteration, coverage, true), nil
		}
	}

	_, coverage := r.ValidateConsensus()
	return r.buildResult(r.maxIterations, coverage, false), nil
}

// ValidateConsensus deterministically checks signoff and use-case coverage completion.
func (r *ReadyRoom) ValidateConsensus() (bool, map[string]CoverageState) {
	if r == nil {
		return false, nil
	}

	for _, mission := range r.missionPlan {
		if !mission.Signoffs.Captain || !mission.Signoffs.Commander || !mission.Signoffs.DesignOfficer {
			return false, r.BuildUseCaseCoverage()
		}
	}

	coverage := r.BuildUseCaseCoverage()
	for _, status := range coverage {
		if status != CoverageCovered {
			return false, coverage
		}
	}

	return true, coverage
}

// BuildUseCaseCoverage computes covered/partial/uncovered across commission use cases.
func (r *ReadyRoom) BuildUseCaseCoverage() map[string]CoverageState {
	coverage := make(map[string]CoverageState, len(r.commission.UseCases))
	for _, useCase := range r.commission.UseCases {
		coverage[useCase.ID] = CoverageUncovered
	}

	for _, mission := range r.missionPlan {
		for _, useCaseID := range mission.UseCaseIDs {
			if _, ok := coverage[useCaseID]; !ok {
				continue
			}
			if mission.Signoffs.Captain && mission.Signoffs.Commander && mission.Signoffs.DesignOfficer {
				coverage[useCaseID] = CoverageCovered
				continue
			}
			if coverage[useCaseID] != CoverageCovered {
				coverage[useCaseID] = CoveragePartial
			}
		}
	}

	return coverage
}

func (r *ReadyRoom) spawnSessions(ctx context.Context) error {
	for _, role := range requiredRoles {
		if _, exists := r.sessions[role]; exists {
			continue
		}

		session, err := r.factory.Spawn(ctx, SpawnRequest{
			Role:       role,
			Commission: r.commission,
		})
		if err != nil {
			return fmt.Errorf("spawn session %s: %w", role, err)
		}
		r.sessions[role] = session
	}

	return nil
}

func (r *ReadyRoom) closeSessions(ctx context.Context) error {
	for role, session := range r.sessions {
		if err := session.Close(ctx); err != nil {
			return fmt.Errorf("close session %s: %w", role, err)
		}
	}
	return nil
}

func (r *ReadyRoom) mergeMissionContributions(role AgentRole, contributions []MissionContribution) {
	for _, contribution := range contributions {
		missionID := strings.TrimSpace(contribution.MissionID)
		if missionID == "" {
			continue
		}

		mission, ok := r.missionPlan[missionID]
		if !ok {
			mission = &MissionPlan{
				ID:         missionID,
				UseCaseIDs: make([]string, 0),
			}
			r.missionPlan[missionID] = mission
		}

		for _, useCaseID := range contribution.UseCaseIDs {
			useCaseID = strings.TrimSpace(useCaseID)
			if useCaseID == "" || slices.Contains(mission.UseCaseIDs, useCaseID) {
				continue
			}
			mission.UseCaseIDs = append(mission.UseCaseIDs, useCaseID)
		}

		if !contribution.SignOff {
			continue
		}

		switch role {
		case RoleCaptain:
			mission.Signoffs.Captain = true
		case RoleCommander:
			mission.Signoffs.Commander = true
		case RoleDesignOfficer:
			mission.Signoffs.DesignOfficer = true
		}
	}
}

func (r *ReadyRoom) routeMessages(from AgentRole, messages []ReadyRoomMessage) error {
	for _, message := range messages {
		normalized := ReadyRoomMessage{
			From:      string(from),
			To:        strings.TrimSpace(message.To),
			Type:      strings.TrimSpace(message.Type),
			Domain:    strings.TrimSpace(message.Domain),
			Content:   message.Content,
			Timestamp: message.Timestamp,
		}
		if normalized.Timestamp.IsZero() {
			normalized.Timestamp = r.now().UTC()
		}
		if normalized.To == "" {
			return fmt.Errorf("route message from=%s: recipient is required", from)
		}

		r.messages = append(r.messages, normalized)

		switch normalized.To {
		case "all", "broadcast":
			for _, role := range requiredRoles {
				if role == from {
					continue
				}
				r.mailboxes[role] = append(r.mailboxes[role], normalized)
			}
		default:
			role := AgentRole(normalized.To)
			if !slices.Contains(requiredRoles, role) {
				return fmt.Errorf("route message from=%s: unknown recipient %q", from, normalized.To)
			}
			r.mailboxes[role] = append(r.mailboxes[role], normalized)
		}
	}

	return nil
}

func (r *ReadyRoom) handleQuestions(
	ctx context.Context,
	role AgentRole,
	questions []admiral.AdmiralQuestion,
) error {
	if len(questions) == 0 {
		return nil
	}
	if r.questionGate == nil {
		return errors.New("question gate is not configured")
	}

	for _, question := range questions {
		question.AskingAgent = string(role)

		if r.eventBus != nil {
			r.eventBus.Publish(events.Event{
				Type:       events.EventTypeAdmiralQuestion,
				EntityType: "planning_question",
				EntityID:   strings.TrimSpace(question.QuestionID),
				Payload:    question,
				Severity:   events.SeverityInfo,
			})
		}

		answer, err := r.questionGate.Ask(ctx, question)
		if err != nil {
			return fmt.Errorf("question gate ask role=%s question_id=%s: %w", role, question.QuestionID, err)
		}
		if err := admiral.ValidateAnswer(question, answer); err != nil {
			return fmt.Errorf("invalid admiral answer role=%s question_id=%s: %w", role, question.QuestionID, err)
		}
		r.routeAdmiralAnswer(role, question, answer)
	}

	return nil
}

func (r *ReadyRoom) routeAdmiralAnswer(
	askingRole AgentRole,
	question admiral.AdmiralQuestion,
	answer admiral.AdmiralAnswer,
) {
	if r == nil {
		return
	}

	message := ReadyRoomMessage{
		From:      "admiral",
		To:        string(askingRole),
		Type:      "admiral_answer",
		Domain:    strings.TrimSpace(question.Domain),
		Content:   formatAdmiralAnswer(answer),
		Timestamp: r.now().UTC(),
	}
	r.messages = append(r.messages, message)
	r.mailboxes[askingRole] = append(r.mailboxes[askingRole], message)

	if !answer.Broadcast {
		return
	}

	broadcastMessage := message
	broadcastMessage.To = "broadcast"
	for _, role := range requiredRoles {
		if role == askingRole {
			continue
		}
		r.mailboxes[role] = append(r.mailboxes[role], broadcastMessage)
	}
	r.messages = append(r.messages, broadcastMessage)
}

func formatAdmiralAnswer(answer admiral.AdmiralAnswer) string {
	parts := []string{fmt.Sprintf("question_id=%s", strings.TrimSpace(answer.QuestionID))}
	if option := strings.TrimSpace(answer.SelectedOption); option != "" {
		parts = append(parts, fmt.Sprintf("selected_option=%s", option))
	}
	if freeText := strings.TrimSpace(answer.FreeText); freeText != "" {
		parts = append(parts, fmt.Sprintf("free_text=%s", freeText))
	}
	if answer.SkipFlag {
		parts = append(parts, "skip=true")
	}
	if answer.Broadcast {
		parts = append(parts, "broadcast=true")
	}
	return strings.Join(parts, "\n")
}

func (r *ReadyRoom) buildResult(iterations int, coverage map[string]CoverageState, consensus bool) PlanResult {
	missions := make([]MissionPlan, 0, len(r.missionPlan))
	for _, mission := range r.missionPlan {
		missions = append(missions, MissionPlan{
			ID:         mission.ID,
			UseCaseIDs: append([]string(nil), mission.UseCaseIDs...),
			Signoffs:   mission.Signoffs,
		})
	}
	slices.SortFunc(missions, func(a, b MissionPlan) int {
		return strings.Compare(a.ID, b.ID)
	})

	messages := make([]ReadyRoomMessage, len(r.messages))
	copy(messages, r.messages)
	questionLog := make([]admiral.QuestionRecord, 0)
	if r.questionGate != nil {
		questionLog = r.questionGate.History()
	}

	return PlanResult{
		Missions:    missions,
		Coverage:    coverage,
		Messages:    messages,
		QuestionLog: questionLog,
		Iterations:  iterations,
		Consensus:   consensus,
	}
}
