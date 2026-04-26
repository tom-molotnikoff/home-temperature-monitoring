import {useEffect, useState} from "react";
import {WEBSOCKET_BASE} from "../environment/Environment.ts";
import { useAuth } from '../providers/AuthContext.tsx';
import { logger } from '../tools/logger';


export function useProperties() {
  const [properties, setProperties] = useState<Record<string, string>>({});
  const { user } = useAuth();

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;

    const ws = new WebSocket(`${WEBSOCKET_BASE}/properties/ws`);
    ws.onmessage = (event) => {
      try {
        if (!event.data || event.data === "null") return;
        const updatedProperties = JSON.parse(event.data) as Record<string, string>;
        setProperties(updatedProperties);
        logger.debug("Received properties update via WebSocket:", updatedProperties);
      } catch (err) {
        logger.error("Failed to handle properties WebSocket message:", err);
      }
    };
    ws.onerror = (event) => {
      logger.error("WebSocket error:", event);
    };
    return () => {
      ws.close();
    };
  }, [user]);

  return properties;
}