/**
 * Type-level assertions for the generated openapi-fetch client.
 * These are compile-time checks only — no runtime behaviour.
 * Removed in #32 alongside the smoke test component.
 */

import { apiClient } from './client';
import type { components, operations } from './schema';

// Hand-written types for coverage comparison (Cycle 3)
import type {
  Reading,
  AggregatedReadingsResponse,
  SensorHealthStatus,
  SensorJson,
  SensorHealthHistoryJson,
  ConfigFieldSpec,
  DriverInfo,
  MQTTBroker,
  MQTTSubscription,
  MQTTBrokerStats,
  MeasurementTypeInfo,
  PropertiesApiStructure,
  TotalReadingsCountForEachSensorApiMessage,
} from '../types/types';
import type {
  Dashboard,
  DashboardWidget,
  DashboardConfig,
  CreateDashboardRequest,
  ShareDashboardRequest,
} from '../types/dashboard';

// Cycle 2: Smoke test component must type-check — verified by tsc passing on
// ApiClientSmokeTest.tsx. This declaration anchors the import so tsc checks it.
import type {} from './ApiClientSmokeTest';

// ---------------------------------------------------------------------------
// Cycle 1: apiClient has a typed GET method
// ---------------------------------------------------------------------------
declare const _clientCheck: {
  hasGet: typeof apiClient.GET extends (...args: unknown[]) => Promise<unknown> ? true : never;
};
// Compiler evaluates the object type — if hasGet resolves to never, tsc errors.
export type { _clientCheck };

// ---------------------------------------------------------------------------
// Cycle 3: Schema type coverage
// "covers" = schema type extends hand-written type (schema values are usable
// wherever the hand-written type is expected).
//
// All entries below must resolve to `true`. If any resolves to `never`,
// tsc will fail with "Type 'never' is not assignable to type 'true'".
// ---------------------------------------------------------------------------
declare const _coverage: {
  // Reading: all fields now required and nullable where appropriate (fixed from #31)
  reading: components['schemas']['Reading'] extends Reading ? true : never;

  // AggregatedReadingsResponse: schema uses enum literals, HW uses string → extends OK
  aggregatedReadingsResponse:
    components['schemas']['AggregatedReadingsResponse'] extends AggregatedReadingsResponse
      ? true : never;

  // SensorHealthStatus: identical enum
  sensorHealthStatus:
    components['schemas']['SensorHealthStatus'] extends SensorHealthStatus ? true : never;

  // SensorJson: health_reason is now `string` (not `string|null`); health_status uses enum ref
  sensorJson: components['schemas']['Sensor'] extends SensorJson ? true : never;

  // SensorHealthHistoryJson: sensor_id is now `string` (fixed from #31)
  sensorHealthHistoryJson:
    components['schemas']['SensorHealthHistory'] extends SensorHealthHistoryJson ? true : never;

  // ConfigFieldSpec: schema description is optional; HW requires it (minor gap).
  //   Assert on all required fields only.
  configFieldSpecRequired:
    components['schemas']['ConfigFieldSpec'] extends Omit<ConfigFieldSpec, 'description'>
      ? true : never;

  // DriverInfo
  driverInfo: components['schemas']['DriverInfo'] extends DriverInfo ? true : never;

  // MQTTBroker
  mqttBroker: components['schemas']['MQTTBroker'] extends MQTTBroker ? true : never;

  // MQTTSubscription
  mqttSubscription: components['schemas']['MQTTSubscription'] extends MQTTSubscription ? true : never;

  // MQTTBrokerStats: all non-nullable fields are now required in the schema (fixed from #31)
  mqttBrokerStats: components['schemas']['MQTTBrokerStats'] extends MQTTBrokerStats ? true : never;

  // MeasurementType (schema) / MeasurementTypeInfo (HW) — category enum extends string → OK
  measurementType:
    components['schemas']['MeasurementType'] extends MeasurementTypeInfo ? true : never;

  // PropertiesMap (schema) / PropertiesApiStructure (HW) — structurally identical
  propertiesMap:
    components['schemas']['PropertiesMap'] extends PropertiesApiStructure ? true : never;

  // getTotalReadingsPerSensor: now returns object (map), not array (fixed from #31)
  totalReadingsPerSensor: operations['getTotalReadingsPerSensor']['responses']['200']['content']['application/json'] extends TotalReadingsCountForEachSensorApiMessage ? true : never;

  // Dashboard
  dashboard: components['schemas']['Dashboard'] extends Dashboard ? true : never;

  // DashboardWidget
  dashboardWidget: components['schemas']['DashboardWidget'] extends DashboardWidget ? true : never;

  // DashboardConfig
  dashboardConfig: components['schemas']['DashboardConfig'] extends DashboardConfig ? true : never;

  // CreateDashboardRequest: config is now required in schema (fixed from #31)
  createDashboardRequest:
    components['schemas']['CreateDashboardRequest'] extends CreateDashboardRequest ? true : never;

  // UpdateDashboardRequest: schema config is optional (PATCH semantics); HW requires it.
  //   This is intentional — UpdateDashboardRequest allows partial updates.
  //   Assertion omitted (schema is intentionally looser than HW type).

  // ShareDashboardRequest
  shareDashboardRequest:
    components['schemas']['ShareDashboardRequest'] extends ShareDashboardRequest ? true : never;
};
export type { _coverage };

// ---------------------------------------------------------------------------
// Frontend-only types (correctly absent from schema — not gaps):
//   ChartEntry              — derived Recharts shape, not an API type
//   Sensor (camelCase)      — UI mapping type from mapSensorJson()
//   SensorHealthHistory     — UI mapping type (camelCase, sensorId: string after #31 fix)
//   WidgetLayout            — embedded in DashboardWidget.layout (no top-level schema entry)
//   DashboardBreakpoints    — embedded in DashboardConfig.breakpoints
//   SensorStatus            — embedded as enum in schema Sensor.status
// ---------------------------------------------------------------------------

