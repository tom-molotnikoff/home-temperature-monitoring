/**
 * Smoke test / proof-of-concept for the openapi-fetch client.
 *
 * Calls GET /sensors using the generated client and logs the typed result.
 * Proves that:
 *   - apiClient makes authenticated requests (CSRF + credentials headers applied)
 *   - The response is strongly typed as components['schemas']['Sensor'][]
 *
 * THROWAWAY — removed in #32 when real API modules are migrated.
 */

import { useEffect } from 'react';
import { apiClient } from './client';
import type { components } from './schema';

type Sensor = components['schemas']['Sensor'];

export function ApiClientSmokeTest() {
  useEffect(() => {
    apiClient.GET('/sensors').then(({ data, error }) => {
      if (error) {
        console.error('[SmokeTest] GET /sensors error', error);
        return;
      }
      // data is typed as Sensor[] — verified by TypeScript
      const sensors: Sensor[] = data ?? [];
      console.info('[SmokeTest] GET /sensors ok — count:', sensors.length, sensors);
    });
  }, []);

  return null;
}
