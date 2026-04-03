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
