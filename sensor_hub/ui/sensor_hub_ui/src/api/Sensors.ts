import {get, post, put, del, head, type ApiMessage} from './Client';
import type {
  Sensor,
  SensorHealthHistoryJson,
  SensorJson,
  TotalReadingsCountForEachSensorApiMessage,
  DriverInfo,
  MeasurementTypeInfo,
} from "../types/types.ts";
export type { Sensor };


function mapSensorJson(s: SensorJson): Sensor {
  return {
    id: s.id,
    name: s.name,
    sensorDriver: s.sensor_driver,
    config: s.config ?? {},
    healthStatus: s.health_status,
    healthReason: s.health_reason ?? null,
    enabled: Boolean(s.enabled),
    status: s.status || 'active',
  };
}

type SensorPayload = {
  name: string;
  sensor_driver: string;
  config: Record<string, string>;
};

type SensorPayloadUpdate = {
  name?: string;
  sensor_driver?: string;
  config?: Record<string, string | null>;
};

export const SensorsApi = {
  add: (sensor: SensorPayload) => post<ApiMessage>('/sensors', sensor),
  update: (id: number, sensor: SensorPayloadUpdate) => put<ApiMessage>(`/sensors/${id}`, sensor),
  delete: (name: string) => del<ApiMessage>(`/sensors/${encodeURIComponent(name)}`),
  getByName: (name: string) => get<SensorJson>(`/sensors/${encodeURIComponent(name)}`).then(mapSensorJson),
  getAll: () => get<SensorJson[]>('/sensors').then(list => list.map(mapSensorJson)),
  getByDriver: (driver: string) => get<SensorJson[]>(`/sensors/driver/${encodeURIComponent(driver)}`).then(list => list.map(mapSensorJson)),
  exists: (name: string) => head(`/sensors/${encodeURIComponent(name)}`),
  collectAll: () => post<ApiMessage>('/sensors/collect'),
  collectByName: (name: string) => post<ApiMessage>(`/sensors/collect/${encodeURIComponent(name)}`),
  disableByName: (name: string) => post<ApiMessage>(`/sensors/disable/${encodeURIComponent(name)}`),
  enableByName: (name: string) => post<ApiMessage>(`/sensors/enable/${encodeURIComponent(name)}`),
  healthHistoryByName: (name: string, limit?: number) => get<SensorHealthHistoryJson[]>(`/sensors/health/${encodeURIComponent(name)}${limit ? `?limit=${limit}` : ''}`),
  totalReadingsForEachSensor: () => get<TotalReadingsCountForEachSensorApiMessage>('/sensors/stats/total-readings'),
  getByStatus: (status: string) => get<SensorJson[]>(`/sensors/status/${encodeURIComponent(status)}`).then(list => list.map(mapSensorJson)),
  approve: (id: number) => post<ApiMessage>(`/sensors/approve/${id}`),
  dismiss: (id: number) => post<ApiMessage>(`/sensors/dismiss/${id}`),
}

export const DriversApi = {
  list: (type?: 'pull' | 'push') => get<DriverInfo[]>(type ? `/drivers?type=${type}` : '/drivers'),
}

export const MeasurementTypesApi = {
  getAll: (hasReadings?: boolean) => {
    const params = hasReadings ? '?has_readings=true' : '';
    return get<MeasurementTypeInfo[]>(`/measurement-types${params}`);
  },
  getForSensor: (sensorId: number) => get<MeasurementTypeInfo[]>(`/sensors/by-id/${sensorId}/measurement-types`),
}