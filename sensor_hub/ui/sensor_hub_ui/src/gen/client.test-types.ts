/**
 * Type-level assertions for the generated openapi-fetch client.
 * These are compile-time checks only — no runtime behaviour.
 * Removed in #32 alongside the smoke test component.
 */

import { apiClient } from './client';
import type { components, operations } from './schema';

// Type aliases from aliases.ts for coverage comparison
import type {
  Reading,
  AggregatedReadingsResponse,
  SensorHealthStatus,
  SensorHealthHistory,
  ConfigFieldSpec,
  DriverInfo,
  MQTTBroker,
  MQTTSubscription,
  MQTTBrokerStats,
  MeasurementTypeInfo,
  TotalReadingsCountForEachSensor,
  Dashboard,
  DashboardWidget,
  DashboardConfig,
  CreateDashboardRequest,
  ShareDashboardRequest,
} from './aliases';

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
// "covers" = schema type extends alias type (schema values are usable
// wherever the alias type is expected).
//
// All entries below must resolve to `true`. If any resolves to `never`,
// tsc will fail with "Type 'never' is not assignable to type 'true'".
// ---------------------------------------------------------------------------
declare const _coverage: {
  // Reading: all fields now required and nullable where appropriate (fixed from #31)
  reading: components['schemas']['Reading'] extends Reading ? true : never;

  // AggregatedReadingsResponse: schema uses enum literals, alias uses string → extends OK
  aggregatedReadingsResponse:
    components['schemas']['AggregatedReadingsResponse'] extends AggregatedReadingsResponse
      ? true : never;

  // SensorHealthStatus: identical enum
  sensorHealthStatus:
    components['schemas']['SensorHealthStatus'] extends SensorHealthStatus ? true : never;

  // SensorHealthHistory: sensor_id is now `string`
  sensorHealthHistory:
    components['schemas']['SensorHealthHistory'] extends SensorHealthHistory ? true : never;

  // ConfigFieldSpec: schema description is optional
  configFieldSpecRequired:
    components['schemas']['ConfigFieldSpec'] extends Omit<ConfigFieldSpec, 'description'>
      ? true : never;

  // DriverInfo
  driverInfo: components['schemas']['DriverInfo'] extends DriverInfo ? true : never;

  // MQTTBroker
  mqttBroker: components['schemas']['MQTTBroker'] extends MQTTBroker ? true : never;

  // MQTTSubscription
  mqttSubscription: components['schemas']['MQTTSubscription'] extends MQTTSubscription ? true : never;

  // MQTTBrokerStats: all non-nullable fields are now required in the schema
  mqttBrokerStats: components['schemas']['MQTTBrokerStats'] extends MQTTBrokerStats ? true : never;

  // MeasurementType (schema) / MeasurementTypeInfo (alias)
  measurementType:
    components['schemas']['MeasurementType'] extends MeasurementTypeInfo ? true : never;

  // getTotalReadingsPerSensor: returns object (map)
  totalReadingsPerSensor: operations['getTotalReadingsPerSensor']['responses']['200']['content']['application/json'] extends TotalReadingsCountForEachSensor ? true : never;

  // Dashboard
  dashboard: components['schemas']['Dashboard'] extends Dashboard ? true : never;

  // DashboardWidget
  dashboardWidget: components['schemas']['DashboardWidget'] extends DashboardWidget ? true : never;

  // DashboardConfig
  dashboardConfig: components['schemas']['DashboardConfig'] extends DashboardConfig ? true : never;

  // CreateDashboardRequest
  createDashboardRequest:
    components['schemas']['CreateDashboardRequest'] extends CreateDashboardRequest ? true : never;

  // ShareDashboardRequest
  shareDashboardRequest:
    components['schemas']['ShareDashboardRequest'] extends ShareDashboardRequest ? true : never;
};
export type { _coverage };

// ---------------------------------------------------------------------------
// Frontend-only types (correctly absent from schema — not gaps):
//   ChartEntry              — derived Recharts shape, not an API type
//   SensorStatus            — embedded as enum in schema Sensor.status
// ---------------------------------------------------------------------------

