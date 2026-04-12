import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useCurrentReadings } from '../../hooks/useCurrentReadings';

export default function CurrentReadingWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const readings = useCurrentReadings();

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    if (!sensor) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Configure sensor</Typography>
            </Box>
        );
    }

    const sensorReadings = readings[sensor.name];
    const reading = sensorReadings ? Object.values(sensorReadings)[0] : undefined;

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" color="text.secondary">{sensor.name}</Typography>
            <Typography variant="h1" sx={{ fontSize: '4rem', fontWeight: 'bold', textAlign: 'center' }}>
                {reading ? `${reading.numeric_value?.toFixed(1)}${reading.unit ? ` ${reading.unit}` : ''}` : '—'}
            </Typography>
            {reading && (
                <Typography variant="caption" color="text.secondary">
                    {new Date(reading.time).toLocaleString()}
                </Typography>
            )}
        </Box>
    );
}
