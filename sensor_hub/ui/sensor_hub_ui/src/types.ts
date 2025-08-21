export type TemperatureReading = {
  sensor_name: string;
  reading: {
    temperature: number;
    time: string;
  };
};

export type ChartEntry = {
  time: string;
  [sensor: string]: number | string | null;
};