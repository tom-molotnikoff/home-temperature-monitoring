import { DateTime } from 'luxon';

export const TIME_RANGE_PRESETS = [
    { value: '1h', label: 'Last 1 hour' },
    { value: '6h', label: 'Last 6 hours' },
    { value: '24h', label: 'Last 24 hours' },
    { value: '3d', label: 'Last 3 days' },
    { value: '7d', label: 'Last 7 days' },
    { value: '30d', label: 'Last 30 days' },
    { value: 'custom', label: 'Custom range' },
] as const;

export type TimeRangePreset = (typeof TIME_RANGE_PRESETS)[number]['value'];

interface ResolvedRange {
    startDate: DateTime;
    endDate: DateTime;
}

const PRESET_DURATIONS: Record<string, { hours?: number; days?: number }> = {
    '1h': { hours: 1 },
    '6h': { hours: 6 },
    '24h': { hours: 24 },
    '3d': { days: 3 },
    '7d': { days: 7 },
    '30d': { days: 30 },
};

/**
 * Resolves a widget config into concrete start/end DateTimes.
 *
 * Supports three shapes:
 *  1. timeRange preset (e.g. "24h") — resolves relative to now
 *  2. timeRange "custom" + customStart/customEnd — uses the stored dates
 *  3. Legacy: no timeRange but has startDate/endDate — treated as custom (backward compat)
 *
 * Falls back to last 24 hours if nothing is configured.
 *
 * Note: the readings API uses SQL BETWEEN with date-only strings, so the
 * end date is effectively exclusive (timestamps like "2026-04-11T14:00:00Z"
 * sort after "2026-04-11" lexicographically). We add one day to the end so
 * that all of today's readings are included.
 */
export function resolveTimeRange(config: Record<string, unknown>): ResolvedRange {
    const now = DateTime.now();
    const tomorrow = now.plus({ days: 1 }).startOf('day');
    const fallback: ResolvedRange = {
        startDate: now.minus({ hours: 24 }),
        endDate: tomorrow,
    };

    const timeRange = config.timeRange as string | undefined;

    // Case 1: relative preset
    if (timeRange && timeRange !== 'custom') {
        const duration = PRESET_DURATIONS[timeRange];
        if (duration) {
            return { startDate: now.minus(duration), endDate: tomorrow };
        }
        return fallback;
    }

    // Case 2: explicit custom range — add a day to end date so the
    // selected end day is included in the BETWEEN query
    if (timeRange === 'custom') {
        const start = parseDate(config.customStart);
        const end = parseDate(config.customEnd);
        return {
            startDate: start ?? fallback.startDate,
            endDate: end ? end.plus({ days: 1 }) : fallback.endDate,
        };
    }

    // Case 3: legacy startDate/endDate (backward compat) — these already
    // had the +1 day baked in by callers, so pass through as-is
    const legacyStart = parseDate(config.startDate);
    const legacyEnd = parseDate(config.endDate);
    if (legacyStart || legacyEnd) {
        return {
            startDate: legacyStart ?? fallback.startDate,
            endDate: legacyEnd ?? fallback.endDate,
        };
    }

    return fallback;
}

function parseDate(value: unknown): DateTime | null {
    if (typeof value !== 'string' || !value) return null;
    const dt = DateTime.fromISO(value);
    return dt.isValid ? dt : null;
}
