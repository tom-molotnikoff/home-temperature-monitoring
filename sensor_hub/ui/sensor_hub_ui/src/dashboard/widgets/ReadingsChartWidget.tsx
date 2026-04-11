import type { WidgetProps } from '../types';
import { useSensorContext } from '../../hooks/useSensorContext';
import ReadingsChart from '../../components/ReadingsChart';
import NeedsConfiguration from '../NeedsConfiguration';
import { resolveTimeRange } from '../timeRange';

export default function ReadingsChartWidget({ config }: WidgetProps) {
    const { sensors } = useSensorContext();
    const measurementType = config.measurementType as string | undefined;

    if (!measurementType) {
        return <NeedsConfiguration message="Select a measurement type to display" />;
    }

    const { startDate, endDate } = resolveTimeRange(config);
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
