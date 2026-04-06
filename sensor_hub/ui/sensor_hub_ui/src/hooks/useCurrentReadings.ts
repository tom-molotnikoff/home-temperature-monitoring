import { useEffect, useState } from "react";
import type { Reading } from "../types/types";
import { WEBSOCKET_BASE } from "../environment/Environment";
import { useAuth } from "../providers/AuthContext.tsx";
import { logger } from '../tools/logger';

export function useCurrentReadings() {
  const [currentReadings, setCurrentReadings] = useState<{
    [sensor: string]: Reading;
  }>({});
  const { user } = useAuth();

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/readings/ws/current`);
    ws.onmessage = (event) => {
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
        const next: { [sensor: string]: Reading } = { ...prev };
        readings.forEach((reading) => {
          if (!reading) return;
          next[reading.sensor_name] = reading;
        });
        return next;
      });
    };
    ws.onerror = (err) => {
      logger.error("Readings WebSocket error:", err);
    };
    ws.onclose = (event) => {
      logger.debug("Readings WebSocket closed", event);
    };
    return () => ws.close();
  }, [user]);
  return currentReadings;
}
