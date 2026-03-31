# Customisable Widget Dashboard — Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Replace fixed MUI Grid layouts with a drag-and-drop widget dashboard powered by react-grid-layout, persisted via a new backend CRUD API with RBAC integration.

**Architecture:** New `dashboards` table stores per-user dashboard configurations as JSON. A Go service + API layer provides CRUD + sharing + default-setting endpoints. The React frontend introduces a widget registry mapping type strings to components, a `DashboardEngine` that renders widgets via react-grid-layout, and an edit mode for drag/resize. Existing card components are wrapped as widgets with zero behaviour change; 9 new widget types are added incrementally.

**Tech Stack:** Go (Gin, golang-migrate, SQLite), React 19 (react-grid-layout, MUI 7, TypeScript, Vite), Cobra CLI

**Issue:** [#9 — Customisable Widget Dashboard](https://github.com/tom-molotnikoff/home-temperature-monitoring/issues/9)

---

## Phase Overview

| Phase | Description | Tasks |
|-------|-------------|-------|
| 1 | **Backend Foundation** — DB migration, types, repository, service, API, tests | 1–6 |
| 2 | **Frontend Infrastructure** — react-grid-layout, widget registry, dashboard API client, dashboard engine | 7–11 |
| 3 | **Existing Component Widget Adapters** — wrap 8 existing cards as widgets | 12–13 |
| 4 | **Dashboard Pages & UX** — edit mode, widget picker, save/load, responsive, routing | 14–17 |
| 5 | **New Widgets** — 9 new widget types | 18–26 |
| 6 | **CLI & LLM Skills** — dashboard CLI commands, skill file updates | 27–28 |
| 7 | **Sharing & Default Dashboard** — share endpoint, set-default endpoint | 29–30 |
| 8 | **Documentation & OpenAPI** — developer docs, API docs, OpenAPI spec | 31–32 |

---

## Task List (Outline)

### Phase 1: Backend Foundation
- **Task 1:** Database migration — `dashboards` table + new permissions
- **Task 2:** Go types — Dashboard, Widget, WidgetLayout structs
- **Task 3:** Repository — `DashboardRepository` interface + SQLite implementation
- **Task 4:** Repository tests — go-sqlmock unit tests
- **Task 5:** Service — `DashboardService` with business logic
- **Task 6:** API handlers + routes — CRUD endpoints with auth/permissions

### Phase 2: Frontend Infrastructure
- **Task 7:** Install react-grid-layout + TypeScript types
- **Task 8:** Dashboard TypeScript types + API client module
- **Task 9:** Widget registry — type → component mapping
- **Task 10:** `DashboardEngine` component — renders widgets via react-grid-layout
- **Task 11:** `DashboardProvider` context — load/save dashboard state

### Phase 3: Existing Component Widget Adapters
- **Task 12:** Widget wrapper HOC + adapter pattern
- **Task 13:** Wrap all 8 existing components as widgets

### Phase 4: Dashboard Pages & UX
- **Task 14:** Edit mode toggle — lock/unlock layout
- **Task 15:** Widget picker dialog — add widgets from catalog
- **Task 16:** Dashboard page — replaces TemperatureDashboard, with save/load/create/delete
- **Task 17:** Routing + navigation — dashboard list, individual dashboards

### Phase 5: New Widgets
- **Task 18:** `current-reading` — big number display with trend arrow
- **Task 19:** `gauge` — visual gauge with colour zones
- **Task 20:** `group-summary` — room/group average reading
- **Task 21:** `alert-summary` — active alerts compact list
- **Task 22:** `min-max-avg` — period statistics card
- **Task 23:** `heatmap-calendar` — daily heatmap calendar
- **Task 24:** `sensor-uptime` — uptime indicator
- **Task 25:** `multi-sensor-comparison` — overlay chart
- **Task 26:** `markdown-note` — user-defined text/markdown block

### Phase 6: CLI & LLM Skills
- **Task 27:** CLI — `sensor-hub dashboards` commands (list, create, get, update, delete, share, set-default)
- **Task 28:** LLM skill files — document dashboard JSON schema for AI generation

### Phase 7: Sharing & Default Dashboard
- **Task 29:** Share endpoint — `POST /api/dashboards/:id/share`
- **Task 30:** Default dashboard — `PUT /api/dashboards/:id/default`

### Phase 8: Documentation & OpenAPI
- **Task 31:** OpenAPI spec — document all dashboard endpoints
- **Task 32:** Developer docs — architecture, widget development guide

---

## Detailed Task Steps

### Task 1: Database Migration — `dashboards` table + new permissions

**Files:**
- Create: `sensor_hub/db/migrations/000004_dashboards.up.sql`
- Create: `sensor_hub/db/migrations/000004_dashboards.down.sql`

**Step 1: Write the UP migration**

```sql
-- Dashboard storage
CREATE TABLE IF NOT EXISTS dashboards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    config TEXT NOT NULL DEFAULT '{"widgets":[],"breakpoints":{"lg":12,"md":8,"sm":4}}',
    shared INTEGER NOT NULL DEFAULT 0,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_dashboards_user_id ON dashboards(user_id);

-- New permissions
INSERT OR IGNORE INTO permissions (name, description) VALUES
    ('manage_dashboards', 'Create, edit, delete, and share dashboards'),
    ('view_dashboards', 'View dashboards');

-- Grant to admin (all permissions)
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('manage_dashboards', 'view_dashboards');

-- Grant view_dashboards to user and viewer roles
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('manage_dashboards', 'view_dashboards');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'viewer' AND p.name = 'view_dashboards';
```

**Step 2: Write the DOWN migration**

```sql
-- Remove role_permissions for dashboard permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN ('manage_dashboards', 'view_dashboards')
);

-- Remove permissions
DELETE FROM permissions WHERE name IN ('manage_dashboards', 'view_dashboards');

-- Drop table
DROP TABLE IF EXISTS dashboards;
```

**Step 3: Verify migration applies cleanly**

Run: `cd sensor_hub && go test ./db/ -v -run TestInit`

If no specific migration test exists, verify the binary starts:
```bash
cd sensor_hub && go build -o /tmp/sensor-hub-test . && echo "BUILD OK"
```

---

### Task 2: Go Types — Dashboard structs

**Files:**
- Create: `sensor_hub/types/dashboard.go`

**Step 1: Create the Dashboard type definitions**

```go
package types

import "time"

// Dashboard represents a user's saved dashboard configuration.
type Dashboard struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Name      string    `json:"name"`
	Config    string    `json:"config"`
	Shared    bool      `json:"shared"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DashboardConfig is the JSON structure stored in the config column.
type DashboardConfig struct {
	Widgets     []DashboardWidget    `json:"widgets"`
	Breakpoints DashboardBreakpoints `json:"breakpoints"`
}

// DashboardWidget represents a single widget on the dashboard.
type DashboardWidget struct {
	Id     string                 `json:"id"`
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
	Layout WidgetLayout           `json:"layout"`
}

// WidgetLayout defines the position and size of a widget on the grid.
type WidgetLayout struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// DashboardBreakpoints defines the number of columns at each responsive breakpoint.
type DashboardBreakpoints struct {
	Lg int `json:"lg"`
	Md int `json:"md"`
	Sm int `json:"sm"`
}

// CreateDashboardRequest is the request body for creating a dashboard.
type CreateDashboardRequest struct {
	Name   string          `json:"name" binding:"required"`
	Config DashboardConfig `json:"config"`
}

// UpdateDashboardRequest is the request body for updating a dashboard.
type UpdateDashboardRequest struct {
	Name   string          `json:"name"`
	Config DashboardConfig `json:"config"`
}

// ShareDashboardRequest is the request body for sharing a dashboard.
type ShareDashboardRequest struct {
	TargetUserId int `json:"target_user_id" binding:"required"`
}
```

**Step 2: Verify it compiles**

Run: `cd sensor_hub && go build ./types/`

---

### Task 3: Repository — `DashboardRepository` interface + implementation

**Files:**
- Create: `sensor_hub/db/dashboard_repository_interface.go`
- Create: `sensor_hub/db/dashboard_repository.go`

**Step 1: Create the repository interface**

```go
package db

import (
	"context"
	"example/sensorHub/types"
)

type DashboardRepository interface {
	Create(ctx context.Context, dashboard *types.Dashboard) (int, error)
	GetById(ctx context.Context, id int) (*types.Dashboard, error)
	GetByUserId(ctx context.Context, userId int) ([]types.Dashboard, error)
	Update(ctx context.Context, dashboard *types.Dashboard) error
	Delete(ctx context.Context, id int) error
	SetDefault(ctx context.Context, userId int, dashboardId int) error
	GetDefaultForUser(ctx context.Context, userId int) (*types.Dashboard, error)
	CreateCopy(ctx context.Context, sourceDashboardId int, targetUserId int) (int, error)
}
```

**Step 2: Create the repository implementation**

```go
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"example/sensorHub/types"
)

type SqlDashboardRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewDashboardRepository(db *sql.DB, logger *slog.Logger) *SqlDashboardRepository {
	return &SqlDashboardRepository{
		db:     db,
		logger: logger.With("component", "dashboard_repository"),
	}
}

func (r *SqlDashboardRepository) Create(ctx context.Context, dashboard *types.Dashboard) (int, error) {
	query := `INSERT INTO dashboards (user_id, name, config, shared, is_default) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, dashboard.UserId, dashboard.Name, dashboard.Config, dashboard.Shared, dashboard.IsDefault)
	if err != nil {
		return 0, fmt.Errorf("error creating dashboard: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting last insert id: %w", err)
	}
	r.logger.Debug("created dashboard", "id", id, "user_id", dashboard.UserId, "name", dashboard.Name)
	return int(id), nil
}

func (r *SqlDashboardRepository) GetById(ctx context.Context, id int) (*types.Dashboard, error) {
	query := `SELECT id, user_id, name, config, shared, is_default, created_at, updated_at FROM dashboards WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	d := &types.Dashboard{}
	var createdAt, updatedAt SQLiteTime
	err := row.Scan(&d.Id, &d.UserId, &d.Name, &d.Config, &d.Shared, &d.IsDefault, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying dashboard by id: %w", err)
	}
	d.CreatedAt = createdAt.Time
	d.UpdatedAt = updatedAt.Time
	r.logger.Debug("fetched dashboard", "id", id)
	return d, nil
}

func (r *SqlDashboardRepository) GetByUserId(ctx context.Context, userId int) ([]types.Dashboard, error) {
	query := `SELECT id, user_id, name, config, shared, is_default, created_at, updated_at
		FROM dashboards WHERE user_id = ? OR shared = 1 ORDER BY is_default DESC, name ASC`
	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying dashboards for user: %w", err)
	}
	defer rows.Close()

	var dashboards []types.Dashboard
	for rows.Next() {
		var d types.Dashboard
		var createdAt, updatedAt SQLiteTime
		if err := rows.Scan(&d.Id, &d.UserId, &d.Name, &d.Config, &d.Shared, &d.IsDefault, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("error scanning dashboard row: %w", err)
		}
		d.CreatedAt = createdAt.Time
		d.UpdatedAt = updatedAt.Time
		dashboards = append(dashboards, d)
	}
	r.logger.Debug("fetched dashboards for user", "user_id", userId, "count", len(dashboards))
	return dashboards, nil
}

func (r *SqlDashboardRepository) Update(ctx context.Context, dashboard *types.Dashboard) error {
	query := `UPDATE dashboards SET name = ?, config = ?, shared = ?, updated_at = datetime('now') WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, dashboard.Name, dashboard.Config, dashboard.Shared, dashboard.Id)
	if err != nil {
		return fmt.Errorf("error updating dashboard: %w", err)
	}
	r.logger.Debug("updated dashboard", "id", dashboard.Id)
	return nil
}

func (r *SqlDashboardRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM dashboards WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting dashboard: %w", err)
	}
	r.logger.Debug("deleted dashboard", "id", id)
	return nil
}

func (r *SqlDashboardRepository) SetDefault(ctx context.Context, userId int, dashboardId int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Clear existing default for this user
	_, err = tx.ExecContext(ctx, `UPDATE dashboards SET is_default = 0 WHERE user_id = ?`, userId)
	if err != nil {
		return fmt.Errorf("error clearing default dashboard: %w", err)
	}

	// Set new default
	_, err = tx.ExecContext(ctx, `UPDATE dashboards SET is_default = 1, updated_at = datetime('now') WHERE id = ? AND user_id = ?`, dashboardId, userId)
	if err != nil {
		return fmt.Errorf("error setting default dashboard: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing set-default transaction: %w", err)
	}
	r.logger.Debug("set default dashboard", "user_id", userId, "dashboard_id", dashboardId)
	return nil
}

func (r *SqlDashboardRepository) GetDefaultForUser(ctx context.Context, userId int) (*types.Dashboard, error) {
	query := `SELECT id, user_id, name, config, shared, is_default, created_at, updated_at FROM dashboards WHERE user_id = ? AND is_default = 1`
	row := r.db.QueryRowContext(ctx, query, userId)
	d := &types.Dashboard{}
	var createdAt, updatedAt SQLiteTime
	err := row.Scan(&d.Id, &d.UserId, &d.Name, &d.Config, &d.Shared, &d.IsDefault, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error querying default dashboard: %w", err)
	}
	d.CreatedAt = createdAt.Time
	d.UpdatedAt = updatedAt.Time
	return d, nil
}

func (r *SqlDashboardRepository) CreateCopy(ctx context.Context, sourceDashboardId int, targetUserId int) (int, error) {
	source, err := r.GetById(ctx, sourceDashboardId)
	if err != nil {
		return 0, fmt.Errorf("error fetching source dashboard: %w", err)
	}
	if source == nil {
		return 0, fmt.Errorf("source dashboard not found")
	}

	copy := &types.Dashboard{
		UserId: targetUserId,
		Name:   source.Name + " (shared)",
		Config: source.Config,
		Shared: false,
	}
	return r.Create(ctx, copy)
}
```

**Step 3: Verify it compiles**

Run: `cd sensor_hub && go build ./db/`

---

### Task 4: Repository Tests — go-sqlmock unit tests

**Files:**
- Create: `sensor_hub/db/dashboard_repository_test.go`

**Step 1: Write tests for all repository methods**

```go
package db

import (
	"context"
	"log/slog"
	"testing"

	"example/sensorHub/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDashboardRepository_Create(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	mock.ExpectExec("INSERT INTO dashboards").
		WithArgs(1, "My Dashboard", `{"widgets":[]}`, false, false).
		WillReturnResult(sqlmock.NewResult(42, 1))

	dashboard := &types.Dashboard{UserId: 1, Name: "My Dashboard", Config: `{"widgets":[]}`, Shared: false, IsDefault: false}
	id, err := repo.Create(context.Background(), dashboard)

	assert.NoError(t, err)
	assert.Equal(t, 42, id)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetById(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "config", "shared", "is_default", "created_at", "updated_at"}).
		AddRow(1, 1, "Test", `{"widgets":[]}`, false, true, "2026-03-31 00:00:00", "2026-03-31 00:00:00")
	mock.ExpectQuery("SELECT .+ FROM dashboards WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(rows)

	d, err := repo.GetById(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, d)
	assert.Equal(t, "Test", d.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetById_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	mock.ExpectQuery("SELECT .+ FROM dashboards WHERE id = \\?").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "config", "shared", "is_default", "created_at", "updated_at"}))

	d, err := repo.GetById(context.Background(), 999)

	assert.NoError(t, err)
	assert.Nil(t, d)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetByUserId(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	rows := sqlmock.NewRows([]string{"id", "user_id", "name", "config", "shared", "is_default", "created_at", "updated_at"}).
		AddRow(1, 1, "Default", `{"widgets":[]}`, false, true, "2026-03-31 00:00:00", "2026-03-31 00:00:00").
		AddRow(2, 1, "Custom", `{"widgets":[]}`, false, false, "2026-03-31 00:00:00", "2026-03-31 00:00:00")
	mock.ExpectQuery("SELECT .+ FROM dashboards WHERE user_id = \\? OR shared = 1").
		WithArgs(1).
		WillReturnRows(rows)

	dashboards, err := repo.GetByUserId(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, dashboards, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	mock.ExpectExec("UPDATE dashboards SET").
		WithArgs("Updated", `{"widgets":[]}`, false, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), &types.Dashboard{Id: 1, Name: "Updated", Config: `{"widgets":[]}`, Shared: false})

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	mock.ExpectExec("DELETE FROM dashboards WHERE id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_SetDefault(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewDashboardRepository(db, slog.Default())

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE dashboards SET is_default = 0 WHERE user_id = \\?").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE dashboards SET is_default = 1").
		WithArgs(5, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.SetDefault(context.Background(), 1, 5)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
```

**Step 2: Run the tests**

Run: `cd sensor_hub && go test -v ./db/ -run TestDashboard`

Expected: All 6 tests PASS.

---

### Task 5: Service — `DashboardService` with business logic

**Files:**
- Create: `sensor_hub/service/dashboard_service_interface.go`
- Create: `sensor_hub/service/dashboard_service.go`

**Step 1: Create the service interface**

```go
package service

import (
	"context"
	"example/sensorHub/types"
)

type DashboardServiceInterface interface {
	ServiceListDashboards(ctx context.Context, userId int) ([]types.Dashboard, error)
	ServiceGetDashboard(ctx context.Context, id int) (*types.Dashboard, error)
	ServiceGetDefaultDashboard(ctx context.Context, userId int) (*types.Dashboard, error)
	ServiceCreateDashboard(ctx context.Context, userId int, req types.CreateDashboardRequest) (int, error)
	ServiceUpdateDashboard(ctx context.Context, userId int, id int, req types.UpdateDashboardRequest) error
	ServiceDeleteDashboard(ctx context.Context, userId int, id int) error
	ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error
	ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error
}
```

**Step 2: Create the service implementation**

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"example/sensorHub/db"
	"example/sensorHub/types"
)

type DashboardService struct {
	repo   db.DashboardRepository
	logger *slog.Logger
}

func NewDashboardService(repo db.DashboardRepository, logger *slog.Logger) *DashboardService {
	return &DashboardService{
		repo:   repo,
		logger: logger.With("component", "dashboard_service"),
	}
}

func (s *DashboardService) ServiceListDashboards(ctx context.Context, userId int) ([]types.Dashboard, error) {
	dashboards, err := s.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error listing dashboards: %w", err)
	}
	return dashboards, nil
}

func (s *DashboardService) ServiceGetDashboard(ctx context.Context, id int) (*types.Dashboard, error) {
	dashboard, err := s.repo.GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting dashboard: %w", err)
	}
	return dashboard, nil
}

func (s *DashboardService) ServiceGetDefaultDashboard(ctx context.Context, userId int) (*types.Dashboard, error) {
	dashboard, err := s.repo.GetDefaultForUser(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting default dashboard: %w", err)
	}
	return dashboard, nil
}

func (s *DashboardService) ServiceCreateDashboard(ctx context.Context, userId int, req types.CreateDashboardRequest) (int, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return 0, fmt.Errorf("error marshalling dashboard config: %w", err)
	}

	dashboard := &types.Dashboard{
		UserId: userId,
		Name:   req.Name,
		Config: string(configJSON),
	}

	id, err := s.repo.Create(ctx, dashboard)
	if err != nil {
		return 0, fmt.Errorf("error creating dashboard: %w", err)
	}
	s.logger.Info("dashboard created", "id", id, "user_id", userId, "name", req.Name)
	return id, nil
}

func (s *DashboardService) ServiceUpdateDashboard(ctx context.Context, userId int, id int, req types.UpdateDashboardRequest) error {
	existing, err := s.repo.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for update: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to update this dashboard")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}

	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return fmt.Errorf("error marshalling dashboard config: %w", err)
	}
	existing.Config = string(configJSON)

	if err := s.repo.Update(ctx, existing); err != nil {
		return fmt.Errorf("error updating dashboard: %w", err)
	}
	s.logger.Info("dashboard updated", "id", id, "user_id", userId)
	return nil
}

func (s *DashboardService) ServiceDeleteDashboard(ctx context.Context, userId int, id int) error {
	existing, err := s.repo.GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to delete this dashboard")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error deleting dashboard: %w", err)
	}
	s.logger.Info("dashboard deleted", "id", id, "user_id", userId)
	return nil
}

func (s *DashboardService) ServiceShareDashboard(ctx context.Context, userId int, dashboardId int, targetUserId int) error {
	existing, err := s.repo.GetById(ctx, dashboardId)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for share: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to share this dashboard")
	}

	_, err = s.repo.CreateCopy(ctx, dashboardId, targetUserId)
	if err != nil {
		return fmt.Errorf("error sharing dashboard: %w", err)
	}
	s.logger.Info("dashboard shared", "id", dashboardId, "from_user", userId, "to_user", targetUserId)
	return nil
}

func (s *DashboardService) ServiceSetDefaultDashboard(ctx context.Context, userId int, dashboardId int) error {
	existing, err := s.repo.GetById(ctx, dashboardId)
	if err != nil {
		return fmt.Errorf("error fetching dashboard for set-default: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("dashboard not found")
	}
	if existing.UserId != userId {
		return fmt.Errorf("not authorized to set this dashboard as default")
	}

	if err := s.repo.SetDefault(ctx, userId, dashboardId); err != nil {
		return fmt.Errorf("error setting default dashboard: %w", err)
	}
	s.logger.Info("default dashboard set", "dashboard_id", dashboardId, "user_id", userId)
	return nil
}
```

**Step 3: Verify it compiles**

Run: `cd sensor_hub && go build ./service/`

---

### Task 6: API Handlers + Routes — CRUD endpoints with auth/permissions

**Files:**
- Create: `sensor_hub/api/dashboard_api.go`
- Create: `sensor_hub/api/dashboard_routes.go`
- Modify: `sensor_hub/api/api.go` — register dashboard routes
- Modify: `sensor_hub/cmd/serve.go` — wire DashboardService + InitDashboardAPI

**Step 1: Create the API handler**

```go
package api

import (
	"net/http"
	"strconv"

	"example/sensorHub/service"
	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
)

var dashboardService service.DashboardServiceInterface

func InitDashboardAPI(s service.DashboardServiceInterface) {
	dashboardService = s
}

func listDashboardsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)

	dashboards, err := dashboardService.ServiceListDashboards(ctx, user.Id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing dashboards"})
		return
	}
	if dashboards == nil {
		dashboards = []types.Dashboard{}
	}
	c.IndentedJSON(http.StatusOK, dashboards)
}

func getDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	dashboard, err := dashboardService.ServiceGetDashboard(ctx, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error getting dashboard"})
		return
	}
	if dashboard == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Dashboard not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, dashboard)
}

func createDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)

	var req types.CreateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	id, err := dashboardService.ServiceCreateDashboard(ctx, user.Id, req)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error creating dashboard"})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func updateDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req types.UpdateDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := dashboardService.ServiceUpdateDashboard(ctx, user.Id, id, req); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard updated"})
}

func deleteDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	if err := dashboardService.ServiceDeleteDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard deleted"})
}

func shareDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	var req types.ShareDashboardRequest
	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := dashboardService.ServiceShareDashboard(ctx, user.Id, id, req.TargetUserId); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Dashboard shared"})
}

func setDefaultDashboardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	user := c.MustGet("currentUser").(*types.User)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid dashboard ID"})
		return
	}

	if err := dashboardService.ServiceSetDefaultDashboard(ctx, user.Id, id); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Default dashboard set"})
}
```

**Step 2: Create the routes file**

```go
package api

import (
	"example/sensorHub/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(router gin.IRouter) {
	group := router.Group("/dashboards")
	{
		group.GET("/",
			middleware.AuthRequired(),
			middleware.RequirePermission("view_dashboards"),
			listDashboardsHandler)
		group.POST("/",
			middleware.AuthRequired(),
			middleware.RequirePermission("manage_dashboards"),
			createDashboardHandler)
		group.GET("/:id",
			middleware.AuthRequired(),
			middleware.RequirePermission("view_dashboards"),
			getDashboardHandler)
		group.PUT("/:id",
			middleware.AuthRequired(),
			middleware.RequirePermission("manage_dashboards"),
			updateDashboardHandler)
		group.DELETE("/:id",
			middleware.AuthRequired(),
			middleware.RequirePermission("manage_dashboards"),
			deleteDashboardHandler)
		group.POST("/:id/share",
			middleware.AuthRequired(),
			middleware.RequirePermission("manage_dashboards"),
			shareDashboardHandler)
		group.PUT("/:id/default",
			middleware.AuthRequired(),
			middleware.RequirePermission("manage_dashboards"),
			setDefaultDashboardHandler)
	}
}
```

**Step 3: Register routes in `api.go`**

In `sensor_hub/api/api.go`, find the route registration block and add:
```go
RegisterDashboardRoutes(apiGroup)
```

**Step 4: Wire in `cmd/serve.go`**

In the DI block, add after existing repository/service creation:
```go
dashboardRepo := db.NewDashboardRepository(database, logger)
dashboardService := service.NewDashboardService(dashboardRepo, logger)
api.InitDashboardAPI(dashboardService)
```

**Step 5: Verify full backend compiles and tests pass**

Run: `cd sensor_hub && go build ./... && go test ./...`

---

### Task 7: Install react-grid-layout + TypeScript types

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/package.json`

**Step 1: Install dependencies**

Run:
```bash
cd sensor_hub/ui/sensor_hub_ui && npm install react-grid-layout @types/react-grid-layout
```

**Step 2: Verify installation**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm ls react-grid-layout`

Expected: `react-grid-layout@X.X.X` listed without errors.

---

### Task 8: Dashboard TypeScript Types + API Client Module

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/types/dashboard.ts`
- Create: `sensor_hub/ui/sensor_hub_ui/src/api/Dashboards.ts`

**Step 1: Create dashboard TypeScript types**

```typescript
// src/types/dashboard.ts

export type WidgetLayout = {
    x: number;
    y: number;
    w: number;
    h: number;
};

export type DashboardWidget = {
    id: string;
    type: string;
    config: Record<string, unknown>;
    layout: WidgetLayout;
};

export type DashboardBreakpoints = {
    lg: number;
    md: number;
    sm: number;
};

export type DashboardConfig = {
    widgets: DashboardWidget[];
    breakpoints: DashboardBreakpoints;
};

export type Dashboard = {
    id: number;
    user_id: number;
    name: string;
    config: string; // JSON string of DashboardConfig
    shared: boolean;
    is_default: boolean;
    created_at: string;
    updated_at: string;
};

export type CreateDashboardRequest = {
    name: string;
    config: DashboardConfig;
};

export type UpdateDashboardRequest = {
    name?: string;
    config: DashboardConfig;
};

export type ShareDashboardRequest = {
    target_user_id: number;
};

export const DEFAULT_BREAKPOINTS: DashboardBreakpoints = { lg: 12, md: 8, sm: 4 };

export const GRID_BREAKPOINTS = { lg: 1200, md: 768, sm: 480 };
export const GRID_COLS = { lg: 12, md: 8, sm: 4 };
export const GRID_ROW_HEIGHT = 80;
```

**Step 2: Create the API client module**

```typescript
// src/api/Dashboards.ts

import { get, post, put, del } from './Client';
import type { Dashboard, CreateDashboardRequest, UpdateDashboardRequest, ShareDashboardRequest } from '../types/dashboard';
import type { ApiMessage } from './Client';

export const list = () => get<Dashboard[]>('/dashboards/');
export const getById = (id: number) => get<Dashboard>(`/dashboards/${id}`);
export const create = (req: CreateDashboardRequest) => post<{ id: number }>('/dashboards/', req);
export const update = (id: number, req: UpdateDashboardRequest) => put<ApiMessage>(`/dashboards/${id}`, req);
export const remove = (id: number) => del<ApiMessage>(`/dashboards/${id}`);
export const share = (id: number, req: ShareDashboardRequest) => post<ApiMessage>(`/dashboards/${id}/share`, req);
export const setDefault = (id: number) => put<ApiMessage>(`/dashboards/${id}/default`);
```

**Step 3: Verify TypeScript compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 9: Widget Registry — type → component mapping

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/WidgetRegistry.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/types.ts`

**Step 1: Create dashboard internal types**

```typescript
// src/dashboard/types.ts

import type { ComponentType } from 'react';

export interface WidgetProps {
    id: string;
    config: Record<string, unknown>;
    isEditing: boolean;
}

export interface WidgetDefinition {
    type: string;
    label: string;
    description: string;
    component: ComponentType<WidgetProps>;
    defaultConfig: Record<string, unknown>;
    defaultLayout: { w: number; h: number };
    minW?: number;
    minH?: number;
    maxW?: number;
    maxH?: number;
    configFields?: WidgetConfigField[];
}

export interface WidgetConfigField {
    key: string;
    label: string;
    type: 'text' | 'number' | 'boolean' | 'select' | 'sensor-select' | 'multi-sensor-select';
    options?: { value: string; label: string }[];
    defaultValue?: unknown;
}
```

**Step 2: Create the widget registry**

```typescript
// src/dashboard/WidgetRegistry.tsx

import type { WidgetDefinition } from './types';

const registry = new Map<string, WidgetDefinition>();

export function registerWidget(definition: WidgetDefinition): void {
    registry.set(definition.type, definition);
}

export function getWidget(type: string): WidgetDefinition | undefined {
    return registry.get(type);
}

export function getAllWidgets(): WidgetDefinition[] {
    return Array.from(registry.values());
}

export function getWidgetComponent(type: string) {
    return registry.get(type)?.component;
}
```

This registry will be populated by widget adapter modules in Task 12-13 and Task 18-26. Each widget registers itself by calling `registerWidget()`.

**Step 3: Verify TypeScript compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 10: `DashboardEngine` Component — renders widgets via react-grid-layout

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/DashboardEngine.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/WidgetFrame.tsx`

**Step 1: Create the WidgetFrame wrapper**

```typescript
// src/dashboard/WidgetFrame.tsx

import { Paper, IconButton, Box, Typography } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import SettingsIcon from '@mui/icons-material/Settings';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import { getWidget } from './WidgetRegistry';
import type { WidgetProps } from './types';
import type { DashboardWidget } from '../types/dashboard';

interface WidgetFrameProps {
    widget: DashboardWidget;
    isEditing: boolean;
    onRemove: (id: string) => void;
    onConfigure: (id: string) => void;
}

export default function WidgetFrame({ widget, isEditing, onRemove, onConfigure }: WidgetFrameProps) {
    const definition = getWidget(widget.type);
    if (!definition) {
        return (
            <Paper sx={{ p: 2, height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="error">Unknown widget: {widget.type}</Typography>
            </Paper>
        );
    }

    const Component = definition.component;
    const widgetProps: WidgetProps = {
        id: widget.id,
        config: widget.config,
        isEditing,
    };

    return (
        <Paper
            elevation={isEditing ? 3 : 1}
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                border: isEditing ? '1px dashed' : '1px solid',
                borderColor: isEditing ? 'primary.main' : 'divider',
                borderRadius: 2,
                position: 'relative',
            }}
        >
            {isEditing && (
                <Box sx={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    px: 1,
                    py: 0.5,
                    bgcolor: 'action.hover',
                    borderBottom: '1px solid',
                    borderColor: 'divider',
                    cursor: 'grab',
                    className: 'drag-handle',
                }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                        <DragIndicatorIcon fontSize="small" color="action" />
                        <Typography variant="caption" color="text.secondary">{definition.label}</Typography>
                    </Box>
                    <Box>
                        <IconButton size="small" onClick={() => onConfigure(widget.id)}>
                            <SettingsIcon fontSize="small" />
                        </IconButton>
                        <IconButton size="small" onClick={() => onRemove(widget.id)}>
                            <CloseIcon fontSize="small" />
                        </IconButton>
                    </Box>
                </Box>
            )}
            <Box sx={{ flex: 1, overflow: 'auto', p: isEditing ? 1 : 0 }}>
                <Component {...widgetProps} />
            </Box>
        </Paper>
    );
}
```

**Step 2: Create the DashboardEngine**

```typescript
// src/dashboard/DashboardEngine.tsx

import { useCallback, useMemo } from 'react';
import { Responsive, WidthProvider } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import WidgetFrame from './WidgetFrame';
import { getWidget } from './WidgetRegistry';
import type { DashboardConfig, DashboardWidget } from '../types/dashboard';
import { GRID_BREAKPOINTS, GRID_COLS, GRID_ROW_HEIGHT } from '../types/dashboard';

const ResponsiveGridLayout = WidthProvider(Responsive);

interface DashboardEngineProps {
    config: DashboardConfig;
    isEditing: boolean;
    onLayoutChange: (widgets: DashboardWidget[]) => void;
    onRemoveWidget: (id: string) => void;
    onConfigureWidget: (id: string) => void;
}

export default function DashboardEngine({
    config,
    isEditing,
    onLayoutChange,
    onRemoveWidget,
    onConfigureWidget,
}: DashboardEngineProps) {
    const layouts = useMemo(() => {
        const lg = config.widgets.map((w) => ({
            i: w.id,
            x: w.layout.x,
            y: w.layout.y,
            w: w.layout.w,
            h: w.layout.h,
            minW: getWidget(w.type)?.minW,
            minH: getWidget(w.type)?.minH,
            maxW: getWidget(w.type)?.maxW,
            maxH: getWidget(w.type)?.maxH,
        }));
        return { lg };
    }, [config.widgets]);

    const handleLayoutChange = useCallback(
        (layout: ReactGridLayout.Layout[]) => {
            const updated = config.widgets.map((widget) => {
                const item = layout.find((l) => l.i === widget.id);
                if (!item) return widget;
                return {
                    ...widget,
                    layout: { x: item.x, y: item.y, w: item.w, h: item.h },
                };
            });
            onLayoutChange(updated);
        },
        [config.widgets, onLayoutChange],
    );

    return (
        <ResponsiveGridLayout
            layouts={layouts}
            breakpoints={GRID_BREAKPOINTS}
            cols={GRID_COLS}
            rowHeight={GRID_ROW_HEIGHT}
            isDraggable={isEditing}
            isResizable={isEditing}
            draggableHandle=".drag-handle"
            onLayoutChange={handleLayoutChange}
            compactType="vertical"
            margin={[16, 16]}
        >
            {config.widgets.map((widget) => (
                <div key={widget.id}>
                    <WidgetFrame
                        widget={widget}
                        isEditing={isEditing}
                        onRemove={onRemoveWidget}
                        onConfigure={onConfigureWidget}
                    />
                </div>
            ))}
        </ResponsiveGridLayout>
    );
}
```

**Step 3: Verify TypeScript compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 11: `DashboardProvider` Context — load/save dashboard state

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/DashboardContext.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/index.ts`

**Step 1: Create the DashboardContext**

```typescript
// src/dashboard/DashboardContext.tsx

import { createContext, useCallback, useContext, useEffect, useState } from 'react';
import * as DashboardsApi from '../api/Dashboards';
import type { Dashboard, DashboardConfig, DashboardWidget, CreateDashboardRequest } from '../types/dashboard';
import { DEFAULT_BREAKPOINTS } from '../types/dashboard';
import { logger } from '../tools/logger';

interface DashboardContextValue {
    dashboards: Dashboard[];
    activeDashboard: Dashboard | null;
    config: DashboardConfig;
    isEditing: boolean;
    loading: boolean;
    setIsEditing: (editing: boolean) => void;
    setActiveDashboard: (dashboard: Dashboard) => void;
    updateWidgets: (widgets: DashboardWidget[]) => void;
    addWidget: (widget: DashboardWidget) => void;
    removeWidget: (id: string) => void;
    updateWidgetConfig: (id: string, config: Record<string, unknown>) => void;
    saveDashboard: () => Promise<void>;
    createDashboard: (req: CreateDashboardRequest) => Promise<number>;
    deleteDashboard: (id: number) => Promise<void>;
    refreshDashboards: () => Promise<void>;
}

const DashboardContext = createContext<DashboardContextValue | null>(null);

const EMPTY_CONFIG: DashboardConfig = { widgets: [], breakpoints: DEFAULT_BREAKPOINTS };

function parseConfig(raw: string): DashboardConfig {
    try {
        return JSON.parse(raw) as DashboardConfig;
    } catch {
        logger.error('[Dashboard] Failed to parse config', raw);
        return EMPTY_CONFIG;
    }
}

export function DashboardProvider({ children }: { children: React.ReactNode }) {
    const [dashboards, setDashboards] = useState<Dashboard[]>([]);
    const [activeDashboard, setActiveDashboardState] = useState<Dashboard | null>(null);
    const [config, setConfig] = useState<DashboardConfig>(EMPTY_CONFIG);
    const [isEditing, setIsEditing] = useState(false);
    const [loading, setLoading] = useState(true);

    const refreshDashboards = useCallback(async () => {
        try {
            const list = await DashboardsApi.list();
            setDashboards(list ?? []);
        } catch (err) {
            logger.error('[Dashboard] Failed to load dashboards', err);
        }
    }, []);

    useEffect(() => {
        refreshDashboards().then(() => setLoading(false));
    }, [refreshDashboards]);

    // Auto-select default dashboard when list loads
    useEffect(() => {
        if (dashboards.length > 0 && !activeDashboard) {
            const defaultDb = dashboards.find((d) => d.is_default) ?? dashboards[0];
            setActiveDashboardState(defaultDb);
            setConfig(parseConfig(defaultDb.config));
        }
    }, [dashboards, activeDashboard]);

    const setActiveDashboard = useCallback((dashboard: Dashboard) => {
        setActiveDashboardState(dashboard);
        setConfig(parseConfig(dashboard.config));
        setIsEditing(false);
    }, []);

    const updateWidgets = useCallback((widgets: DashboardWidget[]) => {
        setConfig((prev) => ({ ...prev, widgets }));
    }, []);

    const addWidget = useCallback((widget: DashboardWidget) => {
        setConfig((prev) => ({ ...prev, widgets: [...prev.widgets, widget] }));
    }, []);

    const removeWidget = useCallback((id: string) => {
        setConfig((prev) => ({ ...prev, widgets: prev.widgets.filter((w) => w.id !== id) }));
    }, []);

    const updateWidgetConfig = useCallback((id: string, widgetConfig: Record<string, unknown>) => {
        setConfig((prev) => ({
            ...prev,
            widgets: prev.widgets.map((w) => (w.id === id ? { ...w, config: widgetConfig } : w)),
        }));
    }, []);

    const saveDashboard = useCallback(async () => {
        if (!activeDashboard) return;
        await DashboardsApi.update(activeDashboard.id, { name: activeDashboard.name, config });
        await refreshDashboards();
        setIsEditing(false);
    }, [activeDashboard, config, refreshDashboards]);

    const createDashboard = useCallback(async (req: CreateDashboardRequest) => {
        const result = await DashboardsApi.create(req);
        await refreshDashboards();
        return result.id;
    }, [refreshDashboards]);

    const deleteDashboard = useCallback(async (id: number) => {
        await DashboardsApi.remove(id);
        if (activeDashboard?.id === id) {
            setActiveDashboardState(null);
            setConfig(EMPTY_CONFIG);
        }
        await refreshDashboards();
    }, [activeDashboard, refreshDashboards]);

    return (
        <DashboardContext.Provider value={{
            dashboards, activeDashboard, config, isEditing, loading,
            setIsEditing, setActiveDashboard, updateWidgets,
            addWidget, removeWidget, updateWidgetConfig,
            saveDashboard, createDashboard, deleteDashboard, refreshDashboards,
        }}>
            {children}
        </DashboardContext.Provider>
    );
}

export function useDashboard() {
    const ctx = useContext(DashboardContext);
    if (!ctx) throw new Error('useDashboard must be used within DashboardProvider');
    return ctx;
}
```

**Step 2: Create the barrel export**

```typescript
// src/dashboard/index.ts

export { DashboardProvider, useDashboard } from './DashboardContext';
export { default as DashboardEngine } from './DashboardEngine';
export { registerWidget, getWidget, getAllWidgets, getWidgetComponent } from './WidgetRegistry';
export type { WidgetProps, WidgetDefinition, WidgetConfigField } from './types';
```

**Step 3: Verify TypeScript compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 12: Widget Wrapper HOC + Adapter Pattern

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/index.ts` (barrel + registration)

The pattern for adapting an existing component: import the existing component, create a thin wrapper accepting `WidgetProps`, register with `registerWidget()`.

**Step 1: Create the widget registration barrel**

```typescript
// src/dashboard/widgets/index.ts
export function registerAllWidgets(): void {
    // Populated in Task 13
}
```

---

### Task 13: Wrap All 8 Existing Components as Widgets

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/TemperatureChartWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/LiveReadingsTableWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/WeatherForecastWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/SensorHealthPieWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/SensorTypePieWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/HealthTimelineWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/ReadingStatsWidget.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/NotificationsFeedWidget.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/index.ts` — register all 8

**Step 1: Create each adapter widget (same pattern)**

Each adapter follows: import existing component → render inside thin wrapper accepting `WidgetProps`. For example:

```typescript
// src/dashboard/widgets/TemperatureChartWidget.tsx
import type { WidgetProps } from '../types';
import IndoorTemperatureDataCard from '../../components/IndoorTemperatureDataCard';
export default function TemperatureChartWidget(_props: WidgetProps) {
    return <IndoorTemperatureDataCard />;
}
```

Create analogous files for: `LiveReadingsTableWidget` (wraps `CurrentTemperatures`), `WeatherForecastWidget` (wraps `WeatherForecastCard`), `SensorHealthPieWidget` (wraps `SensorHealthCard`), `SensorTypePieWidget` (wraps `SensorTypeCard`), `HealthTimelineWidget` (wraps `SensorHealthHistoryChartCard` — check if it needs a `sensorName` prop from config), `ReadingStatsWidget` (wraps `TotalReadingsForEachSensorCard`), `NotificationsFeedWidget` (wraps `NotificationsCard`).

**Step 2: Register all 8 in `widgets/index.ts`**

Call `registerWidget()` for each with type, label, description, component, defaultConfig, defaultLayout, and min sizes. See the widget catalog in the issue for type strings.

**Step 3: Call `registerAllWidgets()` at app startup** in `src/SensorHub.tsx` or `src/main.tsx`.

**Step 4: Verify** — `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 14: Edit Mode Toggle — `DashboardToolbar` component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/DashboardToolbar.tsx`

Toolbar contains: dashboard selector (dropdown), edit/lock toggle, save button, add widget button, new dashboard button, delete dashboard button. Uses `useDashboard()` context and `hasPerm(user, 'manage_dashboards')` to gate edit controls.

---

### Task 15: Widget Picker + Config Dialogs

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/WidgetPickerDialog.tsx`
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/WidgetConfigDialog.tsx`

WidgetPickerDialog: lists all registered widgets from `getAllWidgets()`, on select creates a `DashboardWidget` with unique id, default config, and layout `y: Infinity` (auto-place at bottom).

WidgetConfigDialog: renders config fields from `WidgetDefinition.configFields[]` — supports text, number, boolean, select, sensor-select, multi-sensor-select field types. Uses `useSensorContext()` for sensor dropdowns.

---

### Task 16: Dashboard Page — main page component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/dashboard/DashboardPage.tsx`

Wraps everything in `<DashboardProvider>`. Inner component uses `useDashboard()` for state. Renders `DashboardToolbar`, `DashboardEngine`, `WidgetPickerDialog`, `WidgetConfigDialog`, and a create-dashboard dialog. Empty state shows "Add your first widget" button.

---

### Task 17: Routing + Navigation

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx` — add `/dashboard` route
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx` — add Dashboard nav item with `DashboardIcon`

Add route: `<Route path="/dashboard" element={<RequireAuth><DashboardPage /></RequireAuth>} />`

Keep existing `/` (TemperatureDashboard) route intact.

**Verify:** `cd sensor_hub/ui/sensor_hub_ui && npm run build`

---

### Task 18: `current-reading` Widget — Big Number Display

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/CurrentReadingWidget.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/index.ts` — register

Single sensor's current value as a large number with trend arrow (↑/↓/→). Uses `useCurrentTemperatures` hook. Config: `sensorId` (sensor-select), `showUnit` (boolean). Shows min/max context from last readings.

Register with: `type: 'current-reading'`, `defaultLayout: { w: 3, h: 2 }`, `minW: 2, minH: 2`.

Config fields: `sensorName` (sensor-select), `showUnit` (boolean, default true).

---

### Task 19: `gauge` Widget — Visual Gauge

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/GaugeWidget.tsx`

Visual gauge/thermometer showing current value within a min-max range. Uses SVG or a lightweight gauge library. Colour zones: blue (cold) → green (normal) → red (hot).

Config fields: `sensorName` (sensor-select), `min` (number, default 0), `max` (number, default 40), `unit` (text, default '°C').

Register with: `type: 'gauge'`, `defaultLayout: { w: 3, h: 3 }`, `minW: 2, minH: 2`.

---

### Task 20: `group-summary` Widget — Room/Group Average

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/GroupSummaryWidget.tsx`

Average reading across selected sensors. Big number + list of contributing sensors.

Config fields: `sensorNames` (multi-sensor-select), `displayName` (text, e.g. "Upstairs Average").

Register with: `type: 'group-summary'`, `defaultLayout: { w: 4, h: 2 }`, `minW: 3, minH: 2`.

---

### Task 21: `alert-summary` Widget — Active Alerts

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/AlertSummaryWidget.tsx`

Compact list of currently firing alert rules (enabled + recently triggered). Read-only — links to alerts page for management. Uses Alerts API.

No config fields.

Register with: `type: 'alert-summary'`, `defaultLayout: { w: 4, h: 3 }`, `minW: 3, minH: 2`.

---

### Task 22: `min-max-avg` Widget — Period Statistics

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/MinMaxAvgWidget.tsx`

Shows today's (or configurable period) high, low, and average for a sensor. Clean three-number card layout.

Config fields: `sensorName` (sensor-select), `period` (select: 'today', '24h', '7d', '30d').

Register with: `type: 'min-max-avg'`, `defaultLayout: { w: 3, h: 2 }`, `minW: 2, minH: 2`.

---

### Task 23: `heatmap-calendar` Widget — Daily Heatmap

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/HeatmapCalendarWidget.tsx`

GitHub contribution-graph style calendar showing daily average temperature as colour intensity. Build with SVG grid.

Config fields: `sensorName` (sensor-select), `months` (number, default 6).

Register with: `type: 'heatmap-calendar'`, `defaultLayout: { w: 8, h: 3 }`, `minW: 6, minH: 2`.

---

### Task 24: `sensor-uptime` Widget — Uptime Indicator

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/SensorUptimeWidget.tsx`

Percentage uptime over configurable period displayed as a bar or large number. Uses sensor health history API.

Config fields: `sensorName` (sensor-select), `period` (select: '24h', '7d', '30d').

Register with: `type: 'sensor-uptime'`, `defaultLayout: { w: 3, h: 2 }`, `minW: 2, minH: 2`.

---

### Task 25: `multi-sensor-comparison` Widget — Overlay Chart

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/MultiSensorComparisonWidget.tsx`

Multiple sensors on one time-series Recharts line chart for direct comparison. Reuses `TemperatureGraph` logic with configurable sensor list.

Config fields: `sensorNames` (multi-sensor-select), `period` (select: '24h', '7d', '30d').

Register with: `type: 'multi-sensor-comparison'`, `defaultLayout: { w: 8, h: 4 }`, `minW: 4, minH: 3`.

---

### Task 26: `markdown-note` Widget — Text/Markdown Block

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/dashboard/widgets/MarkdownNoteWidget.tsx`

User-defined text block. In view mode renders markdown. In edit mode the widget config dialog shows a textarea.

Config fields: `content` (text, default 'Add a note...').

Register with: `type: 'markdown-note'`, `defaultLayout: { w: 4, h: 2 }`, `minW: 2, minH: 1`.

---

### Task 27: CLI — `sensor-hub dashboards` Commands

**Files:**
- Create: `sensor_hub/cmd/dashboards.go`

**Step 1: Create the CLI commands**

Follow the exact pattern from `cmd/sensors.go`. Create a parent `dashboardsCmd` with subcommands:

```
dashboards
├── list                     — GET /api/dashboards/
├── get [id]                 — GET /api/dashboards/:id
├── create --name "..." --config '{"widgets":[...]}'  — POST /api/dashboards/
├── update [id] --name "..." --config '...'           — PUT /api/dashboards/:id
├── delete [id]              — DELETE /api/dashboards/:id
├── share [id] --user-id N   — POST /api/dashboards/:id/share
└── set-default [id]         — PUT /api/dashboards/:id/default
```

Each command uses `loadClientConfig(cmd)` → `NewClient(serverURL, apiKey, insecure)` → HTTP method → `printJSON(data)`.

**Step 2: Verify it compiles**

Run: `cd sensor_hub && go build ./cmd/`

---

### Task 28: LLM Skill Files — Document Dashboard JSON Schema

**Files:**
- Modify: `sensor_hub/skills/claude.md` — add dashboard CLI commands section
- Modify: `sensor_hub/skills/copilot.md` — add dashboard CLI commands section

**Step 1: Add dashboard documentation to skill files**

Add a new section documenting:
- `sensor-hub dashboards list` / `get` / `create` / `update` / `delete` / `share` / `set-default`
- The dashboard JSON schema with all widget types and their config options
- Example: "Create a dashboard showing bedroom and office temperatures with weather"

This enables LLMs to generate dashboard configurations from natural language.

---

### Task 29: Share Endpoint — `POST /api/dashboards/:id/share`

Already implemented in Task 6 (API handler) and Task 5 (service). This task covers **testing**:

**Files:**
- Create: `sensor_hub/service/dashboard_service_test.go`

**Step 1: Write service unit tests**

Test all service methods with mock repository:
- `TestDashboardService_Create`
- `TestDashboardService_Update_NotOwner` (expect error)
- `TestDashboardService_Delete_NotOwner` (expect error)
- `TestDashboardService_Share`
- `TestDashboardService_SetDefault`

Run: `cd sensor_hub && go test -v ./service/ -run TestDashboard`

---

### Task 30: Default Dashboard — `PUT /api/dashboards/:id/default`

Already implemented in Tasks 5-6. This task covers **API handler tests**:

**Files:**
- Create: `sensor_hub/api/dashboard_api_test.go`

**Step 1: Write API handler tests**

Test each handler with httptest + mock service:
- `TestListDashboardsHandler_Success`
- `TestCreateDashboardHandler_Success`
- `TestCreateDashboardHandler_BadRequest`
- `TestGetDashboardHandler_NotFound`
- `TestUpdateDashboardHandler_Success`
- `TestDeleteDashboardHandler_Success`
- `TestShareDashboardHandler_Success`
- `TestSetDefaultDashboardHandler_Success`

Run: `cd sensor_hub && go test -v ./api/ -run TestDashboard`

---

### Task 31: OpenAPI Spec — Document Dashboard Endpoints

**Files:**
- Modify: `sensor_hub/openapi.yaml`

**Step 1: Add dashboard endpoints to OpenAPI spec**

Add paths for all 7 dashboard endpoints following the existing pattern:
- `x-required-permission` custom extension
- Request/response schemas
- Components: `Dashboard`, `DashboardConfig`, `DashboardWidget`, `WidgetLayout`, `CreateDashboardRequest`, `UpdateDashboardRequest`, `ShareDashboardRequest`

---

### Task 32: Developer Docs — Architecture + Widget Development Guide

**Files:**
- Create: `docs/docs/api/dashboards.md`
- Modify: `docs/docs/development/architecture.md` — add dashboard architecture section
- Modify: `docs/sidebars.ts` — add dashboard docs to sidebar

**Step 1: Create API documentation page**

Document all dashboard endpoints with request/response examples, following the pattern of existing API docs.

**Step 2: Add architecture section**

Document:
- Dashboard JSON schema
- Widget registry pattern
- How to create a new widget (file, registration, config fields)
- Frontend data flow (DashboardProvider → DashboardEngine → WidgetFrame → Component)

**Step 3: Update sidebar**

Add the new docs page to the Docusaurus sidebar configuration.
