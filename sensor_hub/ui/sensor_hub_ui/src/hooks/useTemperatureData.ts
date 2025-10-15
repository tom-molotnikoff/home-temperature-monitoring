import { useEffect, useState } from "react";
import type {ChartEntry, Sensor, TemperatureReading} from "../types/types";
import { API_BASE } from "../environment/Environment";
import type { DateTime } from "luxon";

interface useTemperatureDataProps {
  startDate: DateTime<boolean> | null;
  endDate: DateTime<boolean> | null;
  sensors: Sensor[];
  useHourlyAverages: boolean;
}

export function useTemperatureData({
  startDate,
  endDate,
  sensors,
  useHourlyAverages,
}: useTemperatureDataProps) {
  const [mergedData, setMergedData] = useState<ChartEntry[]>([]);
  useEffect(() => {
    if (!startDate || !endDate) {
      return;
    }

    const fetchReadings = async () => {
      let response: Response;
      if (useHourlyAverages) {
        response = await fetch(
          `${API_BASE}/temperature/readings/hourly/between?start=${startDate.toISODate()}&end=${endDate.toISODate()}`
        );
      } else {
        response = await fetch(
          `${API_BASE}/temperature/readings/between?start=${startDate.toISODate()}&end=${endDate.toISODate()}`
        );
      }
      if (!response.ok) {
        throw new Error("Failed to fetch readings");
      }

      const data: TemperatureReading[] = await response.json();

      const times = Array.from(
        new Set((data ?? []).map((r) => r.time.replace(" ", "T")))
      );

      const mergedData: ChartEntry[] = times.map((time) => {
        const entry: ChartEntry = { time };
        sensors.forEach((sensor) => {
          const found = data.find(
            (r) =>
              r.sensor_name === sensor.name &&
              r.time.replace(" ", "T") === time
          );
          entry[sensor.name] = found ? found.temperature : null;
        });
        return entry;
      });
      setMergedData(mergedData);
    };

    fetchReadings().catch((error) => {
      console.error("Error fetching readings:", error);
    });
  }, [startDate, endDate, sensors, useHourlyAverages]);

  return mergedData;
}
