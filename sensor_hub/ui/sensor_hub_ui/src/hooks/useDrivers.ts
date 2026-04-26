import { useEffect, useState, useCallback } from 'react';
import type { DriverInfo } from '../gen/aliases';
import { apiClient } from '../gen/client';
import { logger } from '../tools/logger';

export function useDrivers(type?: 'pull' | 'push') {
  const [drivers, setDrivers] = useState<DriverInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const { data } = await apiClient.GET('/drivers', { params: { query: type ? { type } : undefined } });
      setDrivers(data ?? []);
    } catch (err) {
      logger.error('Failed to fetch drivers:', err);
    } finally {
      setLoaded(true);
    }
  }, [type]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { drivers, loaded, refresh };
}
