import { renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useWidgetSubtitle } from './useWidgetSubtitle';
import type { Sensor } from '../gen/aliases';

const sensors: Sensor[] = [];
const properties: Record<string, string> = {};

vi.mock('../hooks/useSensorContext', () => ({
  useSensorContext: () => ({
    sensors,
  }),
}));

vi.mock('../hooks/useProperties', () => ({
  useProperties: () => properties,
}));

describe('useWidgetSubtitle', () => {
  beforeEach(() => {
    sensors.splice(0, sensors.length);
    Object.keys(properties).forEach((key) => delete properties[key]);
  });

  it('includes the selected property for sensor toggle widgets', () => {
    sensors.splice(0, sensors.length, {
      id: 7,
      name: 'office-plug',
      external_id: 'office-plug',
      sensor_driver: 'zigbee2mqtt',
      config: {},
      metadata: {},
      capabilities: [],
      health_status: 'good',
      health_reason: 'ok',
      enabled: true,
      status: 'active',
      retention_hours: null,
    });

    const { result } = renderHook(() => useWidgetSubtitle('sensor-toggle', { sensorId: 7, property: 'state' }));

    expect(result.current).toBe('office-plug · state');
  });

  it('keeps the existing sensor-only subtitle for other sensor widgets', () => {
    sensors.splice(0, sensors.length, {
      id: 7,
      name: 'office-plug',
      external_id: 'office-plug',
      sensor_driver: 'zigbee2mqtt',
      config: {},
      metadata: {},
      capabilities: [],
      health_status: 'good',
      health_reason: 'ok',
      enabled: true,
      status: 'active',
      retention_hours: null,
    });

    const { result } = renderHook(() => useWidgetSubtitle('gauge', { sensorId: 7, property: 'state' }));

    expect(result.current).toBe('office-plug');
  });
});
