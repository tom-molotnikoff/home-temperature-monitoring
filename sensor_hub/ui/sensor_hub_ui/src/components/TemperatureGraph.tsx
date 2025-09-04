import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ResponsiveContainer,
  Legend,
} from "recharts";
import type { ChartEntry, TemperatureReading } from "../types/types";
import React from "react";

const TemperatureGraph = React.memo(function TemperatureGraph({
  readings,
  sensors,
}: {
  readings: TemperatureReading[];
  sensors: string[];
}) {
  const times = Array.from(
    new Set((readings ?? []).map((r) => r.reading.time.replace(" ", "T")))
  );

  const mergedData: ChartEntry[] = times.map((time) => {
    const entry: ChartEntry = { time };
    sensors.forEach((sensor) => {
      const found = readings.find(
        (r) =>
          r.sensor_name === sensor && r.reading.time.replace(" ", "T") === time
      );
      entry[sensor] = found ? found.reading.temperature : null;
    });
    return entry;
  });

  return (
    <div style={{ width: "100%", height: "350px", marginTop: 24 }}>
      {!Array.isArray(readings) || readings.length === 0 ? (
        <p>No readings found for the selected date range.</p>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={mergedData}>
            <CartesianGrid stroke="#eee" />
            <XAxis
              dataKey="time"
              tickFormatter={(t) => new Date(t).toLocaleTimeString()}
            />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line
              type="monotone"
              dataKey="Upstairs"
              stroke="#1976d2"
              strokeWidth={2}
              dot={false}
              connectNulls={true}
            />
            <Line
              type="monotone"
              dataKey="Downstairs"
              strokeWidth={2}
              dot={false}
              stroke="#82ca9d"
              connectNulls={true}
            />
          </LineChart>
        </ResponsiveContainer>
      )}
    </div>
  );
});

export default TemperatureGraph;
