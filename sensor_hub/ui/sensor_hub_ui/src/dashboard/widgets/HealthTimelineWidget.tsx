import type { WidgetProps } from '../types';
import SensorHealthHistoryChart from '../../components/SensorHealthHistoryChart';
import { useSensorContext } from '../../hooks/useSensorContext';
import { Typography } from '@mui/material';

export default function HealthTimelineWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : sensors[0];
    const limit = typeof config.limit === 'number' && config.limit > 0 ? config.limit : 1000;

    if (!sensor) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>Select a sensor in widget settings</Typography>;
    }

    return <SensorHealthHistoryChart sensor={sensor} limit={limit} />;
}
