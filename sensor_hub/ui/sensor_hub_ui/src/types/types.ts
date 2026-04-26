export type Reading = {
    id:               number;
    sensor_name:      string;
    measurement_type: string;
    numeric_value:    number | null;
    text_state:       string | null;
    unit:             string;
    time:             string;
};

export type AggregatedReadingsResponse = {
    aggregation_interval:  string;
    aggregation_function:  string;
    readings:              Reading[];
};

export type PropertiesApiStructure = Record<string, string>;

export type ChartEntry = {
    time:               string;
    [sensor: string]:   number | string | null;
};

export type TotalReadingsCountForEachSensorApiMessage = Record<string, number>;

export type SensorHealthStatus = 'good' | 'bad' | 'unknown';

export type Sensor = {
    id:                       number;
    name:                     string;
    external_id?:             string;
    sensorDriver:             string;
    config:                   Record<string, string>;
    healthStatus:             SensorHealthStatus;
    healthReason:             string;
    enabled:                  boolean;
    status:                   SensorStatus;
    retentionHours:           number | null;
    effectiveRetentionHours?: number;
}

export type SensorHealthHistory = {
    id:            number;
    sensorId:      string;
    healthStatus:  SensorHealthStatus;
    recordedAt:    Date;
}

export type SensorHealthHistoryJson = {
    id:             number;
    sensor_id:      string;
    health_status:  SensorHealthStatus;
    recorded_at:    string;
}

export type SensorJson = {
    id:                         number;
    name:                       string;
    sensor_driver:              string;
    config:                     Record<string, string>;
    health_status:              SensorHealthStatus;
    health_reason:              string;
    enabled:                    boolean;
    status:                     SensorStatus;
    retention_hours?:           number | null;
    effective_retention_hours?: number;
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

// MQTT types

export type MQTTBroker = {
    id:               number;
    name:             string;
    type:             string;
    host:             string;
    port:             number;
    username?:        string;
    password?:        string;
    client_id?:       string;
    enabled:          boolean;
    created_at:       string;
    updated_at:       string;
}

export type MQTTSubscription = {
    id:             number;
    broker_id:      number;
    topic_pattern:  string;
    driver_type:    string;
    enabled:        boolean;
    created_at:     string;
    updated_at:     string;
}

export type MQTTBrokerStats = {
    broker_id:          number;
    broker_name:        string;
    connected:          boolean;
    messages_received:  number;
    parse_errors:       number;
    processing_errors:  number;
    devices_discovered: number;
    last_message_at:    string | null;
    connected_since:    string | null;
}

export type SensorStatus = 'active' | 'pending' | 'dismissed';

export type MeasurementTypeInfo = {
    id:                             number;
    name:                           string;
    display_name:                   string;
    unit:                           string;
    category:                       string; // "numeric" or "binary"
    default_aggregation_function:   string;
    supported_aggregation_functions: string[];
}