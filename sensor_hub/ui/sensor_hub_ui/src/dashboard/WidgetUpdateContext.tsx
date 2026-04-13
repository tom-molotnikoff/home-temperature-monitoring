import { createContext, useCallback, useContext, useState, type ReactNode } from 'react';

// Two separate contexts prevent widgets from re-rendering when only the timestamp changes.
// Widgets read ReportContext (stable), only the badge reads ValueContext (changes on updates).
const ReportContext = createContext<(date: Date) => void>(() => {});
const ValueContext = createContext<Date | null>(null);

export function WidgetUpdateProvider({ children }: { children: ReactNode }) {
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
    const reportUpdate = useCallback((date: Date) => setLastUpdated(date), []);

    return (
        <ReportContext.Provider value={reportUpdate}>
            <ValueContext.Provider value={lastUpdated}>
                {children}
            </ValueContext.Provider>
        </ReportContext.Provider>
    );
}

/** Returns the reportUpdate function. Safe to call outside a provider (returns no-op). */
export function useReportWidgetUpdate(): (date: Date) => void {
    return useContext(ReportContext);
}

/** Returns the last updated timestamp for the current widget. */
export function useWidgetLastUpdated(): Date | null {
    return useContext(ValueContext);
}
