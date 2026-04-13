export type RetentionUnit = 'hours' | 'days' | 'weeks';

const unitMultipliers: Record<RetentionUnit, number> = {
  hours: 1,
  days: 24,
  weeks: 168,
};

export function unitToHours(value: number, unit: RetentionUnit): number {
  return Math.round(value * unitMultipliers[unit]);
}

export function hoursToUnit(hours: number, unit: RetentionUnit): number {
  return Math.round((hours / unitMultipliers[unit]) * 100) / 100;
}

export function formatRetention(hours: number): string {
  if (hours >= 168 && hours % 168 === 0) return `${hours / 168} week${hours / 168 !== 1 ? 's' : ''}`;
  if (hours >= 24 && hours % 24 === 0) return `${hours / 24} day${hours / 24 !== 1 ? 's' : ''}`;
  return `${hours} hour${hours !== 1 ? 's' : ''}`;
}

export function bestUnit(hours: number): RetentionUnit {
  if (hours >= 168 && hours % 168 === 0) return 'weeks';
  if (hours >= 24 && hours % 24 === 0) return 'days';
  return 'hours';
}
