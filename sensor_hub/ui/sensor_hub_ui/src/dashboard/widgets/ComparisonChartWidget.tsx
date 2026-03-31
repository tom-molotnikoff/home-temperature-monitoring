import type { WidgetProps } from '../types';
import { Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useTemperatureData } from '../../hooks/useTemperatureData';
import { DateTime } from 'luxon';
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    Legend,
    ResponsiveContainer,
} from 'recharts';

const LINE_COLORS = ['#1976d2', '#d32f2f', '#4caf50', '#ff9800', '#9c27b0', '#00bcd4', '#795548', '#607d8b'];

export default function ComparisonChartWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const startDate = DateTime.now().minus({ hours: 24 });
    const endDate = DateTime.now();

    const selectedIds = Array.isArray(config.sensorIds) ? (config.sensorIds as number[]) : [];
    const filteredSensors = selectedIds.length > 0
        ? sensors.filter((s) => selectedIds.includes(s.id))
        : sensors;

    const chartData = useTemperatureData({
        startDate,
        endDate,
        sensors: filteredSensors,
        useHourlyAverages: true,
    });

    if (filteredSensors.length === 0) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>No sensors available</Typography>;
    }

    if (chartData.length === 0) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>No data for the last 24 hours</Typography>;
    }

    return (
        <div style={{ flex: 1, minHeight: 0, width: '100%' }}>
            <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                    <CartesianGrid stroke="#eee" strokeDasharray="3 3" />
                    <XAxis
                        dataKey="time"
                        tickFormatter={(t: string) => new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        minTickGap={50}
                    />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    {filteredSensors.map((sensor, index) => (
                        <Line
                            key={sensor.name}
                            type="monotone"
                            dataKey={sensor.name}
                            stroke={LINE_COLORS[index % LINE_COLORS.length]}
                            dot={false}
                            connectNulls
                        />
                    ))}
                </LineChart>
            </ResponsiveContainer>
        </div>
    );
}
