import {get, post, put, del, head, type ApiMessage} from './Client';
import type { Sensor, SensorJson } from "../types/types.ts";


function mapSensorJson(s: SensorJson): Sensor {
  return {
    id: s.id,
    name: s.name,
    type: s.type,
    url: s.url,
    healthStatus: s.health_status,
    healthReason: s.health_reason ?? null,
    enabled: Boolean(s.enabled),
  };
}

export const SensorsApi = {
  add: (sensor: Omit<Sensor, 'id' | 'enabled' | 'healthReason' | 'healthStatus'>) => post<ApiMessage>('/sensors/', sensor),
  update: (id: number, sensor: Partial<Omit<Sensor, 'id' | 'healthReason' | 'healthStatus' | 'enabled'>>) => put<ApiMessage>(`/sensors/${id}`, sensor),
  delete: (name: string) => del<ApiMessage>(`/sensors/${encodeURIComponent(name)}`),
  getByName: (name: string) => get<SensorJson>(`/sensors/${encodeURIComponent(name)}`).then(mapSensorJson),
  getAll: () => get<SensorJson[]>('/sensors/').then(list => list.map(mapSensorJson)),
  getByType: (type: string) => get<SensorJson[]>(`/sensors/type/${encodeURIComponent(type)}`).then(list => list.map(mapSensorJson)),
  exists: (name: string) => head(`/sensors/${encodeURIComponent(name)}`),
  collectAll: () => post<ApiMessage>('/sensors/collect'),
  collectByName: (name: string) => post<ApiMessage>(`/sensors/collect/${encodeURIComponent(name)}`),
  disableByName: (name: string) => post<ApiMessage>(`/sensors/disable/${encodeURIComponent(name)}`),
  enableByName: (name: string) => post<ApiMessage>(`/sensors/enable/${encodeURIComponent(name)}`),
}