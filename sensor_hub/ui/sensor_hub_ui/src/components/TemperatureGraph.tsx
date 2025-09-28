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
import React, { useContext, type CSSProperties } from "react";
import { DateContext } from "../providers/DateContext";
import { useTemperatureData } from "../hooks/useTemperatureData";

const TemperatureGraph = React.memo(function TemperatureGraph({
  sensors,
  useHourlyAverages,
}: {
  sensors: string[];
  useHourlyAverages: boolean;
}) {
  const { startDate, endDate } = useContext(DateContext);

  const chartData = useTemperatureData({
    startDate: startDate ? startDate : null,
    endDate: endDate ? endDate : null,
    sensors,
    useHourlyAverages,
  });

  return (
    <div style={graphContainerStyle}>
      {!Array.isArray(chartData) || chartData.length === 0 ? (
        <></>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
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

const graphContainerStyle: CSSProperties = {
  width: "100%",
  height: "350px",
  marginTop: 24,
};

export default TemperatureGraph;
