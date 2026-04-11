import type { WidgetProps } from '../types';
import { Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useMeasurementTypes } from '../../hooks/useMeasurementTypes';
import { useReadingsData } from '../../hooks/useReadingsData';
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
import { useChartColours } from '../../theme/chartColours';
import NeedsConfiguration from '../NeedsConfiguration';

export default function ComparisonChartWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const chartColours = useChartColours();
    const measurementType = config.measurementType as string | undefined;
    const { measurementTypes } = useMeasurementTypes();

    const mtInfo = measurementTypes.find(mt => mt.name === measurementType);
    const yAxisLabel = measurementType
        ? {
            value: mtInfo
                ? `${mtInfo.display_name}${mtInfo.unit ? ` (${mtInfo.unit})` : ''}`
                : measurementType.charAt(0).toUpperCase() + measurementType.slice(1),
            angle: -90,
            position: 'insideLeft' as const,
            style: { textAnchor: 'middle' as const, fontSize: 12 },
        }
        : undefined;

    const startDate = typeof config.startDate === 'string' && config.startDate
        ? DateTime.fromISO(config.startDate)
        : DateTime.now().startOf('day');
    const endDate = typeof config.endDate === 'string' && config.endDate
        ? DateTime.fromISO(config.endDate)
        : DateTime.now().plus({ days: 1 }).startOf('day');

    const selectedIds = Array.isArray(config.sensorIds) ? (config.sensorIds as number[]) : [];
    const filteredSensors = selectedIds.length > 0
        ? sensors.filter((s) => selectedIds.includes(s.id))
        : sensors;

    const hourlyAverages = config.useHourlyAverages ? config.useHourlyAverages as boolean : false;

    const chartData = useReadingsData({
        startDate,
        endDate,
        sensors: filteredSensors,
        useHourlyAverages: hourlyAverages,
        measurementType,
    });

    if (!measurementType) {
        return <NeedsConfiguration message="Select a measurement type to compare" />;
    }

    if (filteredSensors.length === 0) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>No sensors available</Typography>;
    }

    if (chartData.length === 0) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>No data for the selected range</Typography>;
    }

    return (
        <div style={{ flex: 1, minHeight: 0, width: '100%' }}>
            <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData}>
                    <CartesianGrid stroke={chartColours.grid} strokeDasharray="3 3" />
                    <XAxis
                        dataKey="time"
                        tickFormatter={(t: string) => new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        minTickGap={50}
                    />
                    <YAxis label={yAxisLabel} />
                    <Tooltip />
                    <Legend />
                    {filteredSensors.map((sensor, index) => (
                        <Line
                            key={sensor.name}
                            type="linear"
                            dataKey={sensor.name}
                            stroke={chartColours.categorical[index % chartColours.categorical.length]}
                            dot={false}
                            connectNulls
                        />
                    ))}
                </LineChart>
            </ResponsiveContainer>
        </div>
    );
}
