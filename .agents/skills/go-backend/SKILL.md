---
name: go-backend
description: Go backend conventions for Cubby — net/http 1.22+ routing, service layer, middleware, JSON patterns
---

## When to use

Use when adding or modifying Go code in `backend/`. Covers HTTP handlers,
service layer, middleware, JSON encoding, and the request lifecycle.

## No external libraries

Always prefer stdlib. Before adding any dependency, explain why it's needed and
list alternatives. The only approved external dep is `modernc.org/sqlite`.

## Router setup (Go 1.22+ enhanced ServeMux)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /v1/locations", s.handleListLocations)
mux.HandleFunc("POST /v1/locations", s.handleCreateLocation)
mux.HandleFunc("GET /v1/locations/{id}", s.handleGetLocation)
mux.HandleFunc("PATCH /v1/locations/{id}", s.handleUpdateLocation)
mux.HandleFunc("DELETE /v1/locations/{id}", s.handleDeleteLocation)
mux.HandleFunc("GET /v1/items", s.handleListItems)
mux.HandleFunc("POST /v1/items", s.handleCreateItem)
mux.HandleFunc("GET /v1/items/{id}", s.handleGetItem)
```

## Path parameters

```go
func (s *Server) handleGetLocation(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid_id", "Invalid location ID")
        return
    }
    // ...
}
```

## Server struct

```go
type Server struct {
    queries *db.Queries
    mux     *http.ServeMux
}

func NewServer(q *db.Queries) *Server {
    s := &Server{queries: q}
    s.mux = http.NewServeMux()
    s.routes()
    return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    s.mux.ServeHTTP(w, r)
}
```

## JSON helpers

```go
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
    r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1 MB limit
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields()
    return dec.Decode(v)
}
```

## Error response (consistent shape)

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
    writeJSON(w, status, ErrorResponse{Code: code, Message: message})
}
```

## Handler pattern

```go
func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
    var req models.CreateItemRequest
    if err := readJSON(r, &req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid_json", "Invalid request body")
        return
    }

    item, err := s.queries.CreateItem(req)
    if err != nil {
        log.Printf("create item: %v", err)
        writeError(w, http.StatusInternalServerError, "internal_error", "Internal server error")
        return
    }

    w.Header().Set("Location", fmt.Sprintf("/v1/items/%d", item.ID))
    writeJSON(w, http.StatusCreated, item)
}
```

## Middleware

```go
func logRequest(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
    })
}

func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Chain: outermost runs first
func chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
    for i := len(mw) - 1; i >= 0; i-- {
        h = mw[i](h)
    }
    return h
}
```

## Graceful shutdown

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      chain(server, logRequest, cors),
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}

go func() { log.Fatal(srv.ListenAndServe()) }()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

## Service layer

Business logic in `internal/service/`. Services take `*db.Queries` and
orchestrate multi-step operations.

```go
type ItemService struct {
    queries *db.Queries
}

func NewItemService(q *db.Queries) *ItemService {
    return &ItemService{queries: q}
}
```

## Static file serving (frontend)

```go
mux.Handle("GET /", http.FileServer(http.Dir("static")))
```

## TDD checklist

1. Write test in `*_test.go` next to source
2. Run: `mise run test` (or `cd backend && go test ./internal/<pkg>/ -run TestName -v`)
3. See FAIL → write minimum code
4. See PASS → stop, tidy, run `mise run verify`

## Rules

- Use Go 1.22+ method-based routing (`"GET /path/{param}"`)
- Extract path params with `r.PathValue("param")`
- Limit request body with `http.MaxBytesReader`
- Set `Content-Type: application/json` on all JSON responses
- Return `Location` header on 201 Created
- Use consistent `ErrorResponse` shape for all errors
- Inject dependencies via constructors — no global state

## Don'ts

- No external routers (chi, gorilla, echo) — stdlib only
- No assertion libraries — use `t.Fatalf` / `t.Errorf`
- No `interface{}` or `any` unless truly generic
- No `http.Error()` for JSON APIs — use `writeError()`
- No manual method checks — use `"METHOD /path"` pattern
