import { useEffect } from 'react';
import type { WidgetProps } from '../types';
import NotificationsCard from '../../components/NotificationsCard';
import { useReportWidgetUpdate } from '../WidgetUpdateContext';

export default function NotificationsFeedWidget(_props: WidgetProps) {
    const reportUpdate = useReportWidgetUpdate();
    useEffect(() => { reportUpdate(new Date()); }, [reportUpdate]);
    return <NotificationsCard showTitle={false} />;
}
