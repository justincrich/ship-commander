# Testing Infrastructure Summary

## ✅ What Changed

### Moved Testing Docs to Root

**Before**: `.spec/GO-TESTING-GUIDELINES.md` (buried in specs)
**After**: `TESTING.md` (root level, alongside AGENTS.md)

**Why**:
- `.spec/` feels like "finished requirements" not "ongoing practices"
- Root-level `TESTING.md` is more discoverable for all agents
- Matches the pattern: `AGENTS.md`, `DEV_GUIDE.md`, `TESTING.md`

## 📁 Go Test File Organization (Clarified)

### Where Tests Go

```
ship-commander-3/
├── internal/
│   ├── beads/
│   │   ├── client.go              # Implementation
│   │   ├── client_test.go         # ✅ Unit tests HERE (co-located)
│   │   └── client_external_test.go # ✅ Black-box tests HERE
│   ├── execution/
│   │   ├── commander.go
│   │   └── commander_test.go      # ✅ Unit tests HERE
│   └── tui/
│       ├── views/
│       │   ├── mission.go
│       │   └── mission_test.go    # ✅ Unit tests HERE
│
├── test/                           # Only for SHARED utilities
│   ├── test_helpers.go             # Shared test functions
│   ├── example_test.go             # Table-driven test examples
│   ├── AGENTS.md                   # Test package guidance
│   └── integration/                # Cross-package integration tests
│       └── e2e_test.go
│
└── TESTING.md                      # ⭐ Main testing guide (root level)
```

### Key Point

**Tests live NEXT to code**, not in a single test directory:
- `internal/mypackage/mypackage_test.go` ✅
- `test/` is only for **shared helpers** and **integration tests**

## 🚨 Agent Memory Preservation

### Testing Program Disclosure (NEW)

Added to top of `TESTING.md` - a persistent memory section that:

1. **States what agents MUST do** when writing tests
2. **Quick anti-pattern detection checklist** (6 questions)
3. **Quality gate commands** to run before completion
4. **Clearly states this is MANDATORY and NON-NEGOTIABLE**

### Section Contents

```markdown
## 🚨 AGENT TESTING PROGRAM DISCLOSURE (MUST READ)

### What You MUST Do When Writing Tests
1. ALWAYS test BEHAVIOR, never implementation
2. NEVER write self-validating tests
3. ALWAYS verify actual state
4. Use REALISTIC test data
5. Follow Go test file organization

### Quick Anti-Pattern Detection
Checklist of 6 vanity test patterns

### Before Marking Work Complete
Commands to run: test, coverage, lint

### This Is MANDATORY
Clear statement that rules are non-negotiable
```

## 📊 Documentation Hierarchy

```
Root Level (High Visibility)
├── AGENTS.md
│   └─ References: "All agents MUST read TESTING.md"
│
├── TESTING.md ⭐ (Main testing guide)
│   ├─ Agent Testing Program Disclosure (persists across sessions)
│   ├─ 5 Test Integrity Rules
│   ├─ Go-specific anti-patterns
│   ├─ Mocking strategy
│   └─ Test quality checklist
│
└── DEV_GUIDE.md
    └─ Development workflow (make dev, test, lint)

Package Level (Specific Guidance)
├── test/AGENTS.md
│   └─ "ALL agents MUST read TESTING.md"
│
└── internal/*/AGENTS.md
    └─ Package-specific testing patterns

Reference Material
├── ~/Projects/brain/docs/AGENTIC-TESTING-RULES.md (Universal principles)
└── .spec/research-findings/GO_CODING_STANDARDS.md (Go idioms)
```

## 🎯 For Codex Agents

Every agent writing tests now:

1. **Sees mandatory requirement** in root `AGENTS.md`
2. **Opens `TESTING.md`** (root level, easy to find)
3. **Sees Agent Disclosure** at top (persistent memory)
4. **Checks anti-pattern checklist** before writing
5. **References examples** in `test/example_test.go`
6. **Uses helpers** from `test/test_helpers.go`

## 🔒 Enforcement Mechanism

### Root AGENTS.md (Highest Authority)
```
### Testing Requirements (MANDATORY)

**All agents MUST read and follow**: TESTING.md

- Core anti-vanity rules (5 rules from brain docs)
- Go-specific requirements (no mock-only verification)
- Quality gates (test, coverage, lint before completion)
```

### Before Work Complete
```bash
make test              # All tests pass
make test-coverage     # >80% coverage
make lint              # No linter warnings
```

### Anti-Pattern Detection

Tests flagged as vanity if they:
- Create data and assert same data
- Only verify "mock was called"
- Test struct field existence
- Would pass even if feature is broken

## 📚 Quick Reference

| Document | Location | Purpose |
|----------|----------|---------|
| **TESTING.md** | Root | Main testing guide (MANDATORY) |
| **test/AGENTS.md** | test/ | Test package guidance |
| **test/example_test.go** | test/ | Table-driven test examples |
| **test/test_helpers.go** | test/ | Shared test utilities |
| **DEV_GUIDE.md** | Root | Development workflow |

## ✅ Benefits of New Structure

1. **Discoverability**: `TESTING.md` at root level is easy to find
2. **Agent Memory**: Disclosure section persists across sessions
3. **Correct Organization**: Matches Go practice (tests co-located with code)
4. **Clear Hierarchy**: Root → Package → Reference material
5. **Enforcement**: Multiple references ensure agents see requirements

---

**Version**: 1.0
**Last Updated**: 2025-02-10
