import { useEffect, useState, useCallback } from "react";
import type { Sensor } from "../types/types.ts";
import { SensorsApi } from "../api/Sensors.ts";


interface useSensorsProps {
  types: string[];
  refreshIntervalMs?: number;
}

function arraysEqual(a: Sensor[], b: Sensor[]) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    if (JSON.stringify(a[i]) !== JSON.stringify(b[i])) return false;
  }
  return true;
}

export function useSensors({ types, refreshIntervalMs = 3000 }: useSensorsProps) {
  const [sensors, setSensors] = useState<Sensor[]>([]);

  const fetchSensors = useCallback(async () => {
    try {
      const allSensors: Sensor[] = [];
      for (const type of types) {
        const sensorsOfType: Sensor[] = await SensorsApi.getByType(type);

        allSensors.push(...sensorsOfType);
      }
      const sortedSensors = allSensors.sort((a, b) => a.name.localeCompare(b.name));
      if (!arraysEqual(sensors, sortedSensors)) {
        setSensors(sortedSensors);
      }
    } catch (error) {
      console.error("Error fetching sensors:", error);
    }
  }, [types, sensors]);

  useEffect(() => {
    fetchSensors();
    const interval = setInterval(fetchSensors, refreshIntervalMs);
    return () => clearInterval(interval);
  }, [fetchSensors, refreshIntervalMs]);

  return sensors;
}