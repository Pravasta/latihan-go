# Sprint 04 — Code Review Guide

Act as a Senior Golang Backend Engineer.

Review the implementation against `docs/sprint-04-task.md`.

Do not modify the implementation before completing the initial review.

---

# 1. Requirement Compliance

Review every:

- Functional requirement
- Validation rule
- HTTP status requirement
- Architecture requirement
- Testing requirement
- Acceptance criterion

Classify each important requirement as:

- PASS
- PARTIAL
- FAIL

---

# 2. HTTP Review

Review:

- REST endpoint design
- HTTP method handling
- Status codes
- Request body decoding
- Invalid JSON handling
- Path parameter parsing
- Query parameter parsing
- Response Content-Type
- Empty response body for 204
- Method Not Allowed behavior

Check whether the implementation follows HTTP semantics correctly.

---

# 3. Architecture Review

Verify the flow:

```text
HTTP
→ Handler
→ Service
→ Storage Interface
→ JSON File
```

Check for:

- Business logic leaking into handlers
- HTTP concerns leaking into service
- Persistence logic leaking into service
- Concrete storage created inside service or handler
- Missing dependency injection

---

# 4. Service Review

Review:

- ID generation
- Validation
- Search behavior
- Case-insensitive search
- Update behavior
- CreatedAt preservation
- UpdatedAt modification
- Delete behavior

---

# 5. Storage Review

Review:

- Correct JSON persistence
- Correct file handling
- Corrupted JSON behavior
- Error propagation
- Test isolation using temporary files

---

# 6. Error Handling Review

Check:

- Sentinel errors
- `errors.Is`
- Internal error leakage
- Incorrect nil error returns
- Ignored errors
- Panic usage

---

# 7. Testing Review

Review:

- Service tests
- Table-driven tests
- Storage tests
- Handler tests using `httptest`
- Status code assertions
- Response body assertions
- Error scenarios
- Coverage quality

Do not judge only by the coverage percentage.

Identify important untested behavior.

---

# 8. Security and Robustness Review

For the scope of this sprint, check:

- Whether malformed JSON can crash the server
- Whether unexpected methods are handled
- Whether invalid IDs are handled
- Whether internal filesystem errors leak to clients
- Whether request body handling is reasonably safe

Do not require authentication or advanced production security because those are outside Sprint 04 scope.

---

# 9. Required Review Output

## Requirement Compliance

List important PASS, PARTIAL, and FAIL items.

## Major Issues

List issues that cause:

- Incorrect API behavior
- Wrong status codes
- Data corruption
- Architecture violations
- Requirement failures

## Minor Issues

List:

- Naming issues
- Readability improvements
- Small refactoring opportunities
- Non-critical HTTP improvements

## Missing Tests

List meaningful missing test scenarios.

## Scores

- Requirement Compliance: /10
- HTTP Fundamentals: /10
- Architecture: /10
- Service Logic: /10
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

Do not give PASS if a core API endpoint is broken or a major architecture requirement is violated.
