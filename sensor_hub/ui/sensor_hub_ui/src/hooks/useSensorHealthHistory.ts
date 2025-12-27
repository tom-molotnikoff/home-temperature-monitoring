import type {SensorHealthHistory, SensorHealthHistoryJson} from "../types/types.ts";
import {useCallback, useEffect, useState} from "react";
import {SensorsApi} from "../api/Sensors.ts";

function useSensorHealthHistory(sensorName: string, limit?: number): [SensorHealthHistory[], () => Promise<void>] {
  const [healthHistory, setHealthHistory] = useState<SensorHealthHistory[]>([]);

  if (!limit) {
    limit = 5000;
  }

  const fetchHistory = useCallback(async () => {
    try {
      const data = await SensorsApi.healthHistoryByName(sensorName, limit);
      setHealthHistory(mapSensorHealthHistoryJson(data));
    } catch (err) {
      console.error("Failed to load sensor health history", err);
    }
  }, [sensorName, limit]);

  useEffect(() => {
    void fetchHistory();
  }, [fetchHistory]);

  return [healthHistory, fetchHistory];
}

function mapSensorHealthHistoryJson(shh: SensorHealthHistoryJson[]): SensorHealthHistory[] {
  return shh.map(s => ({
    id: s.id,
    sensorId: s.sensor_id,
    healthStatus: s.health_status,
    recordedAt: new Date(s.recorded_at),
  }));
}

export default useSensorHealthHistory;