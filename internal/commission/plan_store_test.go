package commission

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type runnerResponse struct {
	output []byte
	err    error
}

type runnerCall struct {
	name string
	args []string
}

type scriptedPlanRunner struct {
	responses []runnerResponse
	calls     []runnerCall
	index     int
}

func (r *scriptedPlanRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, runnerCall{
		name: name,
		args: append([]string(nil), args...),
	})
	if r.index >= len(r.responses) {
		return nil, errors.New("unexpected runner call")
	}
	resp := r.responses[r.index]
	r.index++
	return resp.output, resp.err
}

func TestSavePlanWithRunnerPersistsFullStateInBeadsNotes(t *testing.T) {
	t.Parallel()

	state := samplePlanState()
	runner := &scriptedPlanRunner{
		responses: []runnerResponse{{output: []byte(`{"ok":true}`)}},
	}

	err := SavePlanWithRunner(context.Background(), "ship-commander-3-comm-1", state, runner)
	if err != nil {
		t.Fatalf("save plan: %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("runner calls = %d, want 1", len(runner.calls))
	}
	call := runner.calls[0]
	if call.name != "bd" {
		t.Fatalf("command = %q, want bd", call.name)
	}
	if len(call.args) < 4 || call.args[0] != "update" || call.args[1] != "ship-commander-3-comm-1" || call.args[2] != "--notes" {
		t.Fatalf("unexpected args: %v", call.args)
	}

	var payload persistedPlanEnvelope
	if err := json.Unmarshal([]byte(call.args[3]), &payload); err != nil {
		t.Fatalf("parse persisted payload: %v", err)
	}
	if payload.CommissionStatus != PlanningStatusApproved {
		t.Fatalf("status = %q, want %q", payload.CommissionStatus, PlanningStatusApproved)
	}
	if payload.State.IterationCount != state.IterationCount {
		t.Fatalf("iteration count = %d, want %d", payload.State.IterationCount, state.IterationCount)
	}
	if len(payload.State.MissionList) != len(state.MissionList) {
		t.Fatalf("mission list len = %d, want %d", len(payload.State.MissionList), len(state.MissionList))
	}
	if len(payload.State.ReadyRoomMessages) != len(state.ReadyRoomMessages) {
		t.Fatalf("message len = %d, want %d", len(payload.State.ReadyRoomMessages), len(state.ReadyRoomMessages))
	}
	if len(payload.State.SignoffMap) != len(state.SignoffMap) {
		t.Fatalf("signoff map len = %d, want %d", len(payload.State.SignoffMap), len(state.SignoffMap))
	}
	if len(payload.State.CoverageMap) != len(state.CoverageMap) {
		t.Fatalf("coverage map len = %d, want %d", len(payload.State.CoverageMap), len(state.CoverageMap))
	}
	if len(payload.State.WaveAssignments) != len(state.WaveAssignments) {
		t.Fatalf("wave assignment len = %d, want %d", len(payload.State.WaveAssignments), len(state.WaveAssignments))
	}
}

func TestLoadPlanWithRunnerReadsPersistedState(t *testing.T) {
	t.Parallel()

	state := samplePlanState()
	envelope := persistedPlanEnvelope{
		Version:          planStorageVersion,
		CommissionID:     "ship-commander-3-comm-1",
		CommissionStatus: PlanningStatusApproved,
		State:            state,
	}
	rawEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	showOutput := []byte(`[{"id":"ship-commander-3-comm-1","notes":` + strconvQuote(string(rawEnvelope)) + `}]`)

	runner := &scriptedPlanRunner{
		responses: []runnerResponse{{output: showOutput}},
	}
	loaded, err := LoadPlanWithRunner(context.Background(), "ship-commander-3-comm-1", runner)
	if err != nil {
		t.Fatalf("load plan: %v", err)
	}

	if loaded.IterationCount != state.IterationCount {
		t.Fatalf("iteration count = %d, want %d", loaded.IterationCount, state.IterationCount)
	}
	if loaded.MissionList[0].ID != "M-1" {
		t.Fatalf("mission id = %q, want M-1", loaded.MissionList[0].ID)
	}
	if loaded.SignoffMap["M-1"].Captain != state.SignoffMap["M-1"].Captain {
		t.Fatalf("captain signoff = %v, want %v", loaded.SignoffMap["M-1"].Captain, state.SignoffMap["M-1"].Captain)
	}
}

func TestShelvePlanWithRunnerMarksCommissionShelvedAndPreservesState(t *testing.T) {
	t.Parallel()

	state := samplePlanState()
	envelope := persistedPlanEnvelope{
		Version:          planStorageVersion,
		CommissionID:     "ship-commander-3-comm-1",
		CommissionStatus: PlanningStatusApproved,
		State:            state,
	}
	rawEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	showOutput := []byte(`[{"id":"ship-commander-3-comm-1","notes":` + strconvQuote(string(rawEnvelope)) + `}]`)

	runner := &scriptedPlanRunner{
		responses: []runnerResponse{
			{output: showOutput},
			{output: []byte(`{"ok":true}`)},
		},
	}

	err = ShelvePlanWithRunner(context.Background(), "ship-commander-3-comm-1", "pause for review", runner)
	if err != nil {
		t.Fatalf("shelve plan: %v", err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("runner calls = %d, want 2", len(runner.calls))
	}
	updateCall := runner.calls[1]
	if len(updateCall.args) < 4 || updateCall.args[0] != "update" || updateCall.args[2] != "--notes" {
		t.Fatalf("unexpected update args: %v", updateCall.args)
	}

	var payload persistedPlanEnvelope
	if err := json.Unmarshal([]byte(updateCall.args[3]), &payload); err != nil {
		t.Fatalf("parse shelved payload: %v", err)
	}
	if payload.CommissionStatus != PlanningStatusShelved {
		t.Fatalf("status = %q, want %q", payload.CommissionStatus, PlanningStatusShelved)
	}
	if payload.State.IterationCount != state.IterationCount {
		t.Fatalf("iteration count = %d, want %d", payload.State.IterationCount, state.IterationCount)
	}
	if payload.FeedbackText != "pause for review" {
		t.Fatalf("feedback text = %q, want %q", payload.FeedbackText, "pause for review")
	}
}

func TestResumePlanWithRunnerReturnsRespawnRoles(t *testing.T) {
	t.Parallel()

	state := samplePlanState()
	envelope := persistedPlanEnvelope{
		Version:          planStorageVersion,
		CommissionID:     "ship-commander-3-comm-1",
		CommissionStatus: PlanningStatusShelved,
		State:            state,
	}
	rawEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	showOutput := []byte(`[{"id":"ship-commander-3-comm-1","notes":` + strconvQuote(string(rawEnvelope)) + `}]`)

	runner := &scriptedPlanRunner{
		responses: []runnerResponse{{output: showOutput}},
	}
	result, err := ResumePlanWithRunner(context.Background(), "ship-commander-3-comm-1", runner)
	if err != nil {
		t.Fatalf("resume plan: %v", err)
	}
	if len(result.RespawnRoles) != 3 {
		t.Fatalf("respawn role count = %d, want 3", len(result.RespawnRoles))
	}
	if strings.Join(result.RespawnRoles, ",") != "captain,commander,designOfficer" {
		t.Fatalf("respawn roles = %v", result.RespawnRoles)
	}
	if result.State.IterationCount != state.IterationCount {
		t.Fatalf("iteration count = %d, want %d", result.State.IterationCount, state.IterationCount)
	}
}

func TestReexecutePlanWithRunnerLoadsApprovedPlan(t *testing.T) {
	t.Parallel()

	state := samplePlanState()
	envelope := persistedPlanEnvelope{
		Version:          planStorageVersion,
		CommissionID:     "ship-commander-3-comm-1",
		CommissionStatus: PlanningStatusApproved,
		State:            state,
	}
	rawEnvelope, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}
	showOutput := []byte(`[{"id":"ship-commander-3-comm-1","notes":` + strconvQuote(string(rawEnvelope)) + `}]`)

	runner := &scriptedPlanRunner{
		responses: []runnerResponse{{output: showOutput}},
	}
	loaded, err := ReexecutePlanWithRunner(context.Background(), "ship-commander-3-comm-1", runner)
	if err != nil {
		t.Fatalf("reexecute plan: %v", err)
	}
	if loaded.IterationCount != state.IterationCount {
		t.Fatalf("iteration count = %d, want %d", loaded.IterationCount, state.IterationCount)
	}
	if len(loaded.WaveAssignments) != len(state.WaveAssignments) {
		t.Fatalf("wave assignment count = %d, want %d", len(loaded.WaveAssignments), len(state.WaveAssignments))
	}
}

func samplePlanState() PlanState {
	return PlanState{
		MissionList: []PlanMission{
			{ID: "M-1", Title: "Mission One", UseCaseIDs: []string{"UC-1"}},
		},
		ReadyRoomMessages: []PlanMessage{
			{
				From:    "captain",
				To:      "commander",
				Type:    "proposal",
				Domain:  "functional",
				Content: "split by API boundary",
			},
		},
		SignoffMap: map[string]PlanSignoff{
			"M-1": {
				Captain:       true,
				Commander:     true,
				DesignOfficer: false,
			},
		},
		IterationCount: 2,
		CoverageMap: map[string]string{
			"UC-1": "covered",
		},
		WaveAssignments: []PlanWave{
			{Index: 1, MissionIDs: []string{"M-1"}},
		},
	}
}

func strconvQuote(value string) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}
