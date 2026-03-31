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
    config: string;
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
