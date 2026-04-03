import type { WidgetProps } from '../types';
import NotificationsCard from '../../components/NotificationsCard';

export default function NotificationsFeedWidget(_props: WidgetProps) {
    return <NotificationsCard showTitle={false} />;
}
