import { useCallback, useRef, useState } from "react";
import type { Reading } from "../types/types";
import { WEBSOCKET_BASE } from "../environment/Environment";
import { useAuth } from "../providers/AuthContext.tsx";
import { logger } from '../tools/logger';
import { useReconnectingWebSocket } from './useReconnectingWebSocket';

export type CurrentReadingsMap = Record<string, Record<string, Reading>>;

interface UseCurrentReadingsOptions {
    onDataUpdate?: (date: Date) => void;
}

export function useCurrentReadings(options?: UseCurrentReadingsOptions): CurrentReadingsMap {
  const [currentReadings, setCurrentReadings] = useState<CurrentReadingsMap>({});
  const { user } = useAuth();

  const onDataUpdateRef = useRef(options?.onDataUpdate);
  onDataUpdateRef.current = options?.onDataUpdate;

  const handleMessage = useCallback((event: MessageEvent) => {
      if (!event.data || event.data === "null") return;
      let parsed: unknown;
      try {
        parsed = JSON.parse(event.data);
      } catch (e) {
        logger.error("Readings WS: failed to parse message", e, event.data);
        return;
      }

      if (!Array.isArray(parsed)) return;

      const readings = parsed as Reading[];
      setCurrentReadings((prev) => {
        const next: CurrentReadingsMap = { ...prev };
        readings.forEach((reading) => {
          if (!reading) return;
          const sensorEntry = next[reading.sensor_name]
            ? { ...next[reading.sensor_name] }
            : {};
          sensorEntry[reading.measurement_type] = reading;
          next[reading.sensor_name] = sensorEntry;
        });
        return next;
      });
      onDataUpdateRef.current?.(new Date());
  }, []);

  useReconnectingWebSocket({
      url: `${WEBSOCKET_BASE}/readings/ws/current`,
      onMessage: handleMessage,
      enabled: user != null,
  });

  return currentReadings;
}
