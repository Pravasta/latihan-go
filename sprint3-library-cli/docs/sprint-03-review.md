# Sprint 03 — Code Review Guide

Act as a Senior Golang Engineer.

Review the implementation against `docs/sprint-03-task.md`.

Do not modify the implementation before completing the review.

---

# 1. Requirement Compliance

Check every Functional Requirement, Validation Rule, Business Rule, Technical Requirement, and Acceptance Criterion.

For each requirement, classify it as:

- PASS
- PARTIAL
- FAIL

Mention the relevant file and code when an issue is found.

---

# 2. Architecture Review

Review:

- Separation between CLI, service, and storage
- `BookStorage` abstraction
- `LoanStorage` abstraction
- Dependency injection
- Whether business logic leaks into storage
- Whether persistence logic leaks into service
- Whether CLI contains domain business logic

---

# 3. Domain Logic Review

Pay special attention to:

- Unique ID generation
- Book availability state
- Active loan rules
- Returning books
- `ReturnedAt` handling
- Consistency between Book and Loan data

Check for scenarios where `books.json` and `loans.json` could become inconsistent.

---

# 4. Error Handling Review

Check:

- Ignored errors
- Incorrect `return err` where `err` may be nil
- Sentinel error usage
- Error wrapping
- Error comparison using `errors.Is`
- Panic usage

---

# 5. Testing Review

Check:

- Test isolation
- Fake storage implementation
- Table-driven tests
- Happy paths
- Failure paths
- Edge cases
- Test quality
- Coverage

Identify important missing test cases.

---

# 6. Code Quality Review

Check:

- Idiomatic Go
- Naming
- Function responsibilities
- Code duplication
- Over-engineering
- Under-engineering
- Readability

---

# 7. Required Review Output

## Requirement Compliance

List PASS, PARTIAL, and FAIL items.

## Major Issues

Issues that can cause:

- Incorrect behavior
- Data inconsistency
- Requirement failure
- Architecture violation

## Minor Issues

Issues related to:

- Naming
- Readability
- Small refactoring opportunities
- Non-critical improvements

## Missing Tests

List missing test scenarios.

## Scores

- Requirement Compliance: /10
- Architecture: /10
- Business Logic: /10
- Error Handling: /10
- Testing: /10
- Code Quality: /10

## Final Score

Score: /10

## Final Verdict

Choose exactly one:

- PASS
- PASS WITH MINOR IMPROVEMENTS
- NEEDS REVISION
- REJECTED

Explain the verdict briefly.

Do not give PASS if any major functional requirement or business rule is broken.
