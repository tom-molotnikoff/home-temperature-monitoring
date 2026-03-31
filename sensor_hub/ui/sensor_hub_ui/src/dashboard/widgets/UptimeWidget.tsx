import type { WidgetProps } from '../types';
import { Box, LinearProgress, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import useSensorHealthHistory from '../../hooks/useSensorHealthHistory';

export default function UptimeWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;
    const sensorName = sensor?.name ?? '';

    const [history] = useSensorHealthHistory(sensorName, 1000);

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    const total = history.length;
    const good = history.filter((h) => h.healthStatus === 'good').length;
    const uptime = total > 0 ? (good / total) * 100 : 0;

    const getColor = (pct: number): 'success' | 'warning' | 'error' => {
        if (pct > 90) return 'success';
        if (pct >= 70) return 'warning';
        return 'error';
    };

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2, gap: 2 }}>
            <Typography variant="h3" sx={{ fontWeight: 'bold' }}>
                {total > 0 ? `${uptime.toFixed(1)}%` : '—'}
            </Typography>
            <Box sx={{ width: '80%' }}>
                <LinearProgress
                    variant="determinate"
                    value={uptime}
                    color={getColor(uptime)}
                    sx={{ height: 10, borderRadius: 5 }}
                />
            </Box>
            <Typography variant="subtitle2" color="text.secondary">{sensor.name}</Typography>
        </Box>
    );
}
