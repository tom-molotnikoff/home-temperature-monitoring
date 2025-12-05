import { useEffect, useState } from "react";
import type { TemperatureReading } from "../types/types";
import { WEBSOCKET_BASE } from "../environment/Environment";

export function useCurrentTemperatures() {
  const [currentTemperatures, setCurrentTemperatures] = useState<{
    [sensor: string]: TemperatureReading;
  }>({});
  useEffect(() => {
    const ws = new WebSocket(`${WEBSOCKET_BASE}/temperature/ws/current-temperatures`);
    ws.onmessage = (event) => {
      if (!event.data || event.data === "null") return;
      let parsed: unknown;
      try {
        parsed = JSON.parse(event.data);
      } catch (e) {
        console.error("Temperatures WS: failed to parse message", e, event.data);
        return;
      }

      if (!Array.isArray(parsed)) return;

      const readings = parsed as TemperatureReading[];
      setCurrentTemperatures((prev) => {
        const next: { [sensor: string]: TemperatureReading } = { ...prev };
        readings.forEach((reading) => {
          if (!reading) return;
          next[reading.sensor_name] = reading;
        });
        return next;
      });
    };
    ws.onerror = (err) => {
      console.error("Temperatures WebSocket error:", err);
    };
    ws.onclose = (event) => {
      console.debug("Temperatures WebSocket closed", event);
    };
    return () => ws.close();
  }, []);
  return currentTemperatures;
}
