import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  CartesianGrid,
  ResponsiveContainer,
} from "recharts";
import useWeatherApi from "../hooks/useWeatherApi";
import { useContext, type CSSProperties } from "react";
import { DateContext } from "../providers/DateContext";

export default function WeatherChart() {
  const { startDate, endDate } = useContext(DateContext);

  const opts =
    startDate || endDate
      ? {
          hourly: ["temperature_2m", "apparent_temperature", "uv_index"],
          startDate: startDate ? startDate.toJSDate() : null,
          endDate: endDate ? endDate.toJSDate() : null,
        }
      : {
          hourly: ["temperature_2m", "apparent_temperature", "uv_index"],
          days: 7,
        };

  const { data, loading, error } = useWeatherApi(53.383, -1.4659, opts);

  if (loading) return <div>Loading weather...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div style={containerStyle}>
      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={data}>
          <CartesianGrid stroke="#eee" />
          <XAxis
            dataKey="time"
            tickFormatter={(t) => new Date(String(t)).toLocaleTimeString()}
          />
          <YAxis />
          <Tooltip
            labelFormatter={(t) => new Date(String(t)).toLocaleString()}
          />
          <Legend />
          <Line
            type="monotone"
            dataKey="temperature_2m"
            stroke="#ff7300"
            dot={false}
          />
          <Line
            type="monotone"
            dataKey="apparent_temperature"
            stroke="#387908"
            dot={false}
          />
          <Line
            type="monotone"
            dataKey="uv_index"
            stroke="#2018bcff"
            dot={false}
            yAxisId={1}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

const containerStyle: CSSProperties = {
  width: "100%",
  margin: "0 auto",
};
