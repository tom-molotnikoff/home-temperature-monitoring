export type Reading = {
    id:               number;
    sensor_name:      string;
    measurement_type: string;
    numeric_value:    number | null;
    text_state:       string | null;
    unit:             string;
    time:             string;
};

export type PropertiesApiStructure = Record<string, string>;

export const SensorDrivers = [
  "sensor-hub-http-temperature",
] as const;
export type SensorDriver = typeof SensorDrivers[number];

export type ChartEntry = {
    time:               string;
    [sensor: string]:   number | string | null;
};

export type TotalReadingsCountForEachSensorApiMessage = Record<string, number>;

export type SensorHealthStatus = 'good' | 'bad' | 'unknown';

export type Sensor = {
    id:           number;
    name:         string;
    sensorDriver: string;
    url:          string;
    healthStatus: SensorHealthStatus;
    healthReason: string | null;
    enabled:      boolean;
}

export type SensorHealthHistory = {
    id:            number;
    sensorId:      number;
    healthStatus:  SensorHealthStatus;
    recordedAt:    Date;
}

export type SensorHealthHistoryJson = {
    id:             number;
    sensor_id:      number;
    health_status:  SensorHealthStatus;
    recorded_at:    string;
}

export type SensorJson = {
    id:             number;
    name:           string;
    sensor_driver:  string;
    url:            string;
    health_status:  SensorHealthStatus;
    health_reason:  string | null;
    enabled:        boolean;
}