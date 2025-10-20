import { useEffect, useRef, useState, useMemo } from "react";
import type { ChartEntry, Sensor, TemperatureReading } from "../types/types";
import { API_BASE } from "../environment/Environment";
import type { DateTime } from "luxon";

interface useTemperatureDataProps {
  startDate: DateTime<boolean> | null;
  endDate: DateTime<boolean> | null;
  sensors: Sensor[];
  useHourlyAverages: boolean;
  pollIntervalMs?: number;
}

export function useTemperatureData({startDate, endDate, sensors, useHourlyAverages, pollIntervalMs = 10000,}: useTemperatureDataProps) {
  const [mergedData, setMergedData] = useState<ChartEntry[]>([]);
  const prevResponseTextRef = useRef<string | null>(null);
  const prevSensorsKeyRef = useRef<string>("");
  const prevMergedJsonRef = useRef<string>("");
  const sensorsKey = useMemo(() => sensors.map((s) => s.name).join("|"), [sensors]);
  const startIso = startDate?.toISODate() ?? null;
  const endIso = endDate?.toISODate() ?? null;

  useEffect(() => {
    if (!startIso || !endIso) return;

    const urlBase = useHourlyAverages
      ? "/temperature/readings/hourly/between"
      : "/temperature/readings/between";
    const buildUrl = () => `${API_BASE}${urlBase}?start=${startIso}&end=${endIso}`;

    let abortController = new AbortController();

    const fetchAndMaybeUpdate = async (force = false) => {
      try {
        const res = await fetch(buildUrl(), { signal: abortController.signal });
        if (!res.ok) throw new Error(`HTTP ${res.status} ${res.statusText}`);
        const text = await res.text();

        if (!force && prevResponseTextRef.current === text && prevSensorsKeyRef.current === sensorsKey) {
          return;
        }

        const data: TemperatureReading[] = text ? JSON.parse(text) : [];

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

        prevResponseTextRef.current = text;
        prevSensorsKeyRef.current = sensorsKey;
      } catch (err: unknown) {
        if ((err as { name?: string })?.name === "AbortError") return;
        console.error("Error fetching temperature readings:", err);
      }
    };

    void fetchAndMaybeUpdate(true);

    const intervalId = window.setInterval(() => {
      abortController = new AbortController();
      void fetchAndMaybeUpdate();
    }, pollIntervalMs);

    return () => {
      window.clearInterval(intervalId);
      abortController.abort();
    };
  }, [useHourlyAverages, pollIntervalMs, startIso, endIso, sensorsKey, sensors]);

  return mergedData;
}