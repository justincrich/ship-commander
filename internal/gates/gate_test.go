package gates

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunVerifyREDClassifications(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		command            string
		wantClassification string
		wantExitCode       int
	}{
		{
			name:               "exit zero is reject vanity",
			command:            "exit 0",
			wantClassification: ClassificationRejectVanity,
			wantExitCode:       0,
		},
		{
			name:               "syntax output is reject syntax",
			command:            "echo 'syntax error: unexpected EOF'; exit 1",
			wantClassification: ClassificationRejectSyntax,
			wantExitCode:       1,
		},
		{
			name:               "failing test output is accept",
			command:            "echo '--- FAIL: TestRED (0.00s)'; exit 1",
			wantClassification: ClassificationAccept,
			wantExitCode:       1,
		},
		{
			name:               "unknown non zero failure is reject failure",
			command:            "echo 'remote dependency unavailable'; exit 1",
			wantClassification: ClassificationRejectFailure,
			wantExitCode:       1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			workdir := t.TempDir()
			evidence := &fakeEvidenceStore{}
			runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
				Timeout: 500 * time.Millisecond,
				ProjectCommands: map[string][]string{
					GateTypeVerifyRED: {tt.command},
				},
			})
			if err != nil {
				t.Fatalf("new shell runner: %v", err)
			}

			result, err := runner.Run(context.Background(), GateTypeVerifyRED, workdir, "mission-red")
			if err != nil {
				t.Fatalf("run gate: %v", err)
			}

			if result.Classification != tt.wantClassification {
				t.Fatalf("classification = %q, want %q", result.Classification, tt.wantClassification)
			}
			if result.ExitCode != tt.wantExitCode {
				t.Fatalf("exit code = %d, want %d", result.ExitCode, tt.wantExitCode)
			}
			if result.Attempt != 1 {
				t.Fatalf("attempt = %d, want 1", result.Attempt)
			}
			if len(evidence.records) != 1 {
				t.Fatalf("evidence records = %d, want 1", len(evidence.records))
			}
			if evidence.records[0].Classification != tt.wantClassification {
				t.Fatalf("persisted classification = %q, want %q", evidence.records[0].Classification, tt.wantClassification)
			}
		})
	}
}

func TestRunVerifyGREENRejectsWithFailureMessage(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		ProjectCommands: map[string][]string{
			GateTypeVerifyGREEN: {"echo '--- FAIL: TestCriticalPath'; echo 'stack follows'; exit 1"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	result, err := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-green")
	if err != nil {
		t.Fatalf("run gate: %v", err)
	}

	if result.Classification != ClassificationRejectFailure {
		t.Fatalf("classification = %q, want %q", result.Classification, ClassificationRejectFailure)
	}
	if !strings.Contains(result.OutputSnippet, "--- FAIL: TestCriticalPath") {
		t.Fatalf("output snippet = %q, want first failure line", result.OutputSnippet)
	}
}

func TestRunVerifyGREENRunsInfraThreeTimesAndRejectsFlaky(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	infraCounterFile := filepath.Join(workdir, "infra-count.txt")
	infraCmd := "count_file='" + infraCounterFile + "'; n=$(cat \"$count_file\" 2>/dev/null || echo 0); n=$((n+1)); echo \"$n\" > \"$count_file\"; if [ \"$n\" -eq 2 ]; then echo 'infra failed on run 2'; exit 1; fi; exit 0"

	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		ProjectCommands: map[string][]string{
			GateTypeVerifyGREEN: {"exit 0"},
		},
		GreenInfraCommands: []string{infraCmd},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	result, err := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-green-infra")
	if err != nil {
		t.Fatalf("run gate: %v", err)
	}
	if result.Classification != ClassificationRejectFailure {
		t.Fatalf("classification = %q, want %q", result.Classification, ClassificationRejectFailure)
	}

	// #nosec G304 -- infraCounterFile is created from t.TempDir() within this test.
	countRaw, err := os.ReadFile(infraCounterFile)
	if err != nil {
		t.Fatalf("read infra count file: %v", err)
	}
	if strings.TrimSpace(string(countRaw)) != "3" {
		t.Fatalf("infra run count = %q, want 3", strings.TrimSpace(string(countRaw)))
	}
}

func TestRunVerifyIMPLEMENTSequentialAndAutoPass(t *testing.T) {
	t.Parallel()

	t.Run("stops after first failing command", func(t *testing.T) {
		t.Parallel()

		workdir := t.TempDir()
		logFile := filepath.Join(workdir, "verify-implement.log")
		evidence := &fakeEvidenceStore{}
		runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
			ProjectCommands: map[string][]string{
				GateTypeVerifyIMPLEMENT: {
					"echo typecheck >> '" + logFile + "'",
					"echo lint >> '" + logFile + "'; exit 1",
					"echo build >> '" + logFile + "'",
				},
			},
		})
		if err != nil {
			t.Fatalf("new shell runner: %v", err)
		}

		result, runErr := runner.Run(context.Background(), GateTypeVerifyIMPLEMENT, workdir, "mission-implement")
		if runErr != nil {
			t.Fatalf("run gate: %v", runErr)
		}
		if result.Classification != ClassificationRejectFailure {
			t.Fatalf("classification = %q, want %q", result.Classification, ClassificationRejectFailure)
		}

		// #nosec G304 -- logFile is created from t.TempDir() within this test.
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("read log file: %v", err)
		}
		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		if strings.Join(lines, ",") != "typecheck,lint" {
			t.Fatalf("executed steps = %v, want [typecheck lint]", lines)
		}
	})

	t.Run("no commands configured auto passes", func(t *testing.T) {
		t.Parallel()

		workdir := t.TempDir()
		evidence := &fakeEvidenceStore{}
		runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
			ProjectCommands: map[string][]string{
				GateTypeVerifyRED: {"exit 0"},
			},
		})
		if err != nil {
			t.Fatalf("new shell runner: %v", err)
		}

		result, runErr := runner.Run(context.Background(), GateTypeVerifyIMPLEMENT, workdir, "mission-auto-pass")
		if runErr != nil {
			t.Fatalf("run gate: %v", runErr)
		}
		if result.Classification != ClassificationAccept {
			t.Fatalf("classification = %q, want %q", result.Classification, ClassificationAccept)
		}
		if result.ExitCode != 0 {
			t.Fatalf("exit code = %d, want 0", result.ExitCode)
		}
	})
}

func TestRunEnforcesTimeout(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		Timeout: 50 * time.Millisecond,
		ProjectCommands: map[string][]string{
			GateTypeVerifyREFACTOR: {"sleep 1"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	result, runErr := runner.Run(context.Background(), GateTypeVerifyREFACTOR, workdir, "mission-timeout")
	if runErr != nil {
		t.Fatalf("run gate: %v", runErr)
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1 for timeout", result.ExitCode)
	}
	if result.Classification != ClassificationRejectFailure {
		t.Fatalf("classification = %q, want %q", result.Classification, ClassificationRejectFailure)
	}
	if !strings.Contains(result.Output, "timed out") {
		t.Fatalf("output = %q, expected timeout text", result.Output)
	}
}

func TestRunCapturesOutputWithConfiguredLimit(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		OutputLimitBytes: 128,
		ProjectCommands: map[string][]string{
			GateTypeVerifyREFACTOR: {"head -c 4096 /dev/zero | tr '\\000' 'a'; exit 1"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	result, runErr := runner.Run(context.Background(), GateTypeVerifyREFACTOR, workdir, "mission-output-limit")
	if runErr != nil {
		t.Fatalf("run gate: %v", runErr)
	}
	if len(result.Output) > 128 {
		t.Fatalf("output length = %d, want <= 128", len(result.Output))
	}
	if !strings.Contains(result.Output, "truncated") {
		t.Fatalf("output = %q, expected truncation marker", result.Output)
	}
}

func TestRunPersistsEvidenceAndAttemptsIncrement(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		ProjectCommands: map[string][]string{
			GateTypeVerifyGREEN: {"exit 0"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	first, err := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-attempt")
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	second, err := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-attempt")
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if first.Attempt != 1 || second.Attempt != 2 {
		t.Fatalf("attempts = (%d, %d), want (1, 2)", first.Attempt, second.Attempt)
	}
	if len(evidence.records) != 2 {
		t.Fatalf("evidence record count = %d, want 2", len(evidence.records))
	}
	if evidence.records[0].Type != GateTypeVerifyGREEN {
		t.Fatalf("persisted type = %q, want %q", evidence.records[0].Type, GateTypeVerifyGREEN)
	}
	if evidence.records[0].Timestamp.IsZero() {
		t.Fatal("persisted timestamp is zero")
	}
}

func TestRunUsesMissionCommandsAndVariableSubstitution(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	outFile := filepath.Join(workdir, "substitution.txt")
	evidence := &fakeEvidenceStore{}

	missionCommands := missionResolverFunc(func(_ context.Context, missionID, gateType string) ([]string, error) {
		if missionID != "mission-substitution" || gateType != GateTypeVerifyIMPLEMENT {
			return nil, nil
		}
		return []string{"echo '{mission_id}|{worktree_dir}|{test_file}' > '" + outFile + "'"}, nil
	})
	variables := variableResolverFunc(func(_ context.Context, _ string) (map[string]string, error) {
		return map[string]string{
			"test_file": "test/ac_red_test.go",
		}, nil
	})

	runner, err := NewShellRunner(evidence, missionCommands, variables, RunnerConfig{
		ProjectCommands: map[string][]string{
			GateTypeVerifyIMPLEMENT: {"echo project-command-ran; exit 1"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	result, runErr := runner.Run(context.Background(), GateTypeVerifyIMPLEMENT, workdir, "mission-substitution")
	if runErr != nil {
		t.Fatalf("run gate: %v", runErr)
	}
	if result.Classification != ClassificationAccept {
		t.Fatalf("classification = %q, want %q", result.Classification, ClassificationAccept)
	}

	// #nosec G304 -- outFile is created from t.TempDir() within this test.
	raw, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read substitution file: %v", err)
	}
	got := strings.TrimSpace(string(raw))
	want := "mission-substitution|" + workdir + "|test/ac_red_test.go"
	if got != want {
		t.Fatalf("substitution output = %q, want %q", got, want)
	}
}

func TestRunReturnsErrorWhenNonImplementGateHasNoCommands(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		ProjectCommands: map[string][]string{},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	_, runErr := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-missing-commands")
	if runErr == nil {
		t.Fatal("expected error for missing commands")
	}
	if !strings.Contains(runErr.Error(), "no commands configured") {
		t.Fatalf("error = %v, want missing commands message", runErr)
	}
}

func TestNewRunnerRequiresDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewRunner(nil, &fakeEvidenceStore{}, nil, nil, RunnerConfig{})
	if err == nil {
		t.Fatal("expected executor required error")
	}

	_, err = NewRunner(shellExecutor{}, nil, nil, nil, RunnerConfig{})
	if err == nil {
		t.Fatal("expected evidence store required error")
	}
}

type fakeEvidenceStore struct {
	records []GateResult
	err     error
}

func (f *fakeEvidenceStore) RecordGateEvidence(_ context.Context, _ string, result GateResult) error {
	if f.err != nil {
		return f.err
	}
	f.records = append(f.records, result)
	return nil
}

type missionResolverFunc func(ctx context.Context, missionID, gateType string) ([]string, error)

func (f missionResolverFunc) ResolveGateCommands(
	ctx context.Context,
	missionID string,
	gateType string,
) ([]string, error) {
	return f(ctx, missionID, gateType)
}

type variableResolverFunc func(ctx context.Context, missionID string) (map[string]string, error)

func (f variableResolverFunc) ResolveGateVariables(ctx context.Context, missionID string) (map[string]string, error) {
	return f(ctx, missionID)
}

func TestRunReturnsErrorWhenEvidencePersistFails(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	evidence := &fakeEvidenceStore{err: errors.New("beads unavailable")}
	runner, err := NewShellRunner(evidence, nil, nil, RunnerConfig{
		ProjectCommands: map[string][]string{
			GateTypeVerifyGREEN: {"exit 0"},
		},
	})
	if err != nil {
		t.Fatalf("new shell runner: %v", err)
	}

	_, runErr := runner.Run(context.Background(), GateTypeVerifyGREEN, workdir, "mission-evidence-fail")
	if runErr == nil {
		t.Fatal("expected evidence persistence error")
	}
	if !strings.Contains(runErr.Error(), "record gate evidence") {
		t.Fatalf("error = %v, want evidence persistence context", runErr)
	}
}
