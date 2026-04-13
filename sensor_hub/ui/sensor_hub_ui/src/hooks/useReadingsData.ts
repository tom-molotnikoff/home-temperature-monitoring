import { useCallback, useEffect, useRef, useState, useMemo } from "react";
import type { ChartEntry, Sensor, Reading } from "../types/types";
import type { DateTime } from "luxon";
import { ReadingsApi } from "../api/Readings.ts";
import { logger } from '../tools/logger';

interface ResolvedRange {
  startDate: DateTime<boolean> | null;
  endDate: DateTime<boolean> | null;
}

interface useReadingsDataProps {
  startDate: DateTime<boolean> | null;
  endDate: DateTime<boolean> | null;
  sensors: Sensor[];
  useHourlyAverages: boolean;
  pollIntervalMs?: number;
  measurementType?: string;
  /** When provided, called on every poll tick to get fresh dates (for relative presets like "last 24h"). */
  resolveTimeRange?: () => ResolvedRange;
  onDataUpdate?: (date: Date) => void;
}

export function useReadingsData({
                                     startDate,
                                     endDate,
                                     sensors,
                                     useHourlyAverages,
                                     pollIntervalMs = 30000,
                                     measurementType,
                                     resolveTimeRange,
                                     onDataUpdate,
                                   }: useReadingsDataProps) {
  const [mergedData, setMergedData] = useState<ChartEntry[]>([]);
  const prevResponseJsonRef = useRef<string | null>(null);
  const prevSensorsKeyRef = useRef<string>("");
  const prevMergedJsonRef = useRef<string>("");
  const requestIdRef = useRef(0);
  const isMountedRef = useRef(true);

  const sensorsKey = useMemo(() => sensors.map((s) => s.name).join("|"), [sensors]);
  const startIso = startDate?.toUTC().toISO() ?? null;
  const endIso = endDate?.toUTC().toISO() ?? null;

  // Keep a stable ref to the resolver so the interval callback always uses the latest
  const resolveTimeRangeRef = useRef(resolveTimeRange);
  resolveTimeRangeRef.current = resolveTimeRange;

  const onDataUpdateRef = useRef(onDataUpdate);
  onDataUpdateRef.current = onDataUpdate;

  const fetchAndMaybeUpdate = useCallback(async (
    fetchStartIso: string,
    fetchEndIso: string,
    force = false,
  ) => {
    const currentRequestId = ++requestIdRef.current;
    try {
      let data: Reading[] = [];

      if (useHourlyAverages) {
        data = await ReadingsApi.getBetweenDatesHourly(fetchStartIso, fetchEndIso, undefined, measurementType);
      } else {
        data = await ReadingsApi.getBetweenDates(fetchStartIso, fetchEndIso, undefined, measurementType);
      }

      if (requestIdRef.current !== currentRequestId) return;
      const dataJson = JSON.stringify(data ?? []);

      if (!force && prevResponseJsonRef.current === dataJson && prevSensorsKeyRef.current === sensorsKey) {
        onDataUpdateRef.current?.(new Date());
        return;
      }

      const times = Array.from(
        new Set((data ?? []).map((r) => r.time.replace(" ", "T") + "Z"))
      );

      const newMergedData: ChartEntry[] = times.map((time) => {
        const entry: ChartEntry = { time };
        sensors.forEach((sensor) => {
          const found = data.find(
            (r) => r.sensor_name === sensor.name && r.time.replace(" ", "T") + "Z" === time
          );
          entry[sensor.name] = found
            ? found.numeric_value != null
              ? found.numeric_value
              : found.text_state === 'true' ? 1 : found.text_state === 'false' ? 0 : null
            : null;
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
      onDataUpdateRef.current?.(new Date());
    } catch (err: unknown) {
      if (!isMountedRef.current) return;
      logger.error("Error fetching readings:", err);
    }
  }, [useHourlyAverages, sensorsKey, sensors, measurementType]);

  useEffect(() => {
    isMountedRef.current = true;
    if (!startIso || !endIso) return;

    void fetchAndMaybeUpdate(startIso, endIso, true);

    const intervalId = window.setInterval(() => {
      // Re-resolve time range on each tick so relative presets slide forward
      const resolved = resolveTimeRangeRef.current?.();
      const freshStart = resolved?.startDate?.toUTC().toISO() ?? startIso;
      const freshEnd = resolved?.endDate?.toUTC().toISO() ?? endIso;
      void fetchAndMaybeUpdate(freshStart, freshEnd);
    }, pollIntervalMs);

    return () => {
      requestIdRef.current = Number.MAX_SAFE_INTEGER;
      isMountedRef.current = false;
      window.clearInterval(intervalId);
    };
  }, [fetchAndMaybeUpdate, pollIntervalMs, startIso, endIso]);

  return mergedData;
}
