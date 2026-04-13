import { useEffect } from 'react';
import type { WidgetProps } from '../types';
import SensorHealthCard from '../../components/SensorHealthCard';
import { useReportWidgetUpdate } from '../WidgetUpdateContext';

export default function SensorHealthPieWidget(_props: WidgetProps) {
    const reportUpdate = useReportWidgetUpdate();
    useEffect(() => { reportUpdate(new Date()); }, [reportUpdate]);
    return <SensorHealthCard showTitle={false} />;
}
