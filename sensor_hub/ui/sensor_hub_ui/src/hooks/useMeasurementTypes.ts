import { useEffect, useState, useCallback } from 'react';
import type { MeasurementTypeInfo } from '../types/types';
import { MeasurementTypesApi } from '../api/Sensors';
import { logger } from '../tools/logger';

export function useMeasurementTypes() {
  const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const list = await MeasurementTypesApi.getAll();
      setMeasurementTypes(list);
    } catch (err) {
      logger.error('Failed to fetch measurement types:', err);
    } finally {
      setLoaded(true);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { measurementTypes, loaded, refresh };
}

export function useSensorMeasurementTypes(sensorId: number | null) {
  const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    if (sensorId === null) return;
    setLoaded(false);
    try {
      const list = await MeasurementTypesApi.getForSensor(sensorId);
      setMeasurementTypes(list);
    } catch (err) {
      logger.error('Failed to fetch sensor measurement types:', err);
    } finally {
      setLoaded(true);
    }
  }, [sensorId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return { measurementTypes, loaded, refresh };
}
