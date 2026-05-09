import { describe, expect, it } from 'vitest';
import type { Capability, Sensor } from '../gen/aliases';
import { getControllableSensors, getBinaryCapabilities, normalizeSensorToggleProperty } from './sensorToggleConfig';

function makeCapability(overrides: Partial<Capability> = {}): Capability {
  return {
    property: 'state',
    type: 'binary',
    value_on: 'ON',
    value_off: 'OFF',
    ...overrides,
  };
}

function makeSensor(overrides: Partial<Sensor> = {}): Sensor {
  return {
    id: 7,
    name: 'office-plug',
    external_id: 'office-plug',
    sensor_driver: 'zigbee2mqtt',
    config: {},
    metadata: {},
    capabilities: [makeCapability()],
    health_status: 'good',
    health_reason: 'ok',
    enabled: true,
    status: 'active',
    retention_hours: null,
    ...overrides,
  };
}

describe('sensorToggleConfig', () => {
  it('treats only sensors with binary capabilities as controllable toggle targets', () => {
    const sensors = [
      makeSensor(),
      makeSensor({
        id: 8,
        name: 'mode-selector',
        capabilities: [{ property: 'mode', type: 'enum', values: ['off', 'on'] }],
      }),
    ];

    expect(getControllableSensors(sensors).map((sensor) => sensor.name)).toEqual(['office-plug']);
  });

  it('filters the property list down to binary capabilities and normalizes stale enum properties', () => {
    const binaryCapabilities = getBinaryCapabilities(makeSensor({
      capabilities: [
        makeCapability({ property: 'state' }),
        { property: 'power_on_behavior', type: 'enum', values: ['off', 'on'] },
      ],
    }));

    expect(binaryCapabilities.map((capability) => capability.property)).toEqual(['state']);
    expect(normalizeSensorToggleProperty('power_on_behavior', binaryCapabilities)).toBe('state');
  });
});
