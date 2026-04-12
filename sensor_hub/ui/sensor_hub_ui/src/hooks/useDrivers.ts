import { useEffect, useState, useCallback } from 'react';
import type { DriverInfo } from '../types/types';
import { DriversApi } from '../api/Sensors';
import { logger } from '../tools/logger';

export function useDrivers(type?: 'pull' | 'push') {
  const [drivers, setDrivers] = useState<DriverInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const list = await DriversApi.list(type);
      setDrivers(list);
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
