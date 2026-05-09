import type { Sensor } from '../gen/aliases';

type DeviceInfoField = {
  key: string;
  label: string;
};

const deviceInfoFields: DeviceInfoField[] = [
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'model', label: 'Model' },
  { key: 'description', label: 'Description' },
  { key: 'ieee_address', label: 'IEEE Address' },
];

function getMetadataValue(metadata: Sensor['metadata'] | undefined, key: string): string | null {
  const value = metadata?.[key];
  if (typeof value !== 'string') {
    return null;
  }

  const trimmed = value.trim();
  return trimmed === '' ? null : trimmed;
}

export function getDisplayableDeviceInfo(metadata: Sensor['metadata'] | undefined) {
  return deviceInfoFields.flatMap(({ key, label }) => {
    const value = getMetadataValue(metadata, key);
    return value ? [{ key, label, value }] : [];
  });
}

export function getDeviceMetadataSummary(metadata: Sensor['metadata'] | undefined): string | null {
  const manufacturer = getMetadataValue(metadata, 'manufacturer');
  const model = getMetadataValue(metadata, 'model');

  if (manufacturer && model) {
    return `${manufacturer} ${model}`;
  }

  return manufacturer ?? model;
}
