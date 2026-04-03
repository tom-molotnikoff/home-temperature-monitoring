import type { WidgetProps } from '../types';
import TotalReadingsForEachSensorCard from '../../components/TotalReadingsForEachSensorCard';

export default function ReadingStatsWidget(_props: WidgetProps) {
    return <TotalReadingsForEachSensorCard showTitle={false} />;
}
