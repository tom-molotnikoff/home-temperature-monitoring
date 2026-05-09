import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Sensor } from '../gen/aliases';
import type { AuthUser } from '../providers/AuthContext.tsx';
import SensorInfoCard from './SensorInfoCard';

const { getMock, navigateMock } = vi.hoisted(() => ({
  getMock: vi.fn(),
  navigateMock: vi.fn(),
}));

vi.mock('../gen/client', () => ({
  apiClient: {
    GET: getMock,
    POST: vi.fn(),
    DELETE: vi.fn(),
  },
}));

vi.mock('../hooks/useProperties.ts', () => ({
  useProperties: () => ({
    'sensor.data.retention.days': '90',
  }),
}));

vi.mock('react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router')>();
  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

const user: AuthUser = {
  id: 1,
  username: 'operator',
  roles: [],
  permissions: ['manage_sensors'],
};

function makeSensor(overrides: Partial<Sensor> = {}): Sensor {
  return {
    id: 1,
    name: 'front-door',
    external_id: 'front-door',
    sensor_driver: 'zigbee2mqtt',
    config: {
      topic: 'zigbee2mqtt/front-door',
    },
    metadata: {},
    health_status: 'good',
    health_reason: 'ok',
    enabled: true,
    status: 'active',
    retention_hours: null,
    ...overrides,
  };
}

describe('SensorInfoCard', () => {
  beforeEach(() => {
    getMock.mockReset();
    getMock.mockResolvedValue({ data: [] });
    navigateMock.mockReset();
  });

  it('shows a Device Info section for displayable device metadata', () => {
    render(
      <SensorInfoCard
        sensor={makeSensor({
          metadata: {
            manufacturer: 'Aqara',
            model: 'MCCGQ11LM',
            description: 'Door/window contact sensor',
            ieee_address: '0x00158d0003456789',
            exposes: [{ type: 'binary' }],
          },
        })}
        user={user}
      />,
    );

    expect(screen.getByText('Device Info')).toBeInTheDocument();
    expect(screen.getByText('Manufacturer')).toBeInTheDocument();
    expect(screen.getByText('Aqara')).toBeInTheDocument();
    expect(screen.getByText('Model')).toBeInTheDocument();
    expect(screen.getByText('MCCGQ11LM')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Door/window contact sensor')).toBeInTheDocument();
    expect(screen.getByText('IEEE Address')).toBeInTheDocument();
    expect(screen.getByText('0x00158d0003456789')).toBeInTheDocument();
    expect(screen.queryByText('exposes')).not.toBeInTheDocument();
  });

  it('hides the Device Info section when metadata has no displayable fields', () => {
    render(
      <SensorInfoCard
        sensor={makeSensor({
          metadata: {
            exposes: [{ type: 'binary' }],
          },
        })}
        user={user}
      />,
    );

    expect(screen.queryByText('Device Info')).not.toBeInTheDocument();
    expect(screen.queryByText('Manufacturer')).not.toBeInTheDocument();
    expect(screen.queryByText('IEEE Address')).not.toBeInTheDocument();
  });
});
