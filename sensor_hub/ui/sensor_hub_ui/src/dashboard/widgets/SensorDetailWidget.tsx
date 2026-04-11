import type { WidgetProps } from '../types';
import { useSensorContext } from '../../hooks/useSensorContext';
import NeedsConfiguration from '../NeedsConfiguration';
import SensorDetailCard from '../../components/SensorDetailCard';

export default function SensorDetailWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();

    const sensorId = config.sensorId as number | undefined;
    const sensor = sensorId ? sensors.find((s) => s.id === sensorId) : undefined;

    if (!sensor) {
        return <NeedsConfiguration message="Select a sensor to view its readings." />;
    }

    return <SensorDetailCard sensor={sensor} />;
}
