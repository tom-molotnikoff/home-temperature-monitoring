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

export type Sensor = {
    name:  string;
    type:  string;
    url:   string;
}