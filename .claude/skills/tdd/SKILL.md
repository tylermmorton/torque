---
name: tdd
description: Test-driven development with red-green-refactor loop. Use when user wants to build features or fix bugs using TDD, mentions "red-green-refactor", wants integration tests, or asks for test-first development.
---

# Test-Driven Development

## Philosophy

**Core principle**: Tests should verify behavior through public interfaces, not implementation details. Code can change entirely; tests shouldn't.

**Good tests** are integration-style: they exercise real code paths through public APIs. They describe _what_ the system does, not _how_ it does it. A good test reads like a specification — "user can checkout with valid cart" tells you exactly what capability exists. These tests survive refactors because they don't care about internal structure.

**Bad tests** are coupled to implementation. They mock internal collaborators, test private functions, or assert on call counts. The warning sign: your test breaks when you refactor, but behavior hasn't changed.

See [tests.md](tests.md) for examples and [mocking.md](mocking.md) for mocking guidelines.

## Anti-Pattern: Horizontal Slices

**DO NOT write all tests first, then all implementation.** This is "horizontal slicing" — treating RED as "write all tests" and GREEN as "write all code."

This produces **crap tests**:

- Tests written in bulk test _imagined_ behavior, not _actual_ behavior
- You end up testing the _shape_ of things (data structures, function signatures) rather than user-facing behavior
- Tests become insensitive to real changes — they pass when behavior breaks, fail when behavior is fine
- You outrun your headlights, committing to test structure before understanding the implementation

**Correct approach**: Vertical slices via tracer bullets. One test → one implementation → repeat. Each test responds to what you learned from the previous cycle. Because you just wrote the code, you know exactly what behavior matters and how to verify it.

```
WRONG (horizontal):
  RED:   test1, test2, test3, test4, test5
  GREEN: impl1, impl2, impl3, impl4, impl5

RIGHT (vertical):
  RED→GREEN: test1→impl1
  RED→GREEN: test2→impl2
  RED→GREEN: test3→impl3
  ...
```

## Workflow

### 1. Planning

Before writing any code:

- Read the relevant `CONTEXT.md` — use the domain glossary for test names and interface vocabulary
- Confirm with user what interface changes are needed
- List the behaviors to test, in priority order (not implementation steps)
- Identify opportunities for [deep modules](deep-modules.md) (small interface, deep implementation)
- Design interfaces for [testability](interface-design.md)
- Get user approval on the list before starting

Ask: "What should the public interface look like? Which behaviors are most important to test first?"

**You can't test everything.** Confirm exactly which behaviors matter most. Focus on critical paths and complex logic.

### 2. Tracer Bullet

Write ONE test that confirms ONE thing about the system.

**RED — write the test, confirm it fails:**

```
go test -run TestName ./path/to/package/...
```

The test must fail — either a compile error (if the interface doesn't exist yet) or a test failure. A compile error counts as RED. **Do not proceed if the test passes without implementation — it means the test isn't testing anything.**

**GREEN — write minimal code to pass:**

```
go test -run TestName ./path/to/package/...
```

The test must pass. Only then move to the next behavior.

This is your tracer bullet — proves the path works end-to-end.

### 3. Incremental Loop

For each remaining behavior:

```
RED:   Write next test → run it → confirm failure
GREEN: Minimal code to pass → run it → confirm pass
```

Rules:

- One test at a time
- Only enough code to pass the current test
- Don't anticipate future tests
- Show the test output at each RED and GREEN step — do not assume

### 4. Refactor

After all tests pass, look for [refactor candidates](refactoring.md):

- [ ] Extract duplication
- [ ] Deepen modules (move complexity behind simple interfaces)
- [ ] Apply SOLID principles where natural
- [ ] Consider what new code reveals about existing code
- [ ] Run tests after each refactor step

**Never refactor while RED.** Get to GREEN first.

## Test Structure

Top-level test functions follow the pattern `TestType_Method`. Sub-behaviors use `t.Run()` with snake_case names:

```go
func TestUser_Create(t *testing.T) {
    // happy path inline

    t.Run("duplicate_email_returns_error", func(t *testing.T) {
        // ...
    })
}
```

## Checklist Per Cycle

```
[ ] Test describes behavior, not implementation
[ ] Test uses public interface only
[ ] Test would survive internal refactor
[ ] RED confirmed: test runner output shows failure
[ ] GREEN confirmed: test runner output shows pass
[ ] Code is minimal for this test
[ ] No speculative features added
```