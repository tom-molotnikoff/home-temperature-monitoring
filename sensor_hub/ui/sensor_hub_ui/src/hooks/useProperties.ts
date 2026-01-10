import type {PropertiesApiStructure} from "../types/types.ts";
import {useEffect, useState} from "react";
import {WEBSOCKET_BASE} from "../environment/Environment.ts";
import { useAuth } from '../providers/AuthContext.tsx';


export function useProperties() {
  const [properties, setProperties] = useState<PropertiesApiStructure>({});
  const { user } = useAuth();

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/properties/ws`);
    ws.onmessage = (event) => {
      try {
        if (!event.data || event.data === "null") return;
        const updatedProperties = JSON.parse(event.data) as PropertiesApiStructure;
        setProperties(updatedProperties);
        console.log("Received properties update via WebSocket:", updatedProperties);
      } catch (err) {
        console.error("Failed to handle properties WebSocket message:", err);
      }
    };
    ws.onerror = (event) => {
      console.error("WebSocket error:", event);
    };
    return () => {
      ws.close();
    };
  }, [user]);

  return properties;
}