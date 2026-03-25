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
