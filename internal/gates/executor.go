package gates

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tooltrace "github.com/ship-commander/sc3/internal/tracing"
)

type commandExecutor interface {
	Run(ctx context.Context, workdir string, command string, timeout time.Duration, outputLimitBytes int) (commandResult, error)
}

type commandResult struct {
	ExitCode int
	Output   string
	Duration time.Duration
}

type shellExecutor struct{}

func (shellExecutor) Run(
	ctx context.Context,
	workdir string,
	command string,
	timeout time.Duration,
	outputLimitBytes int,
) (commandResult, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return commandResult{}, errors.New("command must not be empty")
	}
	if strings.TrimSpace(workdir) == "" {
		return commandResult{}, errors.New("workdir must not be empty")
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if outputLimitBytes <= 0 {
		outputLimitBytes = DefaultOutputLimitBytes
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output := newLimitedBuffer(outputLimitBytes)

	start := time.Now()
	exitCode, stdout, stderr, err := tooltrace.ExecuteTool(runCtx, "sh", []string{"-c", command}, workdir)
	duration := time.Since(start)

	output.WriteString(stdout)
	if strings.TrimSpace(stdout) != "" && strings.TrimSpace(stderr) != "" {
		output.WriteString("\n")
	}
	output.WriteString(stderr)

	if err != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			if exitCode == 0 {
				exitCode = -1
			}
			output.WriteString(fmt.Sprintf("\ncommand timed out after %s", timeout))
		} else if exitCode == 0 {
			return commandResult{}, fmt.Errorf("run command %q: %w", command, err)
		}
	}

	return commandResult{
		ExitCode: exitCode,
		Output:   strings.TrimSpace(output.String()),
		Duration: duration,
	}, nil
}

type limitedBuffer struct {
	max       int
	data      []byte
	truncated bool
}

func newLimitedBuffer(max int) *limitedBuffer {
	if max <= 0 {
		max = DefaultOutputLimitBytes
	}
	return &limitedBuffer{
		max:  max,
		data: make([]byte, 0, max),
	}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	written := len(p)
	remaining := b.max - len(b.data)
	if remaining > 0 {
		if len(p) <= remaining {
			b.data = append(b.data, p...)
		} else {
			b.data = append(b.data, p[:remaining]...)
			b.truncated = true
		}
	} else {
		b.truncated = true
	}
	return written, nil
}

func (b *limitedBuffer) WriteString(s string) {
	if s == "" {
		return
	}
	remaining := b.max - len(b.data)
	if remaining > 0 {
		if len(s) <= remaining {
			b.data = append(b.data, s...)
		} else {
			b.data = append(b.data, s[:remaining]...)
			b.truncated = true
		}
		return
	}
	b.truncated = true
}

func (b *limitedBuffer) String() string {
	if !b.truncated {
		return string(b.data)
	}
	const marker = "\n...[output truncated]"
	if len(b.data) >= len(marker) {
		prefix := string(b.data[:len(b.data)-len(marker)])
		return prefix + marker
	}
	return string(b.data)
}
