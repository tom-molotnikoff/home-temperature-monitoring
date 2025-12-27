import useSensorHealthHistory from "../hooks/useSensorHealthHistory.ts";
import type {Sensor, SensorHealthHistory} from "../types/types.ts";
import {type CSSProperties, useMemo, useState} from "react";
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
import {Alert, Button, Snackbar, TextField} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";

interface SensorHealthHistoryChartProps {
  sensor: Sensor,
  limit?: number,
}

function SensorHealthHistoryChart({sensor, limit}: SensorHealthHistoryChartProps) {
  const lineColours = ["#4caf50", "#c62828", "#f9a825"];
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [limitInput, setLimitInput] = useState<string>((limit ?? 5000).toString());
  const [limitState, setLimit] = useState<number>(limit ?? 5000);

  const [healthHistoryData, refresh] = useSensorHealthHistory(sensor.name, limitState);

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

    return sortedByRecordedAt.map((h: SensorHealthHistory) => {
      const recorded = h.recordedAt;
      const status = h.healthStatus;
      const value = mapStatusToValue(status);
      return {
        ...h,
        recordedAt: recorded,
        healthStatus: status,
        healthValue: value,
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
        <>
          <ResponsiveContainer width="100%" height={400}>
            <AreaChart data={mappedData} >
              <CartesianGrid stroke="#eee" strokeDasharray="3 3" />
              <ReferenceArea y1={-0.5} y2={0.5} fill="#fff3cd" fillOpacity={0.6} />
              <ReferenceArea y1={0.5} y2={1.5} fill="#f8d7da" fillOpacity={0.45} />
              <ReferenceArea y1={1.5} y2={2.5} fill="#d4edda" fillOpacity={0.45} />
              <XAxis
                dataKey="recordedAt"
                tickFormatter={(t) => {
                  if (!t) return "";
                  return new Date(t).toLocaleTimeString();
                }}
                tick={{ fontSize: 12 }}
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
              <Line type="step" dataKey="goodVal" stroke={lineColours[0]} dot={true} strokeWidth={4} isAnimationActive={false} name="Good" />
              <Line type="step" dataKey="badVal" stroke={lineColours[1]} dot={true} strokeWidth={4} isAnimationActive={false} name="Bad" />
              <Line type="step" dataKey="unknownVal" stroke={lineColours[2]} dot={true} strokeWidth={4} isAnimationActive={false} name="Unknown" />

              <Legend />
            </AreaChart>
          </ResponsiveContainer>

          <div style={{ display: "flex", justifyContent: "flex-end", width: "100%", gap: 16 }}>
            <TextField
              label="Limit History Entries"
              type="number"
              value={limitInput}
              onChange={(e) => setLimitInput(e.target.value)}
              sx={{ mt: 2, width: 200 }}
            />
            <Button
              onClick={() => {
                const parsed = parseInt(limitInput);
                const isNegative = Number.isFinite(parsed) && parsed < 0;
                if (isNegative) {
                  setLimitInput("5000");
                }
                setLimit(Number.isFinite(parsed) ? parsed : 5000);
                refresh().then(() => {
                  setSnackbarOpen(true);
                });
              }}
              variant="outlined" startIcon={<RefreshIcon />}
              sx={{
                mt: 2,
                alignSelf: 'center',
                height: "56px",
              }}
            >
              Refresh
            </Button>
          </div>
        </>
      )}

      <Snackbar
        open={snackbarOpen}
        onClose={() => setSnackbarOpen(false)}
        autoHideDuration={2000}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert sx={{ width: '100%' }}>
          Sensor health history refreshed.
        </Alert>
      </Snackbar>
    </div>
  );
}

const graphContainerStyle: CSSProperties = {
  width: "100%",
  height: "450px",
};


export default SensorHealthHistoryChart;

