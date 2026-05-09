import { beforeAll, describe, expect, it } from 'vitest';
import { getWidget } from './WidgetRegistry';
import { registerAllWidgets } from './widgets';

beforeAll(() => {
  registerAllWidgets();
});

describe('health history widget configuration', () => {
  it('keeps the health timeline widget focused on sensor selection only', () => {
    expect(getWidget('health-timeline')?.configFields?.map((field) => field.key)).toEqual(['sensorId']);
  });

  it('keeps the uptime widget focused on sensor selection only', () => {
    expect(getWidget('uptime')?.configFields?.map((field) => field.key)).toEqual(['sensorId']);
  });
});
