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
      const arr = JSON.parse(event.data);

      const obj: { [key: string]: TemperatureReading } = {};
      arr.forEach((reading: TemperatureReading) => {
        obj[String(reading.sensor_name)] = reading;
      });
      setCurrentTemperatures(obj);
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
