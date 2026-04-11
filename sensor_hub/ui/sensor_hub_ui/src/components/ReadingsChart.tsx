import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  ResponsiveContainer,
  Legend,
  type LegendPayload,
} from "recharts";
import React, {
  useEffect,
  useMemo,
  useReducer,
  type CSSProperties,
} from "react";
import { useReadingsData } from "../hooks/useReadingsData";
import { linesHiddenReducer } from "../reducers/LinesHiddenReducer";
import type {Sensor} from "../types/types.ts";
import type { DateTime } from "luxon";
import EmptyState from "./EmptyState";
import ShowChartOutlinedIcon from "@mui/icons-material/ShowChartOutlined";
import { useChartColours } from "../theme/chartColours";

const ReadingsChart = React.memo(function ReadingsChart({
  sensors,
  useHourlyAverages,
  startDate,
  endDate,
  compact = false,
  measurementType,
}: {
  sensors: Sensor[];
  useHourlyAverages: boolean;
  startDate: DateTime | null;
  endDate: DateTime | null;
  compact?: boolean;
  measurementType?: string;
}) {

  const chartColours = useChartColours();

  const [linesHidden, setLinesHidden] = useReducer(linesHiddenReducer, {});

  const chartData = useReadingsData({
    startDate: startDate ? startDate : null,
    endDate: endDate ? endDate : null,
    sensors,
    useHourlyAverages,
    measurementType,
  });

  // Only include sensors that have at least one non-null data point
  const activeSensors = useMemo(() => {
    if (!Array.isArray(chartData) || chartData.length === 0) return [];
    return sensors.filter((s) =>
      chartData.some((entry) => entry[s.name] != null),
    );
  }, [chartData, sensors]);

  useEffect(() => {
    if (Object.keys(linesHidden).length !== 0) return;
    activeSensors.forEach((sensor) => {
      setLinesHidden({ type: "reset", key: sensor.name });
    });
  }, [activeSensors, linesHidden]);

  const legendClickHandler = (data: LegendPayload) => {
    setLinesHidden({ type: "toggle", key: data.dataKey as string });
  };

  const yAxisLabel = measurementType
    ? { value: measurementType.charAt(0).toUpperCase() + measurementType.slice(1), angle: -90, position: 'insideLeft' as const, style: { textAnchor: 'middle' as const, fontSize: compact ? 10 : 12 } }
    : undefined;

  return (
    <div data-testid="readings-chart" style={{ ...graphContainerStyle, flex: 1, minHeight: 0, position: 'relative' }}>
      {sensors.length === 0 ? (
        <EmptyState
          icon={<ShowChartOutlinedIcon sx={{ fontSize: 48 }} />}
          title="No sensors configured"
          description="Add a sensor to start seeing data here."
          actionLabel="Add a sensor"
          actionHref="/sensors-overview"
          minHeight={200}
        />
      ) : !Array.isArray(chartData) || chartData.length === 0 ? (
        <EmptyState
          icon={<ShowChartOutlinedIcon sx={{ fontSize: 48 }} />}
          title="No readings in selected date range"
          description="Try adjusting the date range or wait for new readings."
          minHeight={200}
        />
      ) : activeSensors.length === 0 ? (
        <EmptyState
          icon={<ShowChartOutlinedIcon sx={{ fontSize: 48 }} />}
          title="No sensors have this reading type"
          description={measurementType
            ? `None of the available sensors report "${measurementType}" readings.`
            : "No matching sensor data found."}
          minHeight={200}
        />
      ) : (
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid stroke={chartColours.grid} strokeDasharray="3 3" />
              <XAxis
                dataKey="time"
                tickFormatter={(t) => {
                  const date = new Date(t);
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
              <YAxis type="number" domain={['auto', 'auto']} tick={{ fontSize: compact ? 10 : 12 }} label={yAxisLabel} />
              <Tooltip />
              <Legend 
                onClick={legendClickHandler}
                wrapperStyle={compact ? { fontSize: 10 } : undefined}
              />
              {activeSensors.map((sensor, index) => (
                <Line
                  key={sensor.name}
                  type="linear"
                  dataKey={sensor.name}
                  stroke={chartColours.categorical[index % chartColours.categorical.length]}
                  dot={false}
                  connectNulls={true}
                  animationEasing="ease-in-out"
                  animationDuration={800}
                  hide={linesHidden[sensor.name]}
                  legendType="plainline"
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
});

const graphContainerStyle: CSSProperties = {
  width: "100%",
};

export default ReadingsChart;
