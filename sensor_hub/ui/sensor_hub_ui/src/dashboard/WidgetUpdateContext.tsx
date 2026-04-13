import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';

interface WidgetUpdateContextValue {
    lastUpdated: Date | null;
    reportUpdate: (date: Date) => void;
}

const WidgetUpdateContext = createContext<WidgetUpdateContextValue>({
    lastUpdated: null,
    reportUpdate: () => {},
});

export function WidgetUpdateProvider({ children }: { children: ReactNode }) {
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
    const reportUpdate = useCallback((date: Date) => setLastUpdated(date), []);

    return (
        <WidgetUpdateContext.Provider value={{ lastUpdated, reportUpdate }}>
            {children}
        </WidgetUpdateContext.Provider>
    );
}

/** Returns the reportUpdate function. Safe to call outside a provider (returns no-op). */
export function useReportWidgetUpdate(): (date: Date) => void {
    return useContext(WidgetUpdateContext).reportUpdate;
}

/** Returns the last updated timestamp for the current widget. */
export function useWidgetLastUpdated(): Date | null {
    return useContext(WidgetUpdateContext).lastUpdated;
}
