import { get, post, put, del, type ApiMessage } from './Client';

export type AlertRule = {
  ID: number;
  SensorID: number;
  SensorName: string;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold: number | null;
  LowThreshold: number | null;
  TriggerStatus: string;
  Enabled: boolean;
  RateLimitHours: number;
  LastAlertSentAt: string | null;
};

export type AlertHistory = {
  id: number;
  sensor_id: number;
  alert_type: string;
  reading_value: string;
  sent_at: string;
};

export type CreateAlertRuleRequest = {
  SensorID: number;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitHours: number;
  Enabled: boolean;
};

export type UpdateAlertRuleRequest = {
  SensorID: number;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitHours: number;
  Enabled: boolean;
};

export const listAlertRules = () => get<AlertRule[]>('/alerts/');

export const getAlertRule = (sensorId: number) => get<AlertRule>(`/alerts/${sensorId}`);

export const createAlertRule = (rule: CreateAlertRuleRequest) => post<ApiMessage>('/alerts/', rule);

export const updateAlertRule = (sensorId: number, rule: UpdateAlertRuleRequest) => put<ApiMessage>(`/alerts/${sensorId}`, rule);

export const deleteAlertRule = (sensorId: number) => del<ApiMessage>(`/alerts/${sensorId}`);

export const getAlertHistory = (sensorId: number, limit?: number) => {
  const params = limit ? `?limit=${limit}` : '';
  return get<AlertHistory[]>(`/alerts/${sensorId}/history${params}`);
};
