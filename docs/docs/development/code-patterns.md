# Code Patterns

This guide describes the conventions and patterns used across the codebase. Use
it as a reference when adding new features or making changes.

## Adding a New Feature (Full Stack)

A complete new feature typically touches these layers, in this order:

| Step | Layer | Location | What to create |
|------|-------|----------|---------------|
| 1 | Database | `db/migrations/` | New migration file (if schema changes needed) |
| 2 | Types | `types/` | Go structs for the new data |
| 3 | Repository | `db/` | Interface + implementation for data access |
| 4 | Service | `service/` | Business logic, WebSocket broadcasts |
| 5 | Handler | `api/` | HTTP handler functions + route registration |
| 6 | OpenAPI | `openapi.yaml` | Document the new endpoints |
| 7 | CLI command | `cmd/` | Cobra command for the CLI client |
| 8 | UI API client | `ui/sensor_hub_ui/src/api/` | TypeScript API client module |
| 9 | UI components | `ui/sensor_hub_ui/src/` | Pages, hooks, components |
| 10 | Integration tests | `integration/` | End-to-end test + test client methods |

## Handler Pattern

Each resource area has two files in `api/`:

**Handler file** (`feature_api.go`):

```go
// Package-level service variable
var featureService service.FeatureServiceInterface

// Initialisation function called from serve.go
func InitFeatureAPI(s service.FeatureServiceInterface) {
    featureService = s
}

// Handler functions named verbNounHandler
func getFeatureByNameHandler(c *gin.Context) {
    name := c.Param("name")
    ctx := c.Request.Context()

    result, err := featureService.ServiceGetByName(ctx, name)
    if err != nil {
        slog.Error("error fetching feature", "name", name, "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
        return
    }
    if result == nil {
        c.JSON(http.StatusNotFound, gin.H{"message": "Feature not found"})
        return
    }

    c.IndentedJSON(http.StatusOK, result)
}
```

**Routes file** (`feature_routes.go`):

```go
func RegisterFeatureRoutes(router gin.IRouter) {
    group := router.Group("/features")
    {
        group.GET("/:name", middleware.AuthRequired(),
            middleware.RequirePermission("view_features"), getFeatureByNameHandler)
        group.POST("/", middleware.AuthRequired(),
            middleware.RequirePermission("manage_features"), createFeatureHandler)
    }
}
```

Routes are registered in `api.go:InitialiseAndListen()`. Middleware is applied
per-route, not globally. The standard middleware stack for a protected route is
`AuthRequired()` → `RequirePermission(perm)`.

### HTTP Status Codes

| Status | When to use |
|--------|-------------|
| 200 OK | Successful GET |
| 201 Created | Successful POST that creates a resource |
| 202 Accepted | Successful PATCH/PUT |
| 400 Bad Request | Invalid input, missing parameters, validation failure |
| 401 Unauthorized | No auth or invalid auth |
| 403 Forbidden | Authenticated but lacking permission |
| 404 Not Found | Resource does not exist |
| 500 Internal Server Error | Database or service layer error |

### Error Response Format

```json
{"message": "Human-readable error description"}
```

Some endpoints include an `"error"` field with the underlying error string.

## Service Pattern

```go
type FeatureService struct {
    repo   database.FeatureRepositoryInterface
    logger *slog.Logger
}

func NewFeatureService(
    repo   database.FeatureRepositoryInterface,
    logger *slog.Logger,
) *FeatureService {
    return &FeatureService{
        repo:   repo,
        logger: logger.With("component", "feature_service"),
    }
}

// Public method naming: ServiceVerbNoun
func (s *FeatureService) ServiceGetByName(ctx context.Context, name string) (*types.Feature, error) {
    feature, err := s.repo.GetByName(ctx, name)
    if err != nil {
        return nil, fmt.Errorf("error fetching feature by name: %w", err)
    }
    return feature, nil
}
```

Key conventions:

- Constructor receives all dependencies. Logger is always the last parameter
- Logger is wrapped with `.With("component", "service_name")`
- Public methods are prefixed with `Service` (e.g. `ServiceGetByName`)
- All methods take `ctx context.Context` as the first parameter
- Errors are wrapped with `fmt.Errorf("context: %w", err)`
- WebSocket broadcasts use `ws.BroadcastToTopic(topic, data)` after state changes

## Repository Pattern

```go
// Interface (db/feature_repository_interface.go)
type FeatureRepositoryInterface interface {
    GetByName(ctx context.Context, name string) (*types.Feature, error)
    Add(ctx context.Context, feature types.Feature) error
}

// Implementation (db/feature_repository.go)
type FeatureRepository struct {
    db     *sql.DB
    logger *slog.Logger
}

func NewFeatureRepository(db *sql.DB, logger *slog.Logger) *FeatureRepository {
    return &FeatureRepository{
        db:     db,
        logger: logger.With("component", "feature_repository"),
    }
}

func (r *FeatureRepository) GetByName(ctx context.Context, name string) (*types.Feature, error) {
    row := r.db.QueryRowContext(ctx,
        "SELECT id, name, value FROM features WHERE LOWER(name) = LOWER(?)", name)

    var f types.Feature
    err := row.Scan(&f.Id, &f.Name, &f.Value)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("error querying feature: %w", err)
    }
    return &f, nil
}
```

Key conventions:

- All string-equality WHERE clauses use `LOWER(col) = LOWER(?)`
- `sql.ErrNoRows` returns `nil, nil` (not found is not an error)
- Use `SQLiteTime` / `NullSQLiteTime` from `db/sqlite_time.go` for time columns
- Logging is DEBUG for successful operations
- Method naming: `VerbNoun` (e.g. `GetByName`, `Add`, `DeleteByName`)

## CLI Command Pattern

```go
// cmd/feature.go
var featureCmd = &cobra.Command{
    Use:   "features",
    Short: "Manage features",
}

var featureListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all features",
    RunE: func(cmd *cobra.Command, args []string) error {
        serverURL, apiKey, insecure, err := loadClientConfig(cmd)
        if err != nil {
            return err
        }
        client := NewClient(serverURL, apiKey, insecure)

        data, err := client.Get("/api/features/", nil)
        if err != nil {
            return err
        }
        printJSON(data)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(featureCmd)
    featureCmd.AddCommand(featureListCmd)
}
```

The CLI uses the same REST API as the frontend. Authentication is via API key
(`X-API-Key` header) — no CSRF tokens needed. Commands call `loadClientConfig`
to read the server URL, API key, and TLS settings from flags or the config file.

## UI API Client Pattern

Each API area has a TypeScript module in `ui/sensor_hub_ui/src/api/`:

```typescript
// api/Features.ts
import { get, post, del } from './Client';

export const getAll = () => get<Feature[]>('/features/');
export const getByName = (name: string) => get<Feature>(`/features/${name}`);
export const create = (feature: Feature) => post<void>('/features/', feature);
export const remove = (name: string) => del<void>(`/features/${name}`);
```

The core client (`Client.ts`) handles:

- `credentials: 'include'` for session cookies
- `X-CSRF-Token` header from the in-memory CSRF token store
- `X-Requested-With: XMLHttpRequest` header
- JSON serialisation and typed responses

## Integration Test Client Pattern

Add methods to `testharness/client.go` for each new endpoint:

```go
func (c *Client) GetFeatures() (int, []byte) {
    return c.get("/api/features/")
}

func (c *Client) CreateFeature(body map[string]interface{}) (int, []byte) {
    return c.post("/api/features/", body)
}
```

Then write tests in `integration/feature_test.go` using the `//go:build integration`
build tag.

## Error Handling

Errors propagate upward through the layers:

```
Repository: return fmt.Errorf("error querying feature: %w", err)
      ↓
Service:    return nil, fmt.Errorf("error fetching feature: %w", err)
      ↓
Handler:    slog.Error("...", "error", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
```

Custom error types are used for specific conditions:

```go
// service/sensor_service.go
type AlreadyExistsError struct{ Message string }

// Checked with errors.As in callers:
var alreadyExistsErr *AlreadyExistsError
if errors.As(err, &alreadyExistsErr) {
    // handle specifically
}
```

## Logging Conventions

The application uses Go's `log/slog` for structured logging.

### Creating Loggers

Services and repositories receive `*slog.Logger` as the last constructor
parameter and create a child logger:

```go
logger: logger.With("component", "sensor_service")
```

### Log Levels

| Level | When to use | Example |
|-------|-------------|---------|
| DEBUG | Detailed operation tracking | `s.logger.Debug("saved reading", "sensor", name)` |
| INFO | Significant events, milestones | `s.logger.Info("sensor added", "name", name)` |
| WARN | Unexpected but non-fatal | `s.logger.Warn("skipping disabled sensor", "name", name)` |
| ERROR | Operation failures | `s.logger.Error("error fetching sensor", "error", err)` |

### Patterns by Layer

- **Handlers**: use `slog` package-level functions (no component context)
- **Services**: use the injected `s.logger` (includes component context)
- **Repositories**: use the injected `r.logger` (includes component context)
- **Periodic tasks**: use the logger from `TaskConfig` (includes task name)

Always include relevant context as structured fields:

```go
s.logger.Error("error fetching temperature from sensor",
    "name", sensor.Name, "url", sensor.URL, "error", err)
```

## Dependency Injection

Dependencies are injected via constructors. There is no DI container — wiring
happens in `cmd/serve.go`. The initialisation order is:

```
repositories (depend on *sql.DB)
    → services (depend on repositories and other services)
        → API handlers (depend on services)
            → middleware (depends on auth service and role repository)
```

This is a simple, explicit pattern. Adding a new dependency means updating the
constructor, the call site in `serve.go`, and (if testing) the mock setup.

## Naming Conventions

| Item | Convention | Example |
|------|-----------|---------|
| Handler function | `verbNounHandler` | `addSensorHandler` |
| Service method | `ServiceVerbNoun` | `ServiceGetSensorByName` |
| Repository method | `VerbNoun` | `GetSensorByName` |
| Service interface | `XxxServiceInterface` | `SensorServiceInterface` |
| Repository interface | `XxxRepositoryInterface` or `XxxRepository` | `SensorRepositoryInterface[T]` |
| Constructor | `NewXxxService` / `NewXxxRepository` | `NewSensorService` |
| Route group | resource plural | `/sensors`, `/alerts`, `/users` |
| API client (TS) | module per resource | `Sensors.ts`, `Alerts.ts` |
| CLI command | resource plural, subcommand verb | `sensors list`, `alerts create` |
| Test file (integration) | `feature_test.go` | `sensor_crud_test.go` |
