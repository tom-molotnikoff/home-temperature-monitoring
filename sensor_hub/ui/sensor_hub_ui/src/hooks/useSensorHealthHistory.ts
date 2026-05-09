import type { SensorHealthHistory } from "../gen/aliases";
import {useCallback, useEffect, useState} from "react";
import { apiClient } from "../gen/client";
import { useAuth } from '../providers/AuthContext.tsx';
import { logger } from '../tools/logger';

function useSensorHealthHistory(sensorName: string): [SensorHealthHistory[], () => Promise<void>] {
  const [healthHistory, setHealthHistory] = useState<SensorHealthHistory[]>([]);

  const fetchHistory = useCallback(async () => {
    try {
      const { data } = await apiClient.GET('/sensors/health/{name}', {
        params: { path: { name: sensorName } },
      });
      setHealthHistory(data ?? []);
    } catch (err) {
      logger.error("Failed to load sensor health history", err);
    }
  }, [sensorName]);

  const { user } = useAuth();

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;
    if (!sensorName) return;
    void fetchHistory();
  }, [fetchHistory, user, sensorName]);

  return [healthHistory, fetchHistory];
}

export default useSensorHealthHistory;
