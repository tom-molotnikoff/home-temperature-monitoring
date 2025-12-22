export type TemperatureReading = {
    id:             number;
    sensor_name:    string;
    temperature:    number;
    time:           string;
};

export type PropertiesApiStructure = Record<string, string>;

export const SensorTypes = [
  "Temperature",
]

export type ChartEntry = {
    time:               string;
    [sensor: string]:   number | string | null;
};

export type SensorHealthStatus = 'good' | 'bad' | 'unknown';

export type Sensor = {
    id:           number;
    name:         string;
    type:         string;
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
    type:           string;
    url:            string;
    health_status:  SensorHealthStatus;
    health_reason:  string | null;
    enabled:        boolean;
}