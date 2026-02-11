package main

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/ship-commander/sc3/internal/config"
)

func TestRootCommandVersionFlag(t *testing.T) {
	originalVersion := Version
	defer func() {
		Version = originalVersion
	}()
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

func TestResolveCommandName(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "subcommand", args: []string{"plan"}, want: "plan"},
		{name: "flags then command", args: []string{"--verbose", "execute"}, want: "execute"},
		{name: "no command defaults to root", args: []string{"--help"}, want: "root"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCommandName(tc.args); got != tc.want {
				t.Fatalf("resolveCommandName(%v) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

func TestRedactArgs(t *testing.T) {
	input := []string{
		"execute",
		"--token",
		"abc123",
		"--password=supersecret",
		"--safe=value",
	}
	want := []string{
		"execute",
		"--token",
		"<redacted>",
		"--password=<redacted>",
		"--safe=value",
	}

	if got := redactArgs(input); !reflect.DeepEqual(got, want) {
		t.Fatalf("redactArgs(%v) = %v, want %v", input, got, want)
	}
}
