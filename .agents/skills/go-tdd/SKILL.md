---
name: go-tdd
description: Go TDD patterns for Cubby tests
---

## When to use

Use for Go tests in backend/ with stdlib `testing`, SQLite in-memory DB, and Cubby query patterns.

## Red → Green → Tidy

```bash
cd backend && go test ./internal/db/ -run TestCreateLocation -v
cd backend && go test ./...
```

## Test naming

`TestFunc` for happy path; `TestFunc_Condition` for edge/error cases.

## Table-driven tests with subtests

```go
func TestCreateLocation(t *testing.T) {
    db := setupTestDB(t)
    q := NewQueries(db)

    tests := []struct {
        name    string
        req     models.CreateLocationRequest
        wantErr bool
    }{
        {"valid", models.CreateLocationRequest{Name: "Garage"}, false},
        {"missing name", models.CreateLocationRequest{Name: ""}, true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            loc, err := q.CreateLocation(tc.req)
            if (err != nil) != tc.wantErr {
                t.Errorf("err = %v, wantErr %v", err, tc.wantErr)
            }
            if err == nil && loc.Name != tc.req.Name {
                t.Errorf("name = %q, want %q", loc.Name, tc.req.Name)
            }
        })
    }
}
```

## t.Fatalf vs t.Errorf

Use `t.Fatalf` for setup failures that make the test invalid.
Use `t.Errorf` for assertion failures when the test can continue.

## Test helpers with t.Helper()

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatalf("open db: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}
```

## t.Cleanup vs defer

`t.Cleanup` runs after the test and after subtests, in LIFO order.
`defer` runs at the end of the current function only.
Prefer `t.Cleanup` when tests use subtests or helpers.

## t.Parallel()

Safe when tests share no mutable state and each has its own DB.
Unsafe with global state, env vars, or shared DB connections.
Call inside each `t.Run` to parallelize subtests.

## HTTP handler testing with httptest

```go
func TestGetLocationHandler(t *testing.T) {
    req := httptest.NewRequest(http.MethodGet, "/locations/1", nil)
    rec := httptest.NewRecorder()

    http.HandlerFunc(GetLocationHandler).ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
    }
}
```

## Testing error responses

```go
func TestCreateLocationHandler_InvalidJSON(t *testing.T) {
    req := httptest.NewRequest(http.MethodPost, "/locations", strings.NewReader("{"))
    rec := httptest.NewRecorder()

    http.HandlerFunc(CreateLocationHandler).ServeHTTP(rec, req)

    if rec.Code != http.StatusBadRequest {
        t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
    }
}
```

## Test file organization

```text
backend/
  internal/
    db/
      queries.go
      queries_test.go
    api/
      handlers.go
      handlers_test.go
```

## Integration tests with build tags

Use `//go:build integration` in *_test.go.
Run with `go test -tags=integration ./...`.
Default `go test ./...` runs unit tests only.

## Benchmarks (Go 1.24+)

```go
func BenchmarkCreateLocation(b *testing.B) {
    db := setupTestDB(b)
    q := NewQueries(db)

    for b.Loop() {
        _, _ = q.CreateLocation(models.CreateLocationRequest{Name: "Garage"})
    }
}
```

## Rules

- Use stdlib testing only; no assertion libraries.
- Keep tests in *_test.go next to code.
- Use in-memory SQLite (`:memory:`) for DB tests.
- Use table-driven tests with `t.Run` for related cases.
- Wrap setup helpers with `t.Helper()`.
- Error messages should include got vs want.

## Don'ts

- Don’t share DB connections across parallel tests.
- Don’t use `defer` for cleanup in subtest-heavy suites.
- Don’t assert via external libraries or reflection-heavy helpers.
