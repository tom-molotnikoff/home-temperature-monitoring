import type { WidgetProps } from '../types';
import { useSensorContext } from '../../hooks/useSensorContext';
import TemperatureGraph from '../../components/TemperatureGraph';
import { DateTime } from 'luxon';

export default function TemperatureChartWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();

    const startDate = typeof config.startDate === 'string' && config.startDate
        ? DateTime.fromISO(config.startDate)
        : DateTime.now().startOf('day');
    const endDate = typeof config.endDate === 'string' && config.endDate
        ? DateTime.fromISO(config.endDate)
        : DateTime.now().plus({ days: 1 }).startOf('day');
    const useHourlyAverages = Boolean(config.useHourlyAverages);

    return (
        <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, width: '100%' }}>
            <TemperatureGraph
                sensors={sensors}
                useHourlyAverages={useHourlyAverages}
                startDate={startDate}
                endDate={endDate}
            />
        </div>
    );
}
