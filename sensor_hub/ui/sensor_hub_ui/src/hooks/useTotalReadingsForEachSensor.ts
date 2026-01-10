import {useCallback, useEffect, useState} from "react";
import {useAuth} from '../providers/AuthContext.tsx';
import {SensorsApi} from "../api/Sensors.ts";


function useTotalReadingsForEachSensor(): [Record<string, number>, () => Promise<void>] {
  const [totalReadingsPerSensor, setTotalReadingsPerSensor] = useState<Record<string, number>>({});
  const {user} = useAuth();

  const fetchTotalReadings = useCallback(async () => {
    try {
      const data = await SensorsApi.totalReadingsForEachSensor();
      setTotalReadingsPerSensor(data);
    } catch (err) {
      console.error("Failed to load total readings for each sensor", err);
    }
  }, []);

  useEffect(() => {
    if (user === undefined) return;
    if (user === null) return;
    void fetchTotalReadings();
  }, [fetchTotalReadings, user]);

  return [totalReadingsPerSensor, fetchTotalReadings];
}

export default useTotalReadingsForEachSensor;