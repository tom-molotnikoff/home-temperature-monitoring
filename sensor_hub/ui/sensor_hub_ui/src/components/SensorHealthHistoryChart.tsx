import useSensorHealthHistory from "../hooks/useSensorHealthHistory.ts";
import type {Sensor, SensorHealthHistory} from "../types/types.ts";
import {type CSSProperties, useMemo} from "react";
import {
  CartesianGrid,
  Legend,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  Area,
  AreaChart,
  ReferenceArea,
} from "recharts";
import { useIsMobile } from "../hooks/useMobile";
import { useChartColours } from "../theme/chartColours";

// Custom dot that only renders at transition points for lines with valid values
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function TransitionDot(props: any) {
  const { cx, cy, payload, stroke, value } = props;
  // Only render if this is a transition AND this line has a value (not null)
  if (!payload?.isTransition || value === null) return null;
  return <circle cx={cx} cy={cy} r={4} fill={stroke} stroke={stroke} />;
}

interface SensorHealthHistoryChartProps {
  sensor: Sensor,
  limit?: number,
}

function SensorHealthHistoryChart({sensor, limit}: SensorHealthHistoryChartProps) {
  const chartColours = useChartColours();
  const isMobile = useIsMobile();

  const [healthHistoryData] = useSensorHealthHistory(sensor.name, limit ?? 1000);

  const mappedData = useMemo(() => {
    if (!Array.isArray(healthHistoryData)) return [];

    const sortedByRecordedAt = [...healthHistoryData].sort((a, b) => {
      const dateA = new Date(a.recordedAt).getTime();
      const dateB = new Date(b.recordedAt).getTime();
      return dateA - dateB;
    });

    const mapStatusToValue = (s: string | undefined | null) => {
      if (!s) return 0;
      const lower = s.toString().toLowerCase();
      if (lower === "good") return 2;
      if (lower === "bad") return 1;
      if (lower === "unknown") return 0;
      return 0;
    };

    return sortedByRecordedAt.map((h: SensorHealthHistory, index: number) => {
      const recorded = h.recordedAt;
      const status = h.healthStatus;
      const value = mapStatusToValue(status);
      const prevValue = index > 0 ? mapStatusToValue(sortedByRecordedAt[index - 1].healthStatus) : null;
      const isTransition = prevValue === null || prevValue !== value;
      return {
        ...h,
        recordedAt: recorded,
        healthStatus: status,
        healthValue: value,
        isTransition,
        // per-state series used to draw colored segments (null when not active so Recharts doesn't connect)
        goodVal: value === 2 ? 2 : null,
        badVal: value === 1 ? 1 : null,
        unknownVal: value === 0 ? 0 : null,
      };
    });
  }, [healthHistoryData]);

  const valueToLabel = (v: number) => {
    if (v === 2) return "good";
    if (v === 1) return "bad";
    return "unknown";
  };

  return (
    <div data-testid="sensor-health-history-chart" style={graphContainerStyle}>
      {!Array.isArray(mappedData) || mappedData.length === 0 ? (
        <></>
      ) : (
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0 }}>
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={mappedData} >
              <CartesianGrid stroke={chartColours.grid} strokeDasharray="3 3" />
              <ReferenceArea y1={-0.5} y2={0.5} fill={chartColours.health[2]} fillOpacity={0.15} />
              <ReferenceArea y1={0.5} y2={1.5} fill={chartColours.health[1]} fillOpacity={0.15} />
              <ReferenceArea y1={1.5} y2={2.5} fill={chartColours.health[0]} fillOpacity={0.15} />
              <XAxis
                dataKey="recordedAt"
                tickFormatter={(t) => {
                  if (!t) return "";
                  const date = new Date(t);
                  return isMobile 
                    ? date.toLocaleTimeString([], { hour: '2-digit' })
                    : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
                }}
                interval="preserveStartEnd"
                minTickGap={isMobile ? 30 : 50}
                tick={{ fontSize: isMobile ? 10 : 12 }}
                angle={isMobile ? -45 : 0}
                textAnchor={isMobile ? 'end' : 'middle'}
                height={isMobile ? 60 : 30}
              />
              <YAxis
                type="number"
                dataKey="healthValue"
                domain={[-0.5, 2.5]}
                ticks={[0, 1, 2]}
                tickFormatter={(v) => valueToLabel(Number(v))}
                allowDataOverflow={false}
                width={80}
              />
              <Tooltip
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                formatter={(value: any, name: any) => {
                  if (name === 'healthValue') return [valueToLabel(Number(value)), 'Health'];
                  return [value, name];
                }}
                labelFormatter={(label) => {
                  if (!label) return '';
                  return new Date(label).toLocaleString();
                }}
              />
              <Area
                type="step"
                dataKey="healthValue"
                stroke="transparent"
                fill="transparent"
                strokeWidth={2}
                dot={false}
                isAnimationActive={false}
              />

              {/* Colored step lines per-state — only present where that state is active */}
              <Line type="step" dataKey="goodVal" stroke={chartColours.health[0]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Good" />
              <Line type="step" dataKey="badVal" stroke={chartColours.health[1]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Bad" />
              <Line type="step" dataKey="unknownVal" stroke={chartColours.health[2]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Unknown" />

              <Legend />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
}

const graphContainerStyle: CSSProperties = {
  width: "100%",
  flex: 1,
  minHeight: 0,
  position: "relative",
};


export default SensorHealthHistoryChart;

