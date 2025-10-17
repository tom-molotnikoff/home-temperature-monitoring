export type TemperatureReading = {
    id:             number;
    sensor_name:    string;
    temperature:    number;
    time:           string;
};

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
}

export type SensorJson = {
    id:             number;
    name:           string;
    type:           string;
    url:            string;
    health_status:  SensorHealthStatus;
    health_reason:  string | null;
}