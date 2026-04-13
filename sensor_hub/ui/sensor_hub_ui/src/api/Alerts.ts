import { get, post, put, del, type ApiMessage } from './Client';

export type AlertRule = {
  ID: number;
  SensorID: number;
  SensorName: string;
  MeasurementTypeID: number;
  MeasurementType: string;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold: number | null;
  LowThreshold: number | null;
  TriggerStatus: string;
  Enabled: boolean;
  RateLimitSeconds: number;
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
  MeasurementTypeID: number;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitSeconds: number;
  Enabled: boolean;
};

export type UpdateAlertRuleRequest = {
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitSeconds: number;
  Enabled: boolean;
};

export const listAlertRules = () => get<AlertRule[]>('/alerts');

export const getAlertRule = (ruleId: number) => get<AlertRule>(`/alerts/${ruleId}`);

export const getAlertRulesBySensorId = (sensorId: number) => get<AlertRule[]>(`/alerts/sensor/${sensorId}`);

export const createAlertRule = (rule: CreateAlertRuleRequest) => post<ApiMessage>('/alerts', rule);

export const updateAlertRule = (ruleId: number, rule: UpdateAlertRuleRequest) => put<ApiMessage>(`/alerts/${ruleId}`, rule);

export const deleteAlertRule = (ruleId: number) => del<ApiMessage>(`/alerts/${ruleId}`);

export const getAlertHistory = (sensorId: number, limit?: number) => {
  const params = limit ? `?limit=${limit}` : '';
  return get<AlertHistory[]>(`/alerts/sensor/${sensorId}/history${params}`);
};
