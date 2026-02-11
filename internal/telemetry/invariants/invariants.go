package invariants

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	// InvariantPatchApplyClean requires patches to apply without fuzz or rejects.
	InvariantPatchApplyClean = "patch_apply_clean"
	// InvariantRepoCleanBeforeMerge requires a clean repository before merge actions.
	InvariantRepoCleanBeforeMerge = "repo_clean_before_merge"
	// InvariantMaxRetriesNotExceeded requires retry/revision ceilings to remain within limits.
	InvariantMaxRetriesNotExceeded = "max_retries_not_exceeded"
	// InvariantEditsWithinAllowedPaths requires edits to remain inside allowed surface area.
	InvariantEditsWithinAllowedPaths = "edits_within_allowed_paths"
	// InvariantStateTransitionLegal requires lifecycle transitions to follow deterministic state machines.
	InvariantStateTransitionLegal = "state_transition_legal"
)

const (
	// SeverityWarn is used for non-fatal invariant violations.
	SeverityWarn = "warn"
	// SeverityError is used for fatal invariant violations.
	SeverityError = "error"
)

var invariantChecksEnabled atomic.Bool

func init() {
	invariantChecksEnabled.Store(true)
}

// ViolationDetails captures invariant violation context for telemetry events.
type ViolationDetails struct {
	WhatInvariant string
	WhereDetected string
	WhyViolated   string
	StackTrace    string
	Additional    map[string]string
}

// SetEnabled globally enables or disables invariant checks.
func SetEnabled(enabled bool) {
	invariantChecksEnabled.Store(enabled)
}

// Enabled reports whether invariant checks are currently enabled.
func Enabled() bool {
	return invariantChecksEnabled.Load()
}

// InvariantViolation emits an invariant.violation telemetry event on the active span.
// If the context has no active span, a short synthetic span is created for observability.
func InvariantViolation(
	ctx context.Context,
	invariantName string,
	severity string,
	details ViolationDetails,
) {
	if !Enabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	invariantName = strings.TrimSpace(invariantName)
	if invariantName == "" {
		invariantName = "unknown_invariant"
	}
	severity = normalizeSeverity(severity)

	attrs := []attribute.KeyValue{
		attribute.String("invariant_name", invariantName),
		attribute.String("severity", severity),
		attribute.String("what_invariant", strings.TrimSpace(details.WhatInvariant)),
		attribute.String("where_detected", strings.TrimSpace(details.WhereDetected)),
		attribute.String("why_violated", strings.TrimSpace(details.WhyViolated)),
	}
	if stack := strings.TrimSpace(details.StackTrace); stack != "" {
		attrs = append(attrs, attribute.String("stack_trace", stack))
	}

	if len(details.Additional) > 0 {
		keys := make([]string, 0, len(details.Additional))
		for key := range details.Additional {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := strings.TrimSpace(details.Additional[key])
			if value == "" {
				continue
			}
			attrs = append(attrs, attribute.String("context."+key, value))
		}
	}

	span := trace.SpanFromContext(ctx)
	if span != nil && span.SpanContext().IsValid() {
		span.AddEvent("invariant.violation", trace.WithAttributes(attrs...))
		return
	}

	tracedCtx, temporarySpan := otel.Tracer("sc3/invariants").Start(ctx, "invariant.violation")
	defer temporarySpan.End()
	temporarySpan.AddEvent("invariant.violation", trace.WithAttributes(attrs...))
	_ = tracedCtx
}

// CheckPatchApplyClean validates the patch_apply_clean invariant.
func CheckPatchApplyClean(ctx context.Context, whereDetected string, clean bool, why string) bool {
	if clean {
		return true
	}
	InvariantViolation(ctx, InvariantPatchApplyClean, SeverityError, ViolationDetails{
		WhatInvariant: "patch applied without fuzz or reject",
		WhereDetected: whereDetected,
		WhyViolated:   firstNonEmpty(why, "patch did not apply cleanly"),
	})
	return false
}

// CheckRepoCleanBeforeMerge validates the repo_clean_before_merge invariant.
func CheckRepoCleanBeforeMerge(ctx context.Context, whereDetected string, clean bool, statusPreview string) bool {
	if clean {
		return true
	}
	InvariantViolation(ctx, InvariantRepoCleanBeforeMerge, SeverityWarn, ViolationDetails{
		WhatInvariant: "repository has no uncommitted changes before merge",
		WhereDetected: whereDetected,
		WhyViolated:   firstNonEmpty(statusPreview, "repository contains uncommitted changes"),
	})
	return false
}

// CheckMaxRetriesNotExceeded validates the max_retries_not_exceeded invariant.
func CheckMaxRetriesNotExceeded(ctx context.Context, whereDetected string, retryCount, maxAllowed int) bool {
	if maxAllowed <= 0 || retryCount <= maxAllowed {
		return true
	}
	InvariantViolation(ctx, InvariantMaxRetriesNotExceeded, SeverityError, ViolationDetails{
		WhatInvariant: "retry/revision count remains within configured max",
		WhereDetected: whereDetected,
		WhyViolated:   fmt.Sprintf("retry_count=%d exceeded max_allowed=%d", retryCount, maxAllowed),
		Additional: map[string]string{
			"retry_count": fmt.Sprintf("%d", retryCount),
			"max_allowed": fmt.Sprintf("%d", maxAllowed),
		},
	})
	return false
}

// CheckEditsWithinAllowedPaths validates the edits_within_allowed_paths invariant.
func CheckEditsWithinAllowedPaths(
	ctx context.Context,
	whereDetected string,
	allowedPatterns []string,
	violatingPaths []string,
) bool {
	if len(allowedPatterns) == 0 {
		InvariantViolation(ctx, InvariantEditsWithinAllowedPaths, SeverityWarn, ViolationDetails{
			WhatInvariant: "agent edits remain within declared surface area",
			WhereDetected: whereDetected,
			WhyViolated:   "no allowed path patterns were provided",
		})
		return false
	}
	if len(violatingPaths) == 0 {
		return true
	}

	InvariantViolation(ctx, InvariantEditsWithinAllowedPaths, SeverityError, ViolationDetails{
		WhatInvariant: "agent edits remain within declared surface area",
		WhereDetected: whereDetected,
		WhyViolated:   fmt.Sprintf("detected edits outside allowed paths: %s", strings.Join(violatingPaths, ", ")),
		Additional: map[string]string{
			"violating_paths": strings.Join(violatingPaths, ","),
		},
	})
	return false
}

// CheckStateTransitionLegal validates the state_transition_legal invariant.
func CheckStateTransitionLegal(
	ctx context.Context,
	whereDetected string,
	entityType string,
	fromState string,
	toState string,
	legal bool,
) bool {
	if legal {
		return true
	}
	InvariantViolation(ctx, InvariantStateTransitionLegal, SeverityError, ViolationDetails{
		WhatInvariant: "state machine transition is legal",
		WhereDetected: whereDetected,
		WhyViolated:   fmt.Sprintf("illegal transition for entity=%s from=%s to=%s", entityType, fromState, toState),
		Additional: map[string]string{
			"entity_type": strings.TrimSpace(entityType),
			"from_state":  strings.TrimSpace(fromState),
			"to_state":    strings.TrimSpace(toState),
		},
	})
	return false
}

func normalizeSeverity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case SeverityWarn:
		return SeverityWarn
	case SeverityError:
		return SeverityError
	default:
		return SeverityError
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
