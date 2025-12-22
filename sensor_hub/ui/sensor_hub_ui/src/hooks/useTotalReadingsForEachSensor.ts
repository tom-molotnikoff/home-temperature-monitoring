import {useCallback, useEffect, useState} from "react";
import {SensorsApi} from "../api/Sensors.ts";


function useTotalReadingsForEachSensor(): [Record<string, number>, () => Promise<void>] {
  const [totalReadingsPerSensor, setTotalReadingsPerSensor] = useState<Record<string, number>>({});

  const fetchTotalReadings = useCallback(async () => {
    try {
      const data = await SensorsApi.totalReadingsForEachSensor();
      setTotalReadingsPerSensor(data);
    } catch (err) {
      console.error("Failed to load total readings for each sensor", err);
    }
  }, []);

  useEffect(() => {
    void fetchTotalReadings();
  }, [fetchTotalReadings]);

  return [totalReadingsPerSensor, fetchTotalReadings];
}

export default useTotalReadingsForEachSensor;