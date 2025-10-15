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
  useContext,
  useEffect,
  useReducer,
  type CSSProperties,
} from "react";
import { DateContext } from "../providers/DateContext";
import { useTemperatureData } from "../hooks/useTemperatureData";
import { linesHiddenReducer } from "../reducers/LinesHiddenReducer";
import type {Sensor} from "../types/types.ts";

const TemperatureGraph = React.memo(function TemperatureGraph({
  sensors,
  useHourlyAverages,
}: {
  sensors: Sensor[];
  useHourlyAverages: boolean;
}) {
  const { startDate, endDate } = useContext(DateContext);

  const lineColours = ["#1976d2", "#82ca9d", "#fffb00ff", "#db5f5fff"];

  const [linesHidden, setLinesHidden] = useReducer(linesHiddenReducer, {});

  useEffect(() => {
    if (Object.keys(linesHidden).length !== 0) return;
    sensors.forEach((sensor) => {
      setLinesHidden({ type: "reset", key: sensor.name });
    });
  }, [sensors, linesHidden]);

  const chartData = useTemperatureData({
    startDate: startDate ? startDate : null,
    endDate: endDate ? endDate : null,
    sensors,
    useHourlyAverages,
  });

  const legendClickHandler = (data: LegendPayload) => {
    setLinesHidden({ type: "toggle", key: data.dataKey as string });
  };

  return (
    <div data-testid="temperature-graph" style={graphContainerStyle}>
      {!Array.isArray(chartData) || chartData.length === 0 ? (
        <></>
      ) : (
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <CartesianGrid stroke="#eee" strokeDasharray="3 3" />
            <XAxis
              dataKey="time"
              tickFormatter={(t) => new Date(t).toLocaleTimeString()}
            />
            <YAxis type="number" domain={[12, 26]} />
            <Tooltip />
            <Legend onClick={legendClickHandler} />
            {sensors.map((sensor, index) => (
              <Line
                key={sensor.name}
                type="natural"
                dataKey={sensor.name}
                stroke={lineColours[index]}
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
      )}
    </div>
  );
});

const graphContainerStyle: CSSProperties = {
  width: "100%",
  height: "350px",
};

export default TemperatureGraph;
