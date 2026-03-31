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
