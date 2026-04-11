import { useEffect, useState } from 'react';
import { Paper, Typography, Grid } from '@mui/material';
import type { Sensor, Reading, MeasurementTypeInfo } from '../types/types';
import { MeasurementTypesApi } from '../api/Sensors';
import { ReadingsApi } from '../api/Readings';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';

interface SensorDetailCardProps {
    sensor: Sensor;
}

export default function SensorDetailCard({ sensor }: SensorDetailCardProps) {
    const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
    const [latestReadings, setLatestReadings] = useState<Record<string, Reading>>({});

    useEffect(() => {
        MeasurementTypesApi.getForSensor(sensor.id).then(setMeasurementTypes);
    }, [sensor.id]);

    useEffect(() => {
        if (measurementTypes.length === 0) return;

        const now = new Date();
        const oneDayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000);
        ReadingsApi.getBetweenDates(
            oneDayAgo.toISOString().slice(0, 10),
            now.toISOString().slice(0, 10),
            sensor.name,
        ).then((readings) => {
            const byType: Record<string, Reading> = {};
            for (const r of readings) {
                if (!byType[r.measurement_type] || new Date(r.time) > new Date(byType[r.measurement_type].time)) {
                    byType[r.measurement_type] = r;
                }
            }
            setLatestReadings(byType);
        });
    }, [sensor.name, measurementTypes]);

    if (measurementTypes.length === 0) return null;

    return (
        <LayoutCard>
            <TypographyH2>{sensor.name} — Details</TypographyH2>
            <Grid container spacing={1} sx={{ mt: 1 }}>
                {measurementTypes.map((mt) => {
                    const reading = latestReadings[mt.name];
                    return (
                        <Grid key={mt.name} size={{ xs: 6, sm: 4, md: 3 }}>
                            <Paper variant="outlined" sx={{ p: 1.5, textAlign: 'center' }}>
                                <Typography variant="caption" color="text.secondary">
                                    {mt.display_name}
                                </Typography>
                                <Typography variant="h6" sx={{ fontWeight: 'bold' }}>
                                    {reading?.numeric_value != null
                                        ? `${reading.numeric_value.toFixed(1)} ${reading.unit ?? mt.unit}`
                                        : reading?.text_state ?? '—'}
                                </Typography>
                            </Paper>
                        </Grid>
                    );
                })}
            </Grid>
        </LayoutCard>
    );
}
