# Fila Project Rules

Guidelines for AI agents working on this codebase. Follow these patterns to maintain consistency and code quality.

## Documentation & Comments

- **Doc comments**: Always write in Portuguese (exported types, functions, constants)
- **Code comments**: Minimal or none. Code should be self-explanatory.
- **Doc comment format**: Start with entity name, single line preferred, use examples in longer docs
- **Example**: 
  ```go
  // ParseQuery extrai e valida os parâmetros de paginação da query string HTTP.
  func ParseQuery(r *http.Request) Params { ... }
  ```

## Code Organization

### Package Structure
- Organize by domain/feature, not by type (not `models/`, `handlers/`, `services/`)
- Services live in `internal/{domain}/` (e.g., `internal/auth/`, `internal/processos/`)
- Data access layer in `internal/database/`
- Keep related code in same package

### File Naming
- `service.go` - Service implementation
- `hooks.go` - Hook/event implementations
- `config.go` - Configuration structs
- `{name}_test.go` - Tests (same package, not `_test` package)
- `handle_{resource}.go` - HTTP handlers in `cmd/api/`

## Services Pattern

Every service must follow this structure:

```go
type Service struct {
    pool   *pgxpool.Pool
    store  *database.Store
    logger *slog.Logger
    // ... other dependencies
}

func New(...dependencies) *Service {
    return &Service{
        pool:   pool,
        store:  store,
        logger: logger.With(slog.String("service", "domain")),
        // ...
    }
}
```

- Constructor always named `New`
- Always add service name to logger with `logger.With(slog.String("service", "name"))`
- Accept context as first parameter in all methods
- Return error as last return value

## Error Handling

### Define Errors
Package-level error variables at top of file:
```go
var (
    ErrNotFound = errors.New("recurso não encontrado")
    ErrExists   = errors.New("recurso já existe")
)
```

### Use Errors
- Check with `errors.Is(err, ErrNotFound)`
- Wrap with context: `fmt.Errorf("falha ao processar: %w", err)`
- Never use `==` for error comparison
- Translate database errors to domain errors (e.g., constraint violations)

## HTTP Handlers

### Handler Pattern
```go
func (app *application) handleResourceAction(w http.ResponseWriter, r *http.Request) {
    // 1. Get auth if needed: usuario := app.getAuth(r.Context())
    // 2. Decode request: var req struct{}; app.decodeJSON(w, r, &req)
    // 3. Get path params: resourceID := chi.URLParam(r, "resourceID")
    // 4. Call service: result, err := app.service.Method(r.Context(), ...)
    // 5. Handle error: app.serverError(w, r, err)
    // 6. Write response: app.writeJSON(w, http.StatusOK, result)
}
```

### Error Response
Use `app.serverError()`, `app.validationFailed()`, `app.tokenError()` for responses.

### Pagination
```go
params := pagination.ParseQuery(r)
items, total, err := app.service.ListItems(r.Context(), params.Page, params.Limit)
result := pagination.NewResult(items, params.Page, total, params.Limit)
app.writeJSON(w, http.StatusOK, result)
```

## Database

### Store Pattern
All data access through `database.Store`:
```go
type Store struct {
    db DBTX
}

func (s *Store) WithTx(tx pgx.Tx) *Store {
    return &Store{db: tx}
}
```

### Query Pattern
- Use positional parameters: `WHERE id = $1`
- Always scan into proper types: `QueryRow().Scan(&id, &name)`
- Check errors immediately after SQL operations
- Use `sql.Null[T]` for nullable columns

### Transaction Pattern
```go
tx, err := s.pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)

store := s.store.WithTx(tx)
// ... operations using store ...

return tx.Commit(ctx)
```

### List Queries
Use `COUNT(*) OVER()` window function for efficient pagination:
```sql
SELECT col1, col2, COUNT(*) OVER() as total_count
FROM table
WHERE ...
ORDER BY created_at DESC
LIMIT $1 OFFSET $2
```

## Dependency Injection

### Constructor DI
All services created with explicit constructor parameters:
```go
svc := auth.New(pool, store, logger, cache)
```

### Application Structure
Single `application` struct holds all dependencies:
```go
type application struct {
    cfg       *config.Config
    logger    *slog.Logger
    pool      *pgxpool.Pool
    auth      *auth.Service
    processos *processos.Service
    // ...
}
```

All handlers use: `func (app *application) handleX(w, r)`

## Testing

### Test File Structure
```go
func TestFunctionName(t *testing.T) {
    // Setup
    // Action
    // Assert
}
```

### Table-Driven Tests
```go
tests := []struct {
    name     string
    input    Type
    expected Type
}{...}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test code
    })
}
```

### Test Helpers
Use `t.Helper()` and options pattern:
```go
type seedOpt func(*Entity)

func withField(val string) seedOpt {
    return func(e *Entity) { e.Field = val }
}

func seedEntity(t *testing.T, opts ...seedOpt) *Entity {
    t.Helper()
    e := &Entity{default: "value"}
    for _, opt := range opts {
        opt(e)
    }
    return e
}
```

### Assertions
Use `google/go-cmp` for deep equality:
```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Fatalf("mismatch:\n%s", diff)
}
```

## Logging

### Logger Setup
```go
func NewLogger(w io.Writer, dev bool) *slog.Logger {
    if dev {
        return slog.New(tint.NewHandler(w, &tint.Options{Level: slog.LevelDebug}))
    }
    return slog.New(slog.NewJSONHandler(w, nil))
}
```

### Context Logger
```go
ctx = logging.WithLogger(ctx, logger)
logger := logging.FromContext(ctx)
```

### Structured Logging
```go
logger.Info("evento",
    slog.String("campo", "valor"),
    slog.Int("numero", 42),
    slog.Duration("tempo", time.Since(start)),
)
```

## Configuration

### Config Structure
```go
type Config struct {
    BaseURL   string `env:"BASE_URL,notEmpty"`
    RedisURL  string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
    Mail      mail.Config
    Postgres  postgres.Config
}

func NewFromEnv() (*Config, error) {
    var cfg Config
    err := env.Parse(&cfg)
    return &cfg, err
}
```

### Env Tag Rules
- `env:"VAR,notEmpty"` - Required field
- `env:"VAR" envDefault:"value"` - Optional with default
- Nest structs for related config

## Generics & Types

### Generic Types
```go
type Result[T any] struct {
    Data      []T
    Limit     int
    TotalCount int
}

type Parametros[T any] struct {
    Items []T
}
```

### Nullable Columns
Use `sql.Null[T]` instead of pointers:
```go
type User struct {
    Email sql.Null[string]  // Can be NULL
}

// Convert helpers
func Ptr[T any](n sql.Null[T]) *T { ... }
func Null[T any](ptr *T) sql.Null[T] { ... }
```

## Task Queue (River)

### Job Definition
```go
type JobNameArgs struct {
    Param1 string
}

func (args JobNameArgs) Kind() string { return "job:name" }

func (args JobNameArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue: river.QueueDefault,
        UniqueOpts: river.UniqueOpts{
            ByArgs:   true,
            ByPeriod: time.Hour,
        },
    }
}
```

### Worker Implementation
```go
type JobNameWorker struct {
    dependency *Dependency
    river.WorkerDefaults[JobNameArgs]
}

func (w *JobNameWorker) Work(ctx context.Context, job *river.Job[JobNameArgs]) error {
    // implementation
}
```

## Adding New Features

### Checklist
- [ ] Define service in `internal/{domain}/service.go`
- [ ] Add database queries to `internal/database/`
- [ ] Create handlers in `cmd/api/handle_{resource}.go`
- [ ] Register routes in `cmd/api/routes.go`
- [ ] Add unit tests in `*_test.go`
- [ ] Define error types in service package
- [ ] Add middleware if needed in `cmd/api/middleware.go`
- [ ] Document with Portuguese doc comments
- [ ] Update `README.md` if the change affects project structure, prerequisites, or setup

## Updating the README

The `README.md` is a lean developer onboarding document written in Portuguese. When updating it, follow these rules:

- **Language**: Portuguese, consistent with all project documentation
- **Tone**: Direct and concise. No emojis, no filler text
- **Structure**: Keep the existing sections in order:
  1. Title and one-line description
  2. Arquitetura (components overview)
  3. Estrutura do Projeto (directory tree)
  4. Requisitos (prerequisites with links)
  5. Configuracao (env setup)
  6. Executando (run commands)
  7. Comandos Disponiveis (task runner table)
  8. Testes (how to run tests)
- **When to update**:
  - New `internal/` package added or removed: update the directory tree in "Estrutura do Projeto"
  - New prerequisite tool required: add to "Requisitos" with install link
  - New task added to `Taskfile.yml`: add to "Comandos Disponiveis" table
  - New infrastructure dependency (e.g., new external service): add to "Arquitetura" components list
  - Changes to how the app is started or configured: update "Executando" or "Configuracao"
- **When NOT to update**: new handlers, routes, bug fixes, refactors, or internal changes that don't affect project setup or structure
- **Do not** add API endpoint documentation to the README. Route definitions live in `cmd/api/routes.go`

### HTTP Endpoint Pattern
1. Define handler method in `cmd/api`
2. Use `app.decodeJSON()` for request parsing
3. Call service method with context
4. Use `app.writeJSON()` or error helpers for response
5. Register route in `routes.go`

## Middleware

### Common Pattern
```go
func (app *application) middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pre-processing
        ctx := context.WithValue(r.Context(), key, value)
        // Pass to next handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Built-in Middleware
- `authenticate` - Validates bearer token, extracts usuario
- `loadResource` - Loads resource from path param, returns 404 if not found
- `reqLogger` - Logs request method, path, status, duration

## Validation

### Validator Pattern
```go
v := &Validator{FieldErrors: make(map[string]string)}
v.Check(len(name) > 0, "nome", "não pode estar vazio")
v.Check(email != "", "email", "não pode estar vazio")

if !v.Valid() {
    app.validationFailed(w, v)
    return
}
```

## Constants & Boundaries

### Pagination
- `DefaultPage = 1`
- `DefaultLimit = 20`
- `MaxLimit = 100`
- `MinLimit = 1`
- `MinPage = 1`

### HTTP
- `maxBodySize = 1 << 20` (1 MB)
- Always validate request body size

## Common Patterns to Use

| Pattern | Where | Purpose |
|---------|-------|---------|
| Service with hooks | auth, processos | Allow external code to hook into lifecycle events |
| Options pattern | ServiceOpts, SeedOpts | Reduce constructor parameters, flexible configuration |
| Interface satisfaction | `var _ Interface = (*Type)(nil)` | Compile-time verification |
| Context helpers | getAuth, setAuth | Type-safe context value access |
| Store.WithTx | Database operations | Enable transaction support |
| Table-driven tests | All tests | Reduce code duplication |
| Custom errors | Each service | Semantic error checking |

## Technology Stack

- **Web**: chi/v5 router, stdlib http
- **Database**: pgx/v5, pgxpool, PostgreSQL
- **Cache**: Redis with in-memory fallback
- **Queue**: River (PostgreSQL-backed task queue)
- **Logging**: stdlib slog with tint formatting
- **Config**: caarlos0/env v11
- **Testing**: stdlib testing, google/go-cmp, ory/dockertest
- **External**: Azure SDK, SOAP client

## Key Entry Points

| File | Purpose |
|------|---------|
| `cmd/api/main.go` | Application bootstrap |
| `cmd/api/routes.go` | Route definitions |
| `cmd/api/handle_*.go` | HTTP handlers |
| `internal/*/service.go` | Business logic |
| `internal/database/*.go` | Data access |
| `internal/config/config.go` | Configuration |

## References

- Package structure: `internal/{domain}/` layout
- Service creation: Follow `Service` struct pattern with logger injection
- Error handling: Package-level error variables + `errors.Is()`
- Testing: Table-driven tests with helpers
- HTTP: chi router with middleware pattern
- Database: pgx with Store pattern and transaction support
