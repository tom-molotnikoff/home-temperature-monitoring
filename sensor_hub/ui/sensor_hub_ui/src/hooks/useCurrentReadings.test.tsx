import { renderHook, act } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

let socketMessageHandler: ((event: MessageEvent) => void) | undefined;

vi.mock('../providers/AuthContext.tsx', () => ({
  useAuth: () => ({
    user: {
      id: 1,
      username: 'operator',
      roles: [],
      permissions: ['view_readings'],
    },
  }),
}));

vi.mock('./useReconnectingWebSocket', () => ({
  useReconnectingWebSocket: (options: { onMessage: (event: MessageEvent) => void }) => {
    socketMessageHandler = options.onMessage;
  },
}));

describe('useCurrentReadings', () => {
  beforeEach(() => {
    socketMessageHandler = undefined;
    vi.resetModules();
  });

  it('reuses the last known readings when the hook remounts', async () => {
    const { useCurrentReadings } = await import('./useCurrentReadings');
    const { result, unmount } = renderHook(() => useCurrentReadings());

    act(() => {
      socketMessageHandler?.(new MessageEvent('message', {
        data: JSON.stringify([{
          id: 99,
          sensor_name: 'office-plug',
          measurement_type: 'state',
          numeric_value: null,
          text_state: 'ON',
          unit: '',
          time: '2026-05-09T20:00:00Z',
        }]),
      }));
    });

    expect(result.current['office-plug']?.state?.text_state).toBe('ON');

    unmount();

    const { useCurrentReadings: useCurrentReadingsRemount } = await import('./useCurrentReadings');
    const { result: remountedResult } = renderHook(() => useCurrentReadingsRemount());

    expect(remountedResult.current['office-plug']?.state?.text_state).toBe('ON');
  });
});
