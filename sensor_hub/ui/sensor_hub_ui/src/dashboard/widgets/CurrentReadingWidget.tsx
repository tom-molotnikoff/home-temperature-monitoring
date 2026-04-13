import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useCurrentReadings } from '../../hooks/useCurrentReadings';
import NeedsConfiguration from '../NeedsConfiguration';

export default function CurrentReadingWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const readings = useCurrentReadings();

    const sensorId = config.sensorId as number | undefined;
    const measurementType = config.measurementType as string | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    if (!sensor || !measurementType) {
        return <NeedsConfiguration message="Select a sensor and measurement type" />;
    }

    const reading = readings[sensor.name]?.[measurementType];

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" color="text.secondary">{sensor.name}</Typography>
            <Typography variant="h1" sx={{ fontSize: '4rem', fontWeight: 'bold', textAlign: 'center' }}>
                {reading
                    ? reading.numeric_value != null
                        ? `${reading.numeric_value.toFixed(1)}${reading.unit ? ` ${reading.unit}` : ''}`
                        : reading.text_state ?? '—'
                    : '—'}
            </Typography>
            {reading && (
                <Typography variant="caption" color="text.secondary">
                    {new Date(reading.time).toLocaleString()}
                </Typography>
            )}
        </Box>
    );
}
