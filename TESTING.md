# Go Testing Guidelines - Anti-Vanity Testing

**Purpose**: Prevent AI agents from generating useless Go tests. Enforce behavior-focused testing idiomatic to Go.

**Audience**: All AI agents (Captain, Commander, Design Officer, Ensigns, Codex)

**Based On**:
- `~/Projects/brain/docs/AGENTIC-TESTING-RULES.md` - 5 Test Integrity Rules
- `~/Projects/brain/docs/TEST-RESEARCH-SUMMARY.md` - Test planning research
- `.spec/research-findings/GO_CODING_STANDARDS.md` - Go idioms and patterns

---

## 🚨 AGENT TESTING PROGRAM DISCLOSURE (MUST READ)

**This section persists across all agent sessions to preserve testing standards.**

### What You MUST Do When Writing Tests

1. **ALWAYS test BEHAVIOR, never implementation**
   - ❌ DON'T: Test that struct fields exist
   - ✅ DO: Test operations that use those fields produce correct results

2. **NEVER write self-validating tests**
   - ❌ DON'T: Create object, then assert same object's properties
   - ✅ DO: Create object through API, verify DB state changed

3. **ALWAYS verify actual state**
   - ❌ DON'T: `mockDB.AssertCalled("Save")`
   - ✅ DO: `savedUser := db.Get(id); assert.Equal(t, user.Email, savedUser.Email)`

4. **Use REALISTIC test data**
   - ❌ DON'T: `email: "a@b.c"`, `name: "x"`
   - ✅ DO: `email: "jane.doe@company.com"`, `name: "Jane Doe"`

5. **Follow Go test file organization**
   - Tests go NEXT to code: `internal/mypackage/mypackage_test.go`
   - Use `test/` directory only for shared helpers and integration tests

### Quick Anti-Pattern Detection

If your test does ANY of these, it's WRONG:
- [ ] Creates data inline and asserts on that same data
- [ ] Only verifies "mock was called"
- [ ] Tests struct field existence (`assert.NotNil(t, user.Field)`)
- [ ] Tests that struct implements interface (Go compiler does this)
- [ ] Uses arbitrary/unrealistic test data
- [ ] Would pass even if feature is completely broken

### Before Marking Work Complete

```bash
# 1. Run tests
make test

# 2. Check coverage
make test-coverage  # Must be >80%

# 3. Verify no vanity tests
# Ask: "If I delete the implementation, does this test fail?"
# If NO → Test is vanity, rewrite it

# 4. Lint
make lint
```

### This Is MANDATORY

- These rules are NON-NEGOTIABLE
- Tests that violate these principles MUST be rewritten
- "But the test passes" is NOT an excuse for vanity tests
- Reference this document in every testing discussion

---

## The Core Problem in Go

AI agents default to generating Go tests that:
- Test struct fields instead of behavior
- Create test data inline and assert on the same data (self-validating)
- Verify mock calls without checking actual state
- Test exported methods in isolation without verifying outcomes
- Write trivial tests that always pass

**Result**: Tests provide false confidence while catching zero bugs.

---

## The 5 Test Integrity Rules (Adapted for Go)

### Rule 1: Test REQUIREMENTS, Not Implementation

| ❌ Wrong (Vanity) | ✅ Right (Valuable) |
|------------------|-------------------|
| `assert.Equal(t, "string", u.EmailField)` | Tests struct field exists | `assert.True(t, u.EmailVerified)` | Tests behavior |
| `assert.NotNil(t, user.Repo)` | Tests dependency was set | `savedUser, _ := db.Get(id)` <br> `assert.Equal(t, user.Email, savedUser.Email)` | Tests actual persistence |

**Question**: If this implementation was refactored (same behavior), would the test still pass?
- **YES** → Good test (tests behavior)
- **NO** → Bad test (tests implementation)

### Rule 2: Never Change Tests Without Investigation

When tests fail:
1. Read the acceptance criteria
2. Determine if TEST or CODE deviates from spec
3. Fix whichever deviates

**NEVER** modify tests just to make failing code pass.

### Rule 3: Use Meaningful, Realistic Test Data

| ❌ Wrong | ✅ Right |
|---------|---------|
| `email: "a@b.c"` | `email: "jane.doe@company.com"` |
| `name: "x"` | `name: "Jane Doe"` |
| `price: 1` | `price: 149.99` |
| `id: "123"` | `id: "user_abc123xyz"` |

Realistic data catches edge cases that arbitrary values miss.

### Rule 4: Ensure Test Isolation

Each test must:
- Run independently (use `t.Parallel()` when safe)
- Not depend on previous test state
- Use setup/teardown functions for initialization
- Pass in any execution order

### Rule 5: Apply Equal Scrutiny to Test and Code

When tests fail, treat both with suspicion:
- Test might be wrong
- Code might be wrong
- Spec might be unclear

Investigate before assuming.

---

## Go-Specific Anti-Patterns

### ❌ Anti-Pattern 1: Self-Validating Tests

```go
// VANITY: Creates struct, asserts same struct fields
func TestUserEmail(t *testing.T) {
    user := &User{
        Email: "test@example.com",  // Created here
    }

    assert.Equal(t, "test@example.com", user.Email)  // Always passes!
    assert.NotNil(t, user.Email)  // Trivial!
}
```

**Why it's bad**: Test passes even if `User` struct is broken or unused.

```go
// VALUABLE: Tests actual code execution
func TestUserCreation(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    svc := NewUserService(db)

    // Create user through actual API
    userID, err := svc.CreateUser(context.Background(), CreateUserRequest{
        Email: "jane.doe@company.com",
        Name:  "Jane Doe",
    })
    require.NoError(t, err)

    // Verify actual database state
    user, err := db.GetByID(userID)
    require.NoError(t, err)
    assert.Equal(t, "jane.doe@company.com", user.Email)
    assert.Equal(t, "Jane Doe", user.Name)
    assert.WithinDuration(t, time.Now(), user.CreatedAt, 2*time.Second)
}
```

### ❌ Anti-Pattern 2: Mock-Only Verification

```go
// VANITY: Only verifies mock was called
func TestSaveUser(t *testing.T) {
    mockDB := new(MockDatabase)
    mockDB.On("Save", mock.Anything).Return(nil)

    svc := NewUserService(mockDB)
    svc.SaveUser(&User{Email: "test@example.com"})

    mockDB.AssertCalled(t, "Save", mock.Anything)  // Data might not be saved!
}
```

**Why it's bad**: Test passes even if `SaveUser` has bugs in data processing.

```go
// VALUABLE: Verifies actual outcome
func TestSaveUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    svc := NewUserService(db)
    user := &User{Email: "jane.doe@company.com", Name: "Jane Doe"}

    err := svc.SaveUser(context.Background(), user)
    require.NoError(t, err)

    // Verify actual database state
    saved, err := db.GetByEmail("jane.doe@company.com")
    require.NoError(t, err)
    assert.Equal(t, "Jane Doe", saved.Name)
    assert.False(t, saved.CreatedAt.IsZero())
}
```

### ❌ Anti-Pattern 3: Testing Interface Satisfaction

```go
// VANITY: Tests that struct implements interface
func TestUserRepositoryImplementsInterface(t *testing.T) {
    var _ Repository = (*UserRepository)(nil)  // Compile-time check, useless test
    assert.True(t, true)  // Trivial assertion
}
```

**Why it's bad**: Go compiler already checks interface satisfaction at compile time.

```go
// VALUABLE: Tests actual repository behavior
func TestUserRepositoryCRUD(t *testing.T) {
    repo := NewUserRepository(setupTestDB(t))

    // Create
    user := &User{ID: "user-123", Email: "jane@company.com"}
    err := repo.Save(context.Background(), user)
    require.NoError(t, err)

    // Read
    found, err := repo.FindByID(context.Background(), "user-123")
    require.NoError(t, err)
    assert.Equal(t, "jane@company.com", found.Email)

    // Update
    found.Email = "jane.doe@company.com"
    err = repo.Save(context.Background(), found)
    require.NoError(t, err)

    // Delete
    err = repo.Delete(context.Background(), "user-123")
    require.NoError(t, err)

    // Verify deleted
    _, err = repo.FindByID(context.Background(), "user-123")
    assert.Error(t, err)
}
```

### ❌ Anti-Pattern 4: Trivial Error Checks

```go
// VANITY: Tests that error is returned
func TestValidateEmail(t *testing.T) {
    err := ValidateEmail("")
    assert.Error(t, err)  // Doesn't verify what went wrong!
}
```

**Why it's bad**: Test passes for any error, doesn't verify correct validation.

```go
// VALUABLE: Tests specific validation behavior
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name      string
        email     string
        wantError bool
        errMsg    string
    }{
        {
            name:      "valid email",
            email:     "jane.doe@company.com",
            wantError: false,
        },
        {
            name:      "empty email",
            email:     "",
            wantError: true,
            errMsg:    "email cannot be empty",
        },
        {
            name:      "invalid format",
            email:     "not-an-email",
            wantError: true,
            errMsg:    "invalid email format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)

            if tt.wantError {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## The Go Testing Pyramid

```
                    ┌─────────────────┐
                    │   E2E Tests      │  10% - Critical CLI workflows
                    │  (Full binary)    │  Slow, expensive, high fidelity
                    └───────────────────┘
                  ┌─────────────────────────────┐
                  │  Integration Tests          │  20% - Component interactions
                  │  (Test DB, Real deps)       │  Medium speed, medium cost
                  └─────────────────────────────┘
                ┌──────────────────────────────────┐
                │      Unit Tests                   │  70% - Business logic
                │      (Pure functions, no mocks)   │  Fast, cheap, isolated
                └──────────────────────────────────┘
```

### Unit Tests (70%) - Pure Functions & Business Logic

**Test these**:
- Validation logic
- Data transformations
- State machine transitions
- String formatting
- Calculations

**Example**:
```go
func TestCalculateCommission(t *testing.T) {
    tests := []struct {
        name           string
        amount         float64
        rate           float64
        expectedResult float64
    }{
        {
            name:           "standard commission",
            amount:         1000.00,
            rate:           0.15,
            expectedResult: 150.00,
        },
        {
            name:           "zero amount",
            amount:         0,
            rate:           0.15,
            expectedResult: 0,
        },
        {
            name:           "maximum rate",
            amount:         5000.00,
            rate:           0.25,  // Max rate
            expectedResult: 1250.00,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateCommission(tt.amount, tt.rate)
            assert.InDelta(t, tt.expectedResult, result, 0.01)
        })
    }
}
```

### Integration Tests (20%) - Database & External Services

**Test these**:
- Database queries and mutations
- File system operations
- HTTP handlers (with test server)
- Component integration points

**Example**:
```go
func TestMissionRepository(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    repo := NewMissionRepository(db)

    t.Run("create and retrieve mission", func(t *testing.T) {
        mission := &Mission{
            ID:          "mission-123",
            CommissionID: "commission-abc",
            Title:       "Implement login",
            Status:      "pending",
        }

        err := repo.Create(context.Background(), mission)
        require.NoError(t, err)

        found, err := repo.GetByID(context.Background(), "mission-123")
        require.NoError(t, err)
        assert.Equal(t, "Implement login", found.Title)
        assert.Equal(t, "commission-abc", found.CommissionID)
    })

    t.Run("update mission status", func(t *testing.T) {
        // Create mission
        mission := &Mission{
            ID:     "mission-456",
            Status: "pending",
        }
        err := repo.Create(context.Background(), mission)
        require.NoError(t, err)

        // Update status
        err = repo.UpdateStatus(context.Background(), "mission-456", "in_progress")
        require.NoError(t, err)

        // Verify
        found, err := repo.GetByID(context.Background(), "mission-456")
        require.NoError(t, err)
        assert.Equal(t, "in_progress", found.Status)
    })
}
```

### E2E Tests (10%) - Critical CLI Workflows

**Test these** ONLY:
- Revenue-critical paths (if applicable)
- Security-critical flows (authentication, authorization)
- Multi-step workflows (plan → execute → complete)
- Integration points with external systems

**Example**:
```go
func TestE2E_CommissionToMissionFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test in short mode")
    }

    // 1. Create commission from PRD
    prdPath := "testdata/sample-commission.md"
    result := runCommand(t, "sc3", "plan", "create", "--prd", prdPath)
    assert.Equal(t, 0, result.ExitCode)
    commissionID := extractID(result.Stdout)

    // 2. Plan missions
    result = runCommand(t, "sc3", "plan", "run", commissionID)
    assert.Equal(t, 0, result.ExitCode)

    // 3. Verify missions created in database
    db := openTestDatabase(t)
    missions, err := db.FindMissionsByCommission(commissionID)
    require.NoError(t, err)
    assert.Greater(t, len(missions), 0)

    // 4. Execute first mission
    result = runCommand(t, "sc3", "execute", missions[0].ID)
    assert.Equal(t, 0, result.ExitCode)

    // 5. Verify mission completed
    mission, err := db.GetMission(missions[0].ID)
    require.NoError(t, err)
    assert.Equal(t, "completed", mission.Status)
}
```

---

## Mocking Strategy for Go

### Decision Tree

```
Can I use a real implementation in a test?
│
├─ Database → Use test database (SQLite or in-memory)
├─ File system → Use temp directory (test.TempDir)
├─ External API → MOCK (unreliable, slow, has costs)
├─ HTTP handlers → Use httptest.NewRecorder
└─ Business logic → NEVER mock (test the real thing!)
```

### When to Mock in Go

| Dependency | ✅ Mock or ❌ Real? | Reason |
|------------|-------------------|---------|
| **Database** | ❌ Use test DB | Verifies actual queries/transactions |
| **File system** | ❌ Use temp dir | Real I/O is fast enough |
| **HTTP handlers** | ❌ Use httptest | Tests request/response handling |
| **External APIs** | ✅ Mock | Unreliable, slow, has costs |
| **Business logic** | ❌ Never mock | Test real implementation |

### Mocking Examples

#### ❌ WRONG: Mock Database

```go
func TestUserService(t *testing.T) {
    mockDB := new(MockDatabase)
    mockDB.On("GetUser", "user-123").Return(&User{Email: "test@example.com"}, nil)

    svc := NewUserService(mockDB)
    user, err := svc.GetUser("user-123")

    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
    mockDB.AssertCalled(t, "GetUser", "user-123")  // Vanity!
}
```

#### ✅ RIGHT: Use Test Database

```go
func TestUserService(t *testing.T) {
    db := setupTestDB(t)  // Real SQLite or in-memory DB
    defer db.Close()

    // Seed test data
    _, err := db.Exec("INSERT INTO users (id, email) VALUES (?, ?)", "user-123", "jane@company.com")
    require.NoError(t, err)

    svc := NewUserService(db)
    user, err := svc.GetUser(context.Background(), "user-123")

    require.NoError(t, err)
    assert.Equal(t, "jane@company.com", user.Email)
}
```

#### ✅ RIGHT: Mock External API

```go
func TestExternalAPICaller(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/users/123", r.URL.Path)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"id": "123", "name": "Jane"}`))
    }))
    defer server.Close()

    client := NewAPIClient(server.URL)
    user, err := client.GetUser("123")

    require.NoError(t, err)
    assert.Equal(t, "Jane", user.Name)
}
```

---

## Test Quality Checklist for Go

Before submitting a test file, verify:

### Coverage Checklist
- [ ] Happy path (normal usage) is tested
- [ ] At least 2 edge cases are tested
- [ ] At least 1 error case is tested
- [ ] All acceptance criteria have corresponding tests
- [ ] Test names describe WHAT is tested, not HOW

### Anti-Pattern Checklist
- [ ] No tests verify "mock was called" (use test DB instead)
- [ ] No tests verify struct field existence (use real operations)
- [ ] No tests assert on unexported fields (test exported behavior)
- [ ] No tests are self-validating (create and assert same data)
- [ ] No tests depend on execution order (use `t.Parallel()` properly)

### Assertion Quality
- [ ] Every test has at least 1 meaningful assertion
- [ ] Assertions verify observable outcomes (DB state, return values)
- [ ] Error assertions verify error messages, not just "error occurred"
- [ ] Table-driven tests use descriptive test names

---

## Given-When-Then Template for Go

```go
func TestMissionExecution(t *testing.T) {
    // GIVEN: A mission in pending state with a surface area
    t.Run("execute mission successfully", func(t *testing.T) {
        // GIVEN
        db := setupTestDB(t)
        defer db.Close()

        mission := &Mission{
            ID:          "mission-123",
            Status:      "pending",
            SurfaceArea: []string{"internal/auth"},
        }
        err := db.CreateMission(context.Background(), mission)
        require.NoError(t, err)

        commander := NewCommander(db)

        // WHEN
        err = commander.ExecuteMission(context.Background(), "mission-123")

        // THEN
        require.NoError(t, err)

        updated, err := db.GetMission(context.Background(), "mission-123")
        require.NoError(t, err)
        assert.Equal(t, "completed", updated.Status)
        assert.False(t, updated.CompletedAt.IsZero())
    })
}
```

---

## Quick Reference: Test Value Detection

| Test Pattern | Classification | Fix |
|--------------|----------------|-----|
| `assert.Equal(t, created, asserted)` | ❌ Vanity | Verify through operation (DB, API) |
| `mock.AssertCalled("Method")` | ❌ Vanity | Verify actual state change |
| `assert.True(t, true)` | ❌ Vanity | Test actual behavior |
| Test struct field exists | ❌ Vanity | Test field usage in operation |
| Test interface satisfaction | ❌ Vanity | Test actual interface behavior |
| Verify DB state after operation | ✅ Valuable | Keep as-is |
| Test with realistic data | ✅ Valuable | Keep as-is |
| Test error messages | ✅ Valuable | Keep as-is |

---

## Sources & References

1. **Internal Testing Standards**:
   - `~/Projects/brain/docs/AGENTIC-TESTING-RULES.md` - 5 Test Integrity Rules
   - `~/Projects/brain/docs/TEST-RESEARCH-SUMMARY.md` - Test planning research
   - `.spec/research-findings/GO_CODING_STANDARDS.md` - Go idioms

2. **External Sources**:
   - [Martin Fowler: Mocks Aren't Stubs](https://martinfowler.com/articles/mocksArentStubs.html)
   - [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
   - [Testify Documentation](https://github.com/stretchr/testify)

---

**Version**: 1.0
**Last Updated**: 2025-02-10
**Aligned With**: Go 1.22+ and Ship Commander 3 standards
