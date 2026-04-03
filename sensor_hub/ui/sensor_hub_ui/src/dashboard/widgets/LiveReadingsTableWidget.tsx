import type { WidgetProps } from '../types';
import CurrentTemperatures from '../../components/CurrentTemperatures';

export default function LiveReadingsTableWidget(_props: WidgetProps) {
    return <CurrentTemperatures cardHeight="100%" showTitle={false} />;
}

