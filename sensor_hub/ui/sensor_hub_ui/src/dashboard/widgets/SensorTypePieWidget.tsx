import { useEffect } from 'react';
import type { WidgetProps } from '../types';
import SensorTypeCard from '../../components/SensorTypeCard';
import { useReportWidgetUpdate } from '../WidgetUpdateContext';

export default function SensorTypePieWidget(_props: WidgetProps) {
    const reportUpdate = useReportWidgetUpdate();
    useEffect(() => { reportUpdate(new Date()); }, [reportUpdate]);
    return <SensorTypeCard showTitle={false} />;
}
