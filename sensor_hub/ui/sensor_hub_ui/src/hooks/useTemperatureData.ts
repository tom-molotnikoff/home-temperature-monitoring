import { useEffect, useRef, useState, useMemo } from "react";
import type { ChartEntry, Sensor, TemperatureReading } from "../types/types";
import type { DateTime } from "luxon";
import { TemperatureApi } from "../api/Temperature.ts";

interface useTemperatureDataProps {
  startDate: DateTime<boolean> | null;
  endDate: DateTime<boolean> | null;
  sensors: Sensor[];
  useHourlyAverages: boolean;
  pollIntervalMs?: number;
}

export function useTemperatureData({
                                     startDate,
                                     endDate,
                                     sensors,
                                     useHourlyAverages,
                                     pollIntervalMs = 10000,
                                   }: useTemperatureDataProps) {
  const [mergedData, setMergedData] = useState<ChartEntry[]>([]);
  const prevResponseJsonRef = useRef<string | null>(null);
  const prevSensorsKeyRef = useRef<string>("");
  const prevMergedJsonRef = useRef<string>("");
  const requestIdRef = useRef(0);
  const isMountedRef = useRef(true);

  const sensorsKey = useMemo(() => sensors.map((s) => s.name).join("|"), [sensors]);
  const startIso = startDate?.toISODate() ?? null;
  const endIso = endDate?.toISODate() ?? null;

  useEffect(() => {
    isMountedRef.current = true;
    if (!startIso || !endIso) return;


    const fetchAndMaybeUpdate = async (force = false) => {
      const currentRequestId = ++requestIdRef.current;
      try {
        let data: TemperatureReading[] = [];

        if (useHourlyAverages) {
          data = await TemperatureApi.getBetweenDatesHourly(startIso, endIso);
        } else {
          data = await TemperatureApi.getBetweenDates(startIso, endIso);
        }



        if (requestIdRef.current !== currentRequestId) return;
        const dataJson = JSON.stringify(data ?? []);

        if (!force && prevResponseJsonRef.current === dataJson && prevSensorsKeyRef.current === sensorsKey) {
          return;
        }

        const times = Array.from(
          new Set((data ?? []).map((r) => r.time.replace(" ", "T")))
        );

        const newMergedData: ChartEntry[] = times.map((time) => {
          const entry: ChartEntry = { time };
          sensors.forEach((sensor) => {
            const found = data.find(
              (r) => r.sensor_name === sensor.name && r.time.replace(" ", "T") === time
            );
            entry[sensor.name] = found ? found.temperature : null;
          });
          return entry;
        });

        const newMergedJson = JSON.stringify(newMergedData);

        if (newMergedJson !== prevMergedJsonRef.current) {
          prevMergedJsonRef.current = newMergedJson;
          setMergedData(newMergedData);
        }

        prevResponseJsonRef.current = dataJson;
        prevSensorsKeyRef.current = sensorsKey;
      } catch (err: unknown) {
        if (!isMountedRef.current) return;
        console.error("Error fetching temperature readings:", err);
      }
    };

    void fetchAndMaybeUpdate(true);

    const intervalId = window.setInterval(() => {
      void fetchAndMaybeUpdate();
    }, pollIntervalMs);

    return () => {
      requestIdRef.current = Number.MAX_SAFE_INTEGER;
      isMountedRef.current = false;
      window.clearInterval(intervalId);
    };
  }, [useHourlyAverages, pollIntervalMs, startIso, endIso, sensorsKey, sensors]);

  return mergedData;
}
