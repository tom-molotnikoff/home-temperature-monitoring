import type { Capability, Sensor } from '../gen/aliases';

export function getBinaryCapabilities(sensor: Sensor | null | undefined): Capability[] {
  return (sensor?.capabilities ?? []).filter((capability) => capability.type === 'binary');
}

export function getControllableSensors(sensors: Sensor[]): Sensor[] {
  return sensors.filter((sensor) => getBinaryCapabilities(sensor).length > 0);
}

export function normalizeSensorToggleProperty(property: unknown, binaryCapabilities: Capability[]): string {
  const currentProperty = typeof property === 'string' ? property : '';
  if (binaryCapabilities.some((capability) => capability.property === currentProperty)) {
    return currentProperty;
  }

  return binaryCapabilities.find((capability) => capability.property === 'state')?.property
    ?? binaryCapabilities[0]?.property
    ?? '';
}
