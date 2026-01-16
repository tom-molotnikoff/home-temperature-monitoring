# Implementation Plan: Comprehensive DB Package Tests

## Summary
Add comprehensive unit tests for all 7 repository implementations in `sensor_hub/db` using `go-sqlmock` for database mocking. Tests will cover happy paths, error cases, edge cases, and concurrent access where relevant.

## Repositories to Test (Priority Order)
1. **SensorRepository** - 13 methods
2. **TemperatureRepository** - 6 methods  
3. **UserRepository** - 17 methods
4. **SessionRepository** - 10 methods
5. **AlertRepository** - 9 methods (replace mock-only tests)
6. **RoleRepository** - 6 methods
7. **FailedLoginRepository** - 5 methods

## Testing Approach
- Use `github.com/DATA-DOG/go-sqlmock` for SQL mocking
- Use `testify/assert` for assertions (matches existing patterns)
- Table-driven tests where appropriate
- Descriptive test names: `TestRepositoryName_MethodName_Scenario`

---

## Task 1: Add sqlmock dependency
**File:** `sensor_hub/go.mod`

Add `github.com/DATA-DOG/go-sqlmock` as a test dependency.

**Verification:** `go mod tidy` succeeds

---

## Task 2: Create test helper utilities
**File:** `sensor_hub/db/test_helpers_test.go`

Create shared test utilities:
- `newMockDB()` - creates sqlmock DB instance
- Common test data factories for Sensor, User, etc.
- Helper functions for common assertions

**Verification:** File compiles with `go build ./db/...`

---

## Task 3: SensorRepository tests
**File:** `sensor_hub/db/sensor_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `SensorExists` | Returns true/false | DB error | Empty name |
| `SetEnabledSensorByName` | Enable/disable | DB error, no rows affected | Already in target state |
| `GetSensorIdByName` | Returns ID | Not found, DB error | |
| `DeleteHealthHistoryOlderThan` | Deletes old records | DB error | No records to delete |
| `GetSensorHealthHistoryById` | Returns history | DB error | Empty result, limit 0 |
| `DeleteSensorByName` | Transaction success | Rollback on error, not found | Cascade delete |
| `GetSensorsByType` | Returns filtered | DB error | No matches, unknown type |
| `UpdateSensorById` | Updates sensor | DB error, not found | No changes |
| `AddSensor` | Inserts sensor | DB error, validation | Empty fields |
| `GetSensorByName` | Returns sensor | Not found, DB error | |
| `GetAllSensors` | Returns all | DB error | Empty table |
| `UpdateSensorHealthById` | Updates health | DB error | Invalid sensor ID |

**Verification:** `go test ./db/... -run TestSensor -v` passes

---

## Task 4: TemperatureRepository tests
**File:** `sensor_hub/db/temperature_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `Add` | Inserts readings | Sensor not found, DB error | Empty slice, batch insert |
| `GetBetweenDates` | Returns range | Invalid table, DB error | Empty range, boundary dates |
| `GetTotalReadingsBySensorId` | Returns count | DB error | Zero readings |
| `GetLatest` | Returns latest per sensor | DB error | Empty table, single sensor |
| `DeleteReadingsOlderThan` | Transaction delete | Rollback on error | Nothing to delete |

**Verification:** `go test ./db/... -run TestTemperature -v` passes

---

## Task 5: UserRepository tests
**File:** `sensor_hub/db/user_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `CreateUser` | Creates with roles | DB error, role not found | No roles |
| `GetUserByUsername` | Returns user+hash | Not found, DB error | Null updated_at |
| `GetUserById` | Returns user | Not found, DB error | |
| `ListUsers` | Returns all | DB error | Empty table |
| `UpdatePassword` | Updates hash | DB error | |
| `SetDisabled` | Sets flag | DB error | |
| `AssignRoleToUser` | Assigns role | Role not found, DB error | Duplicate assign |
| `GetRolesForUser` | Returns roles | DB error | No roles |
| `DeleteSessionsForUser` | Deletes all | DB error | No sessions |
| `DeleteSessionsForUserExcept` | Keeps one | Token not found, DB error | Empty keep token |
| `DeleteUserById` | Transaction delete | Rollback on error | |
| `SetMustChangeFlag` | Updates flag | DB error | |
| `SetRolesForUser` | Replaces roles | Role not found, rollback | Empty roles |

**Verification:** `go test ./db/... -run TestUser -v` passes

---

## Task 6: SessionRepository tests
**File:** `sensor_hub/db/session_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `CreateSession` | Creates with CSRF | DB error | |
| `GetUserIdByToken` | Returns user ID | Not found, DB error | Expired session |
| `GetSessionIdByToken` | Returns session ID | Not found, DB error | Expired session |
| `DeleteSessionByToken` | Deletes session | DB error | |
| `DeleteSessionsForUser` | Deletes all | DB error | |
| `ListSessionsForUser` | Returns sessions | DB error | Empty list |
| `RevokeSessionById` | Revokes session | DB error | |
| `GetCSRFForToken` | Returns CSRF | Not found, DB error | Expired, null CSRF |
| `InsertSessionAudit` | Inserts audit | DB error | Null optional fields |

**Verification:** `go test ./db/... -run TestSession -v` passes

---

## Task 7: AlertRepository tests (replace existing)
**File:** `sensor_hub/db/alert_repository_test.go`

Replace mock-only tests with sqlmock implementation tests.

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `GetAlertRuleBySensorID` | Returns rule | Not found, DB error | Null last_alert, null trigger_status |
| `UpdateLastAlertSent` | No-op (compat) | | |
| `RecordAlertSent` | Inserts history | DB error | |
| `GetAllAlertRules` | Returns all | DB error | Empty, null fields |
| `GetAlertRuleBySensorName` | Returns rule | Not found, DB error | |
| `CreateAlertRule` | Creates rule | DB error | |
| `UpdateAlertRule` | Updates rule | DB error | |
| `DeleteAlertRule` | Deletes rule | DB error | |
| `GetAlertHistory` | Returns history | DB error | Empty, null reading |

**Verification:** `go test ./db/... -run TestAlert -v` passes

---

## Task 8: RoleRepository tests
**File:** `sensor_hub/db/role_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `GetPermissionsForUser` | Returns perms | DB error | No permissions |
| `GetAllRoles` | Returns roles | DB error | Empty |
| `GetAllPermissions` | Returns perms | DB error | Empty |
| `GetPermissionsForRole` | Returns perms | DB error | No permissions |
| `AssignPermissionToRole` | Assigns | DB error | Duplicate |
| `RemovePermissionFromRole` | Removes | DB error | Not assigned |

**Verification:** `go test ./db/... -run TestRole -v` passes

---

## Task 9: FailedLoginRepository tests
**File:** `sensor_hub/db/failed_login_repository_test.go`

### Methods to test:
| Method | Happy Path | Error Cases | Edge Cases |
|--------|------------|-------------|------------|
| `RecordFailedAttempt` | Inserts record | DB error | Null user ID |
| `CountRecentFailedAttemptsByUsername` | Returns count | DB error | Zero count |
| `CountRecentFailedAttemptsByIP` | Returns count | DB error | Zero count |
| `DeleteRecentFailedAttemptsByIP` | Deletes records | DB error | Nothing to delete |
| `DeleteAttemptsOlderThan` | Deletes old | DB error | Nothing to delete |

**Verification:** `go test ./db/... -run TestFailedLogin -v` passes

---

## Task 10: Final verification
Run all DB tests together and ensure no regressions.

**Verification:** 
- `go test ./db/... -v` passes
- `go test ./... -v` passes (full suite)

---

## Implementation Notes

### Test Pattern Template
```go
func TestRepositoryName_MethodName_Scenario(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    repo := NewRepository(db)

    // Setup expectations
    mock.ExpectQuery(...).WillReturnRows(...)

    // Execute
    result, err := repo.Method(...)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Edge Cases to Consider
- SQL injection attempts (parameterized queries should handle)
- NULL values in optional columns
- Empty result sets
- Transaction rollback scenarios
- Concurrent access (where applicable)
- Boundary values (0, negative, max int)

---

## Estimated Effort
- Task 1: 5 minutes
- Task 2: 15 minutes
- Task 3: 45 minutes (largest repository)
- Task 4: 30 minutes
- Task 5: 45 minutes (many methods)
- Task 6: 30 minutes
- Task 7: 30 minutes
- Task 8: 20 minutes
- Task 9: 15 minutes
- Task 10: 10 minutes

**Total: ~4 hours**
