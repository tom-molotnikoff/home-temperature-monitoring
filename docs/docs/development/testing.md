# Testing

## Running Tests

Run all tests from the `sensor_hub` directory:

```bash
cd sensor_hub
go test ./...
```

Run a specific package:

```bash
go test ./api/
```

Verbose output:

```bash
go test -v ./api/
```

## Test Packages

| Package                  | What It Covers                       |
|--------------------------|--------------------------------------|
| `alerting`               | Alert rule evaluation logic          |
| `api`                    | REST API endpoint handlers           |
| `api/middleware`         | Authentication and authorization     |
| `application_properties` | Configuration file parsing           |
| `db`                     | Database repository layer            |
| `notifications`          | Notification dispatch                |
| `oauth`                  | OAuth2 authentication flow           |
| `periodic`               | Supervised periodic task runner      |
| `service`                | Business logic / service layer       |
| `smtp`                   | Email sending                        |
| `utils`                  | Shared utility functions             |
| `ws`                     | WebSocket connections                |

## Testing Patterns

### Interface-Based Mocks (testify/mock)

Services and repositories are defined as interfaces. Tests create mock
implementations using `testify/mock`:

```go
type mockAlertManagementService struct {
    mock.Mock
}

func (m *mockAlertManagementService) ServiceGetAllAlertRules() ([]alerting.AlertRule, error) {
    args := m.Called()
    return args.Get(0).([]alerting.AlertRule), args.Error(1)
}
```

### Database Tests (go-sqlmock)

Repository tests use `go-sqlmock` to mock the `*sql.DB` connection, verifying
SQL queries and parameters without a real database:

```go
db, mock, _ := sqlmock.New()
mock.ExpectQuery("SELECT .+ FROM alert_rules").
    WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "High temp"))
```

### API Tests (httptest)

API handlers are tested using `httptest` with Gin's test context. Requests go
through the same routing as production:

```go
w := httptest.NewRecorder()
c, router := gin.CreateTestContext(w)
c.Request = httptest.NewRequest("GET", "/api/alerts", nil)
router.ServeHTTP(w, c.Request)
assert.Equal(t, http.StatusOK, w.Code)
```

All API test routes use the `/api` prefix, matching the production routing
configuration.

## Integration Tests

Integration tests exercise the full stack: HTTP API → Gin router → middleware →
services → repositories → real SQLite database, with testcontainers-managed
mock sensor Docker containers.

### Prerequisites

- Docker must be running (testcontainers-go manages containers automatically)

### Running Integration Tests

```bash
cd sensor_hub
go test -tags integration -v -timeout 300s ./integration/
```

Run a specific integration test:

```bash
go test -tags integration -v -run TestCollection_CollectAll -timeout 300s ./integration/
```

### Architecture

Integration tests use the `//go:build integration` build tag and are excluded
from normal `go test ./...` runs.

| Component | Location | Purpose |
|---|---|---|
| `testharness/containers.go` | Container lifecycle | Builds and starts mock sensor Docker containers via testcontainers-go |
| `testharness/harness.go` | Server startup | Creates an in-process sensor-hub server with temp SQLite DB and random port |
| `testharness/client.go` | HTTP client | Authenticated HTTP client with session/CSRF management and typed API methods |
| `integration/main_test.go` | Test entry point | `TestMain` starts containers, server, and admin login for all tests |
| `integration/*_test.go` | Test suites | Tests grouped by API area: auth, sensors, collection, readings, alerts, users, notifications, properties |

### How It Works

1. `TestMain` starts 2 mock sensor Docker containers via testcontainers-go
2. An in-process sensor-hub server starts with a fresh temp SQLite DB
3. Database migrations run automatically
4. An admin user is created and authenticated
5. All integration tests share the same server and database
6. Containers and server are cleaned up after all tests complete

### CI

Integration tests run automatically in the CI pipeline (`.github/workflows/ci.yml`)
on pull requests, main branch pushes, and release tags. They run after unit tests
pass.

## Writing Tests

### Writing a New Unit Test

Unit tests live alongside the code they test. Create a `*_test.go` file in the
same package.

**Service test example** (using testify/mock):

```go
// service/feature_service_test.go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type mockFeatureRepo struct {
    mock.Mock
}

func (m *mockFeatureRepo) GetByName(ctx context.Context, name string) (*types.Feature, error) {
    args := m.Called(ctx, name)
    return args.Get(0).(*types.Feature), args.Error(1)
}

func TestFeatureService_GetByName(t *testing.T) {
    repo := new(mockFeatureRepo)
    svc := NewFeatureService(repo, slog.Default())

    expected := &types.Feature{Name: "test"}
    repo.On("GetByName", mock.Anything, "test").Return(expected, nil)

    result, err := svc.ServiceGetByName(context.Background(), "test")

    assert.NoError(t, err)
    assert.Equal(t, "test", result.Name)
    repo.AssertExpectations(t)
}
```

**Repository test example** (using go-sqlmock):

```go
// db/feature_repository_test.go
package db

func TestGetByName(t *testing.T) {
    db, mock, _ := sqlmock.New()
    defer db.Close()

    repo := NewFeatureRepository(db, slog.Default())

    mock.ExpectQuery("SELECT .+ FROM features WHERE LOWER\\(name\\) = LOWER\\(\\?\\)").
        WithArgs("test").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Test"))

    result, err := repo.GetByName(context.Background(), "test")

    assert.NoError(t, err)
    assert.Equal(t, "Test", result.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

Note: Escape parentheses in go-sqlmock regex patterns — `LOWER\\(name\\)` not
`LOWER(name)`.

### Writing a New Integration Test

Integration tests validate the full HTTP stack with a real database. Add a new
test file in `integration/`:

```go
//go:build integration

package integration

import "testing"

func TestFeature_CreateAndGet(t *testing.T) {
    // Use the package-level client (authenticated admin)
    status, _ := client.CreateFeature(map[string]interface{}{
        "name": "test-feature",
    })
    if status != 201 {
        t.Fatalf("expected 201, got %d", status)
    }

    status, body := client.GetFeature("test-feature")
    if status != 200 {
        t.Fatalf("expected 200, got %d", status)
    }
    // Assert on body...
}
```

Then add the corresponding client methods to `testharness/client.go`:

```go
func (c *Client) CreateFeature(body map[string]interface{}) (int, []byte) {
    return c.post("/api/features/", body)
}

func (c *Client) GetFeature(name string) (int, []byte) {
    return c.get("/api/features/" + url.QueryEscape(name))
}
```

### When to Write Which Type

| Scenario | Test Type |
|----------|-----------|
| Pure business logic, error handling | Unit test with mocks |
| SQL query correctness | Unit test with go-sqlmock |
| Handler request/response mapping | Unit test with httptest |
| Full request flow with real DB | Integration test |
| Case sensitivity, auth, permissions | Integration test |
| Cross-layer data flow | Integration test |
