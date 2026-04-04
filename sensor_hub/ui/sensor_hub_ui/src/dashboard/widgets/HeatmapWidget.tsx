import type { WidgetProps } from '../types';
import type { TemperatureReading } from '../../types/types';
import { useState, useEffect, useRef, useCallback } from 'react';
import { Box, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { TemperatureApi } from '../../api/Temperature';
import { useIsDark } from '../../theme/useIsDark';

function tempToColor(temp: number, cold: number, hot: number): string {
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
    const isDark = useIsDark();
    const [days, setDays] = useState<DayData[]>([]);
    const [cellSize, setCellSize] = useState(28);
    const gridRef = useRef<HTMLDivElement>(null);

    const cold = typeof config.tempMin === 'number' ? config.tempMin : 10;
    const hot = typeof config.tempMax === 'number' ? config.tempMax : 30;
    const noDataColor = isDark ? '#333333' : '#E0D8D0';
    const noDataTextColor = isDark ? '#A0A0A0' : '#5C5C5C';

    const cols = 7;
    const rows = Math.ceil(days.length / cols) || 1;
    const gap = 4;

    const recalc = useCallback(() => {
        const el = gridRef.current;
        if (!el) return;
        const { width, height } = el.getBoundingClientRect();
        const maxFromW = (width - (cols - 1) * gap) / cols;
        const maxFromH = (height - (rows - 1) * gap) / rows;
        setCellSize(Math.max(12, Math.floor(Math.min(maxFromW, maxFromH))));
    }, [rows]);

    useEffect(() => {
        const el = gridRef.current;
        if (!el) return;
        const observer = new ResizeObserver(recalc);
        observer.observe(el);
        return () => observer.disconnect();
    }, [recalc]);

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    useEffect(() => {
        if (!sensor) return;

        const now = new Date();
        const start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);

        TemperatureApi.getBetweenDates(start.toISOString().slice(0, 10), now.toISOString().slice(0, 10)).then((readings: TemperatureReading[]) => {
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

    const firstMonth = days.length > 0 ? new Date(new Date().getTime() - 29 * 24 * 60 * 60 * 1000).toLocaleString('default', { month: 'long' }) : '';
    const lastMonth = new Date().toLocaleString('default', { month: 'long' });
    const monthLabel = firstMonth === lastMonth ? firstMonth : `${firstMonth} → ${lastMonth}`;

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 1, overflow: 'hidden' }}>
            <Typography variant="caption" color="text.secondary" sx={{ mb: 0.5, textAlign: 'center', flexShrink: 0 }}>{monthLabel}</Typography>
            <Box
                ref={gridRef}
                sx={{
                    flex: 1,
                    minHeight: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                }}
            >
                <Box
                    sx={{
                        display: 'grid',
                        gridTemplateColumns: `repeat(${cols}, ${cellSize}px)`,
                        gap: `${gap}px`,
                    }}
                >
                    {days.map((d, i) => (
                        <Box
                            key={i}
                            sx={{
                                width: cellSize,
                                height: cellSize,
                                borderRadius: 0.5,
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                backgroundColor: d.avg !== null ? tempToColor(d.avg, cold, hot) : noDataColor,
                                color: d.avg !== null ? '#fff' : noDataTextColor,
                                fontSize: Math.max(9, cellSize * 0.35),
                                fontWeight: 'bold',
                            }}
                        >
                            {d.day}
                        </Box>
                    ))}
                </Box>
            </Box>
        </Box>
    );
}
