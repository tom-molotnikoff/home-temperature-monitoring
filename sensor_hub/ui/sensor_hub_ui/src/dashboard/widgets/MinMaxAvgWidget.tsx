import type { WidgetProps } from '../types';
import type { TemperatureReading } from '../../types/types';
import { useState, useEffect } from 'react';
import { Box, Paper, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { TemperatureApi } from '../../api/Temperature';
import {DateTime} from "luxon";

export default function MinMaxAvgWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const [stats, setStats] = useState<{ min: number; max: number; avg: number } | null>(null);

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    useEffect(() => {
        if (!sensor) return;

        const start = (config.startDate as string) || DateTime.now().minus({ day: 1 }).toISODate();
        const end = (config.endDate as string) || DateTime.now().plus({ day: 1 }).toISODate();

        TemperatureApi.getBetweenDates(start, end).then((readings: TemperatureReading[]) => {
            const sensorReadings = readings.filter((r) => r.sensor_name === sensor.name);
            if (sensorReadings.length === 0) {
                setStats(null);
                return;
            }
            const temps = sensorReadings.map((r) => r.temperature);
            const min = Math.min(...temps);
            const max = Math.max(...temps);
            const avg = temps.reduce((sum, t) => sum + t, 0) / temps.length;
            setStats({ min, max, avg });
        });
    }, [sensor, config.startDate, config.endDate]);

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    if (!stats) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">No data available</Typography>
            </Box>
        );
    }

    const statItems = [
        { label: 'Min', value: stats.min, color: '#1976d2' },
        { label: 'Avg', value: stats.avg, color: '#757575' },
        { label: 'Max', value: stats.max, color: '#d32f2f' },
    ];

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" sx={{ mb: 1 }}>{sensor.name}</Typography>
            <Box sx={{ display: 'flex', flexDirection: 'row', gap: 2, flex: 1, alignItems: 'center' }}>
                {statItems.map((item) => (
                    <Paper key={item.label} sx={{ flex: 1, p: 2, textAlign: 'center' }} elevation={1}>
                        <Typography variant="caption" sx={{ color: item.color, fontWeight: 'bold' }}>
                            {item.label}
                        </Typography>
                        <Typography variant="h5" sx={{ color: item.color }}>
                            {item.value.toFixed(1)}°
                        </Typography>
                    </Paper>
                ))}
            </Box>
        </Box>
    );
}
