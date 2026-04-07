import { useEffect, useState, useCallback } from 'react';
import type { DriverInfo } from '../types/types';
import { DriversApi } from '../api/Sensors';
import { logger } from '../tools/logger';

export function useDrivers() {
  const [drivers, setDrivers] = useState<DriverInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const list = await DriversApi.list();
      setDrivers(list);
    } catch (err) {
      logger.error('Failed to fetch drivers:', err);
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { drivers, loaded, refresh };
}
