import type { WidgetProps } from '../types';
import type { TemperatureReading } from '../../types/types';
import { useState, useEffect } from 'react';
import { Box, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { TemperatureApi } from '../../api/Temperature';

function tempToColor(temp: number): string {
    const cold = 10;
    const hot = 30;
    const ratio = Math.max(0, Math.min(1, (temp - cold) / (hot - cold)));

    if (ratio <= 0.5) {
        const t = ratio * 2;
        const r = Math.round(0 + t * 0);
        const g = Math.round(100 + t * 155);
        const b = Math.round(255 - t * 100);
        return `rgb(${r},${g},${b})`;
    }
    const t = (ratio - 0.5) * 2;
    const r = Math.round(0 + t * 220);
    const g = Math.round(200 - t * 200);
    const b = Math.round(50 - t * 50);
    return `rgb(${r},${g},${b})`;
}

interface DayData {
    day: number;
    avg: number | null;
}

export default function HeatmapWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const [days, setDays] = useState<DayData[]>([]);

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    useEffect(() => {
        if (!sensor) return;

        const now = new Date();
        const start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);

        TemperatureApi.getBetweenDates(start.toISOString(), now.toISOString()).then((readings: TemperatureReading[]) => {
            const sensorReadings = readings.filter((r) => r.sensor_name === sensor.name);
            const grouped: Record<string, number[]> = {};

            for (const r of sensorReadings) {
                const dateKey = new Date(r.time).toISOString().slice(0, 10);
                if (!grouped[dateKey]) grouped[dateKey] = [];
                grouped[dateKey].push(r.temperature);
            }

            const result: DayData[] = [];
            for (let i = 29; i >= 0; i--) {
                const d = new Date(now.getTime() - i * 24 * 60 * 60 * 1000);
                const key = d.toISOString().slice(0, 10);
                const temps = grouped[key];
                result.push({
                    day: d.getDate(),
                    avg: temps ? temps.reduce((s, t) => s + t, 0) / temps.length : null,
                });
            }
            setDays(result);
        });
    }, [sensor]);

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 2 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>{sensor.name} — Last 30 Days</Typography>
            <Box
                sx={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(7, 1fr)',
                    gap: 0.5,
                    flex: 1,
                    alignContent: 'start',
                }}
            >
                {days.map((d, i) => (
                    <Box
                        key={i}
                        sx={{
                            aspectRatio: '1',
                            borderRadius: 0.5,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            backgroundColor: d.avg !== null ? tempToColor(d.avg) : '#e0e0e0',
                            color: d.avg !== null ? '#fff' : '#999',
                            fontSize: '0.7rem',
                            fontWeight: 'bold',
                        }}
                    >
                        {d.day}
                    </Box>
                ))}
            </Box>
        </Box>
    );
}
