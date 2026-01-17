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

export default function WeatherChart({ compact = false }: { compact?: boolean }) {
  const { startDate, endDate } = useContext(DateContext);

  if (!startDate && !endDate) {
    return <div>Loading weather...</div>;
  }

  if (startDate && endDate && startDate > endDate) {
    return <div>Start date cannot be after end date</div>;
  }

  if (startDate && endDate && endDate.diff(startDate, 'days').days > 14) {
    return <div>Cannot render more than 14 days</div>;
  }

  // @ts-expect-error Luxon DateTime type issue
  if ((startDate && startDate.invalid) || (endDate && endDate.invalid)) {
    return <div>Invalid date</div>;
  }

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
      <ResponsiveContainer width="100%" height={compact ? 200 : 300}>
        <LineChart data={data}>
          <CartesianGrid stroke="#eee" />
          <XAxis
            dataKey="time"
            tickFormatter={(t) => {
              const date = new Date(String(t));
              return compact 
                ? date.toLocaleTimeString([], { hour: '2-digit' })
                : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            }}
            interval="preserveStartEnd"
            minTickGap={compact ? 30 : 50}
            tick={{ fontSize: compact ? 10 : 12 }}
            angle={compact ? -45 : 0}
            textAnchor={compact ? 'end' : 'middle'}
            height={compact ? 60 : 30}
          />
          <YAxis tick={{ fontSize: compact ? 10 : 12 }} />
          <Tooltip
            labelFormatter={(t) => new Date(String(t)).toLocaleString()}
          />
          <Legend wrapperStyle={compact ? { fontSize: 10 } : undefined} />
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
