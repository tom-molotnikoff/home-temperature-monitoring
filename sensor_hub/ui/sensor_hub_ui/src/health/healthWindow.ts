import type { SensorHealthHistory, SensorHealthStatus } from '../gen/aliases';

export interface HealthWindowPoint {
  recorded_at: string;
  health_status: SensorHealthStatus;
  synthetic: boolean;
}

export interface HealthWindowModel {
  points: HealthWindowPoint[];
  currentStatus: SensorHealthStatus | null;
  lastTransitionAt: string | null;
  durationsMs: Record<SensorHealthStatus, number>;
  windowDurationMs: number;
  goodRatio: number;
}

export function formatDurationShort(durationMs: number): string {
  const totalMinutes = Math.max(0, Math.round(durationMs / (60 * 1000)));
  const totalHours = Math.floor(totalMinutes / 60);
  const days = Math.floor(totalHours / 24);
  const hours = totalHours % 24;
  const minutes = totalMinutes % 60;

  if (days > 0) return hours > 0 ? `${days}d ${hours}h` : `${days}d`;
  if (totalHours > 0) return minutes > 0 ? `${totalHours}h ${minutes}m` : `${totalHours}h`;
  return `${Math.max(1, totalMinutes)}m`;
}

export function formatWindowLabel(windowDurationMs: number): string {
  const totalHours = Math.round(windowDurationMs / (60 * 60 * 1000));
  if (totalHours % 24 === 0) {
    const days = totalHours / 24;
    return days === 1 ? '24h' : `${days}d`;
  }
  return formatDurationShort(windowDurationMs);
}

interface BuildHealthWindowModelOptions {
  windowStart: Date;
  now: Date;
}

export function buildHealthWindowModel(
  history: SensorHealthHistory[],
  options: BuildHealthWindowModelOptions,
): HealthWindowModel {
  const points = [...history]
    .sort((left, right) => new Date(left.recorded_at).getTime() - new Date(right.recorded_at).getTime())
    .map((entry) => ({
      recorded_at: entry.recorded_at,
      health_status: entry.health_status,
      synthetic: false,
    }));

  if (points.length === 0) {
    return {
      points: [],
      currentStatus: null,
      lastTransitionAt: null,
      durationsMs: { good: 0, bad: 0, unknown: 0 },
      windowDurationMs: Math.max(0, options.now.getTime() - options.windowStart.getTime()),
      goodRatio: 0,
    };
  }

  if (new Date(points[0].recorded_at).getTime() > options.windowStart.getTime()) {
    points.unshift({
      recorded_at: options.windowStart.toISOString(),
      health_status: 'unknown',
      synthetic: true,
    });
  }

  const lastPoint = points[points.length - 1];
  if (new Date(lastPoint.recorded_at).getTime() < options.now.getTime()) {
    points.push({
      recorded_at: options.now.toISOString(),
      health_status: lastPoint.health_status,
      synthetic: true,
    });
  }

  const durationsMs: Record<SensorHealthStatus, number> = { good: 0, bad: 0, unknown: 0 };
  for (let i = 0; i < points.length - 1; i += 1) {
    const current = points[i];
    const next = points[i + 1];
    const segmentDuration = Math.max(0, new Date(next.recorded_at).getTime() - new Date(current.recorded_at).getTime());
    durationsMs[current.health_status] += segmentDuration;
  }

  const windowDurationMs = Math.max(0, options.now.getTime() - options.windowStart.getTime());

  return {
    points,
    currentStatus: points[points.length - 1]?.health_status ?? null,
    lastTransitionAt: points.length > 1 ? points[points.length - 2].recorded_at : points[0].recorded_at,
    durationsMs,
    windowDurationMs,
    goodRatio: windowDurationMs > 0 ? durationsMs.good / windowDurationMs : 0,
  };
}
