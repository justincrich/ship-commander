package state

import "testing"

func TestValidateClassificationProof(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mission ClassifiedMission
		token   DemoTokenProof
		wantErr bool
	}{
		{
			name: "RED_ALERT passes with tests and commands",
			mission: ClassifiedMission{
				ID:             "MISSION-1",
				Classification: ClassificationREDAlert,
			},
			token: DemoTokenProof{
				Tests:    []string{"go test ./..."},
				Commands: []string{"go test ./..."},
			},
		},
		{
			name: "RED_ALERT fails without tests",
			mission: ClassifiedMission{
				ID:             "MISSION-2",
				Classification: ClassificationREDAlert,
			},
			token: DemoTokenProof{
				Commands: []string{"go test ./..."},
			},
			wantErr: true,
		},
		{
			name: "RED_ALERT fails without commands or diff refs",
			mission: ClassifiedMission{
				ID:             "MISSION-3",
				Classification: ClassificationREDAlert,
			},
			token: DemoTokenProof{
				Tests: []string{"TestEvidence"},
			},
			wantErr: true,
		},
		{
			name: "STANDARD_OPS passes with manual steps",
			mission: ClassifiedMission{
				ID:             "MISSION-4",
				Classification: ClassificationStandardOps,
			},
			token: DemoTokenProof{
				ManualSteps: []string{"Run the command and inspect output"},
			},
		},
		{
			name: "STANDARD_OPS fails without evidence",
			mission: ClassifiedMission{
				ID:             "MISSION-5",
				Classification: ClassificationStandardOps,
			},
			wantErr: true,
		},
		{
			name: "unsupported classification fails",
			mission: ClassifiedMission{
				ID:             "MISSION-6",
				Classification: "UNKNOWN",
			},
			token: DemoTokenProof{
				Commands: []string{"echo ok"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateClassificationProof(tt.mission, tt.token)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
