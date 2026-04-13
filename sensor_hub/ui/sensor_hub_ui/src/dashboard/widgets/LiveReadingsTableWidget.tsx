import type { WidgetProps } from '../types';
import CurrentTemperatures from '../../components/CurrentTemperatures';
import { useReportWidgetUpdate } from '../WidgetUpdateContext';

export default function LiveReadingsTableWidget(_props: WidgetProps) {
    const reportUpdate = useReportWidgetUpdate();
    return <CurrentTemperatures cardHeight="100%" showTitle={false} onDataUpdate={reportUpdate} />;
}

