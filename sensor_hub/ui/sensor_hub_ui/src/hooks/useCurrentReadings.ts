import { useCallback, useRef, useState } from "react";
import type { CommandStatusMessage, Reading } from "../gen/aliases";
import { WEBSOCKET_BASE } from "../environment/Environment";
import { useAuth } from "../providers/AuthContext.tsx";
import { logger } from '../tools/logger';
import { useReconnectingWebSocket } from './useReconnectingWebSocket';

export type CurrentReadingsMap = Record<string, Record<string, Reading>>;

interface UseCurrentReadingsOptions {
    onDataUpdate?: (date: Date) => void;
    onCommandStatus?: (message: CommandStatusMessage) => void;
}

function mergeReadings(prev: CurrentReadingsMap, readings: Reading[]): CurrentReadingsMap {
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
}

export function useCurrentReadings(options?: UseCurrentReadingsOptions): CurrentReadingsMap {
  const [currentReadings, setCurrentReadings] = useState<CurrentReadingsMap>({});
  const { user } = useAuth();

  const onDataUpdateRef = useRef(options?.onDataUpdate);
  onDataUpdateRef.current = options?.onDataUpdate;
  const onCommandStatusRef = useRef(options?.onCommandStatus);
  onCommandStatusRef.current = options?.onCommandStatus;

  const handleMessage = useCallback((event: MessageEvent) => {
      if (!event.data || event.data === "null") return;
      let parsed: unknown;
      try {
        parsed = JSON.parse(event.data);
      } catch (e) {
        logger.error("Readings WS: failed to parse message", e, event.data);
        return;
      }

      if (Array.isArray(parsed)) {
        setCurrentReadings((prev) => mergeReadings(prev, parsed as Reading[]));
        onDataUpdateRef.current?.(new Date());
        return;
      }

      if (typeof parsed !== 'object' || parsed === null) return;

      if ('type' in parsed && parsed.type === 'command_status') {
        onCommandStatusRef.current?.(parsed as CommandStatusMessage);
        return;
      }

      if ('readings' in parsed && Array.isArray(parsed.readings)) {
        setCurrentReadings((prev) => mergeReadings(prev, parsed.readings as Reading[]));
        onDataUpdateRef.current?.(new Date());
        return;
      }

      if ('sensor_name' in parsed && 'measurement_type' in parsed) {
        setCurrentReadings((prev) => mergeReadings(prev, [parsed as Reading]));
        onDataUpdateRef.current?.(new Date());
      }
  }, []);

  useReconnectingWebSocket({
      url: `${WEBSOCKET_BASE}/readings/ws/current`,
      onMessage: handleMessage,
      enabled: user != null,
  });

  return currentReadings;
}
