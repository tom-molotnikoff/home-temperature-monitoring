import { describe, expect, it } from 'vitest';
import type { SensorHealthHistory } from '../gen/aliases';
import { buildHealthWindowModel } from './healthWindow';

function makeHistory(overrides: Partial<SensorHealthHistory> = {}): SensorHealthHistory {
  return {
    id: 1,
    sensor_id: 'sensor-1',
    health_status: 'good',
    recorded_at: '2026-05-09T10:00:00Z',
    ...overrides,
  };
}

describe('buildHealthWindowModel', () => {
  it('extends the latest known health state to now for timeline rendering', () => {
    const model = buildHealthWindowModel(
      [
        makeHistory({
          recorded_at: '2026-05-09T10:00:00Z',
          health_status: 'good',
        }),
      ],
      {
        windowStart: new Date('2026-05-09T10:00:00Z'),
        now: new Date('2026-05-09T12:00:00Z'),
      },
    );

    expect(model.points).toHaveLength(2);
    expect(model.points[0]).toMatchObject({
      recorded_at: '2026-05-09T10:00:00Z',
      health_status: 'good',
      synthetic: false,
    });
    expect(model.points[1]).toMatchObject({
      recorded_at: '2026-05-09T12:00:00.000Z',
      health_status: 'good',
      synthetic: true,
    });
  });

  it('computes time-weighted status durations across the retained window', () => {
    const model = buildHealthWindowModel(
      [
        makeHistory({
          recorded_at: '2026-05-09T10:00:00Z',
          health_status: 'good',
        }),
        makeHistory({
          id: 2,
          recorded_at: '2026-05-09T11:00:00Z',
          health_status: 'bad',
        }),
      ],
      {
        windowStart: new Date('2026-05-09T10:00:00Z'),
        now: new Date('2026-05-09T13:00:00Z'),
      },
    );

    expect(model.currentStatus).toBe('bad');
    expect(model.lastTransitionAt).toBe('2026-05-09T11:00:00Z');
    expect(model.durationsMs).toEqual({
      good: 60 * 60 * 1000,
      bad: 2 * 60 * 60 * 1000,
      unknown: 0,
    });
    expect(model.windowDurationMs).toBe(3 * 60 * 60 * 1000);
    expect(model.goodRatio).toBeCloseTo(1 / 3);
  });

  it('treats the retained-window gap before the first record as unknown time', () => {
    const model = buildHealthWindowModel(
      [
        makeHistory({
          recorded_at: '2026-05-09T10:00:00Z',
          health_status: 'good',
        }),
      ],
      {
        windowStart: new Date('2026-05-09T09:00:00Z'),
        now: new Date('2026-05-09T13:00:00Z'),
      },
    );

    expect(model.points[0]).toMatchObject({
      recorded_at: '2026-05-09T09:00:00.000Z',
      health_status: 'unknown',
      synthetic: true,
    });
    expect(model.durationsMs).toEqual({
      good: 3 * 60 * 60 * 1000,
      bad: 0,
      unknown: 60 * 60 * 1000,
    });
    expect(model.windowDurationMs).toBe(4 * 60 * 60 * 1000);
    expect(model.goodRatio).toBeCloseTo(0.75);
  });

  it('clamps a pre-window baseline checkpoint to the requested window start', () => {
    const model = buildHealthWindowModel(
      [
        makeHistory({
          recorded_at: '2026-05-09T08:59:00Z',
          health_status: 'good',
        }),
      ],
      {
        windowStart: new Date('2026-05-09T09:00:00Z'),
        now: new Date('2026-05-09T13:00:00Z'),
      },
    );

    expect(model.points).toEqual([
      {
        recorded_at: '2026-05-09T09:00:00.000Z',
        health_status: 'good',
        synthetic: true,
      },
      {
        recorded_at: '2026-05-09T13:00:00.000Z',
        health_status: 'good',
        synthetic: true,
      },
    ]);
    expect(model.durationsMs).toEqual({
      good: 4 * 60 * 60 * 1000,
      bad: 0,
      unknown: 0,
    });
    expect(model.windowDurationMs).toBe(4 * 60 * 60 * 1000);
    expect(model.goodRatio).toBe(1);
  });
});
