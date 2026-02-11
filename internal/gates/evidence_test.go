package gates

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestBeadsEvidenceStoreRecordGateEvidence(t *testing.T) {
	t.Parallel()

	states := &fakeStateSetter{}
	store, err := NewBeadsEvidenceStore(states)
	if err != nil {
		t.Fatalf("new evidence store: %v", err)
	}

	result := GateResult{
		Type:           GateTypeVerifyGREEN,
		ExitCode:       1,
		Classification: ClassificationRejectFailure,
		OutputSnippet:  "--- FAIL: TestCriticalPath",
		Duration:       250 * time.Millisecond,
		Attempt:        2,
		Timestamp:      time.Date(2026, 2, 11, 12, 0, 0, 0, time.UTC),
	}

	if err := store.RecordGateEvidence(context.Background(), "mission-123", result); err != nil {
		t.Fatalf("record evidence: %v", err)
	}

	if len(states.calls) != 5 {
		t.Fatalf("set-state calls = %d, want 5", len(states.calls))
	}
	joined := strings.Join(states.calls, "\n")
	for _, want := range []string{
		"mission-123|gates.verify_green.attempt_2.exit_code|1",
		"mission-123|gates.verify_green.attempt_2.classification|reject_failure",
		"mission-123|gates.verify_green.attempt_2.output_snippet|--- FAIL: TestCriticalPath",
		"mission-123|gates.verify_green.attempt_2.duration_ms|250",
		"mission-123|gates.verify_green.attempt_2.timestamp|2026-02-11T12:00:00Z",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing expected state update %q in:\n%s", want, joined)
		}
	}
}

func TestBeadsEvidenceStoreValidationAndErrors(t *testing.T) {
	t.Parallel()

	if _, err := NewBeadsEvidenceStore(nil); err == nil {
		t.Fatal("expected nil state setter error")
	}

	states := &fakeStateSetter{err: errors.New("bd unavailable")}
	store, err := NewBeadsEvidenceStore(states)
	if err != nil {
		t.Fatalf("new evidence store: %v", err)
	}

	runErr := store.RecordGateEvidence(context.Background(), "mission-123", GateResult{
		Type:      GateTypeVerifyRED,
		Attempt:   1,
		Timestamp: time.Now(),
	})
	if runErr == nil {
		t.Fatal("expected set-state error")
	}
	if !strings.Contains(runErr.Error(), "set state") {
		t.Fatalf("error = %v, want set-state context", runErr)
	}
}

type fakeStateSetter struct {
	calls []string
	err   error
}

func (f *fakeStateSetter) SetState(id, key, value string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, id+"|"+key+"|"+value)
	return nil
}
