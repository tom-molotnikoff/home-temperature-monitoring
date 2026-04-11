import type { WidgetProps } from '../types';
import { useSensorContext } from '../../hooks/useSensorContext';
import ReadingsChart from '../../components/ReadingsChart';
import NeedsConfiguration from '../NeedsConfiguration';
import { DateTime } from 'luxon';

export default function ReadingsChartWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const measurementType = config.measurementType as string | undefined;

    if (!measurementType) {
        return <NeedsConfiguration message="Select a measurement type to display" />;
    }

    const startDate = typeof config.startDate === 'string' && config.startDate
        ? DateTime.fromISO(config.startDate)
        : DateTime.now().startOf('day');
    const endDate = typeof config.endDate === 'string' && config.endDate
        ? DateTime.fromISO(config.endDate)
        : DateTime.now().plus({ days: 1 }).startOf('day');
    const useHourlyAverages = Boolean(config.useHourlyAverages);

    return (
        <div style={{ display: 'flex', flexDirection: 'column', flex: 1, minHeight: 0, width: '100%' }}>
            <ReadingsChart
                sensors={sensors}
                useHourlyAverages={useHourlyAverages}
                startDate={startDate}
                endDate={endDate}
                measurementType={measurementType}
            />
        </div>
    );
}
