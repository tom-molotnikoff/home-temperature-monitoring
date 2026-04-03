import type { WidgetProps } from '../types';
import { Box, CircularProgress, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useCurrentTemperatures } from '../../hooks/useCurrentTemperatures';

export default function GaugeWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const temperatures = useCurrentTemperatures();

    const sensorId = config.sensorId as number | undefined;
    const min = (config.min as number) ?? 0;
    const max = (config.max as number) ?? 40;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    const reading = temperatures[sensor.name];
    const temp = reading?.temperature ?? null;
    const percentage = temp !== null ? Math.max(0, Math.min(100, ((temp - min) / (max - min)) * 100)) : 0;

    const getColor = (pct: number) => {
        if (pct < 33) return '#1976d2';
        if (pct <= 66) return '#4caf50';
        return '#d32f2f';
    };

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2 }}>
            <Box sx={{ position: 'relative', display: 'inline-flex' }}>
                <CircularProgress
                    variant="determinate"
                    value={temp !== null ? percentage : 0}
                    size={140}
                    thickness={6}
                    sx={{
                        transform: 'rotate(-90deg) !important',
                        color: getColor(percentage),
                    }}
                />
                <Box sx={{ position: 'absolute', top: 0, left: 0, bottom: 0, right: 0, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <Typography variant="h4" sx={{ fontWeight: 'bold' }}>
                        {temp !== null ? `${temp.toFixed(1)}°` : '—'}
                    </Typography>
                </Box>
            </Box>
            <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>
                {sensor.name}
            </Typography>
        </Box>
    );
}
