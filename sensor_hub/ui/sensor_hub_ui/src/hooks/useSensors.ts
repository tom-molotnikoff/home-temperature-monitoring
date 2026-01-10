import { useEffect, useState, useRef } from "react";
import type {Sensor, SensorJson} from "../types/types";
import {WEBSOCKET_BASE} from "../environment/Environment.ts";
import { useAuth } from "../providers/AuthContext.tsx";


interface useSensorsProps {
  type: string;
}

function arraysEqual(a: Sensor[], b: Sensor[]) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) {
    const ai = a[i];
    const bi = b[i];
    if (
      ai.id !== bi.id ||
      ai.name !== bi.name ||
      ai.type !== bi.type ||
      ai.url !== bi.url ||
      ai.enabled !== bi.enabled ||
      ai.healthStatus !== bi.healthStatus ||
      ai.healthReason !== bi.healthReason
    ) return false;
  }
  return true;
}

function mapSensor(sj: SensorJson): Sensor {
  const normalizedHealthStatus = (sj.health_status ?? 'unknown') as Sensor['healthStatus'];
  const normalizedHealthReason = sj.health_reason ?? null;

  return {
    id: sj.id,
    name: sj.name,
    type: sj.type,
    url: sj.url,
    healthStatus: normalizedHealthStatus,
    healthReason: normalizedHealthReason,
    enabled: sj.enabled,
  };
}

export function useSensors({ type }: useSensorsProps) {
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const sensorsRef = useRef<Sensor[]>([]);
  const { user } = useAuth();

  useEffect(() => {
    sensorsRef.current = sensors;
  }, [sensors]);

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/sensors/ws/${encodeURIComponent(type)}`);
    ws.onmessage = (event) => {
      try {
        if (!event.data || event.data === "null") return;
        const arr = JSON.parse(event.data) as SensorJson[];
        const allSensors: Sensor[] = arr.map(mapSensor);
        const sortedSensors = allSensors.sort((a, b) => a.name.localeCompare(b.name));

        if (!arraysEqual(sensorsRef.current, sortedSensors)) {
          sensorsRef.current = sortedSensors;
          setSensors(sortedSensors);
        }
      } catch (err) {
        console.error("Failed to handle sensors WebSocket message:", err);
      }
    };
    ws.onerror = (err) => {
      console.error("Sensors WebSocket error:", err);
    };
    ws.onclose = (event) => {
      console.debug("Sensors WebSocket closed", event);
    };
    return () => ws.close();
  }, [type, user]);

  return sensors;
}