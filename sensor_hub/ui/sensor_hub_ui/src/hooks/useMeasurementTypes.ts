import { useEffect, useState, useCallback } from 'react';
import type { MeasurementTypeInfo } from '../gen/aliases';
import { apiClient } from '../gen/client';
import { logger } from '../tools/logger';

export function useMeasurementTypes() {
  const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const { data } = await apiClient.GET('/measurement-types');
      setMeasurementTypes(data ?? []);
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

export function useMeasurementTypesWithReadings() {
  const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
  const [loaded, setLoaded] = useState(false);

  const refresh = useCallback(async () => {
    setLoaded(false);
    try {
      const { data } = await apiClient.GET('/measurement-types', { params: { query: { has_readings: true } } });
      setMeasurementTypes(data ?? []);
    } catch (err) {
      logger.error('Failed to fetch measurement types with readings:', err);
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
    if (sensorId === null) {
      setMeasurementTypes([]);
      setLoaded(true);
      return;
    }
    setLoaded(false);
    try {
      const { data } = await apiClient.GET('/sensors/by-id/{id}/measurement-types', {
        params: { path: { id: sensorId } },
      });
      setMeasurementTypes(data ?? []);
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
