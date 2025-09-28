import { useEffect, useState } from "react";
import type { TemperatureReading } from "../types/types";
import { WEBSOCKET_BASE } from "../environment/Environment";

export function useCurrentTemperatures() {
  const [currentTemperatures, setCurrentTemperatures] = useState<{
    [sensor: string]: TemperatureReading;
  }>({});
  useEffect(() => {
    const ws = new WebSocket(`${WEBSOCKET_BASE}/ws/current-temperatures`);
    ws.onmessage = (event) => {
      if (!event.data || event.data === "null") return;
      const arr = JSON.parse(event.data);
      // Convert array to object keyed by sensor_name
      const obj: { [key: string]: TemperatureReading } = {};
      arr.forEach((reading: TemperatureReading) => {
        obj[String(reading.sensor_name)] = reading;
      });
      setCurrentTemperatures(obj);
    };
    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };
    return () => ws.close();
  }, []);
  return currentTemperatures;
}
