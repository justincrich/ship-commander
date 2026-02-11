package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/ship-commander/sc3/internal/config"
)

func TestRootCommandVersionFlag(t *testing.T) {
	t.Parallel()

	Version = "v0.1.0-test"
	cmd := newRootCommand(context.Background(), &config.Config{}, testLogger())

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output != "v0.1.0-test" {
		t.Fatalf("version output = %q, want %q", output, "v0.1.0-test")
	}
}

func TestRootCommandHelpListsExpectedSubcommands(t *testing.T) {
	t.Parallel()

	cmd := newRootCommand(context.Background(), &config.Config{}, testLogger())
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := stdout.String()
	expected := []string{"init", "plan", "execute", "tui", "status"}
	for _, name := range expected {
		if !strings.Contains(output, name) {
			t.Fatalf("help output missing %q: %s", name, output)
		}
	}
}

func testLogger() *log.Logger {
	return log.NewWithOptions(&bytes.Buffer{}, log.Options{})
}
