import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useCurrentTemperatures } from '../../hooks/useCurrentTemperatures';

export default function CurrentReadingWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const temperatures = useCurrentTemperatures();

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    const reading = temperatures[sensor.name];

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" color="text.secondary">{sensor.name}</Typography>
            <Typography variant="h1" sx={{ fontSize: '4rem', fontWeight: 'bold', textAlign: 'center' }}>
                {reading ? `${reading.temperature.toFixed(1)}°` : '—'}
            </Typography>
            {reading && (
                <Typography variant="caption" color="text.secondary">
                    {new Date(reading.time).toLocaleString()}
                </Typography>
            )}
        </Box>
    );
}
