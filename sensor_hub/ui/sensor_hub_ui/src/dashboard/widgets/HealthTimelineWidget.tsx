import type { WidgetProps } from '../types';
import SensorHealthHistoryChartCard from '../../components/SensorHealthHistoryChartCard';
import { useSensorContext } from '../../hooks/useSensorContext';
import { Typography } from '@mui/material';

export default function HealthTimelineWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : sensors[0];

    if (!sensor) {
        return <Typography color="text.secondary" sx={{ p: 2 }}>Select a sensor in widget settings</Typography>;
    }

    return <SensorHealthHistoryChartCard sensor={sensor} />;
}
