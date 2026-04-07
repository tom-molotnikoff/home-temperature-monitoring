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
    config:       Record<string, string>;
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
    config:         Record<string, string>;
    health_status:  SensorHealthStatus;
    health_reason:  string | null;
    enabled:        boolean;
}

export type ConfigFieldSpec = {
    key:         string;
    label:       string;
    description: string;
    required:    boolean;
    sensitive:   boolean;
    default?:    string;
}

export type DriverInfo = {
    type:                       string;
    display_name:               string;
    description:                string;
    supported_measurement_types: string[];
    config_fields:              ConfigFieldSpec[];
}