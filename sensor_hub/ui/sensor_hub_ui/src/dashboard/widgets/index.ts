import { registerWidget } from '../WidgetRegistry';
import TemperatureChartWidget from './TemperatureChartWidget';
import LiveReadingsTableWidget from './LiveReadingsTableWidget';
import WeatherForecastWidget from './WeatherForecastWidget';
import SensorHealthPieWidget from './SensorHealthPieWidget';
import SensorTypePieWidget from './SensorTypePieWidget';
import HealthTimelineWidget from './HealthTimelineWidget';
import ReadingStatsWidget from './ReadingStatsWidget';
import NotificationsFeedWidget from './NotificationsFeedWidget';

export function registerAllWidgets(): void {
    registerWidget({
        type: 'temperature-chart',
        label: 'Temperature Chart',
        description: 'Indoor temperature line chart with date range picker',
        component: TemperatureChartWidget,
        defaultConfig: {},
        defaultLayout: { w: 12, h: 4 },
        minW: 6,
        minH: 3,
    });

    registerWidget({
        type: 'live-readings',
        label: 'Live Readings Table',
        description: 'Real-time temperature readings data grid',
        component: LiveReadingsTableWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 5 },
        minW: 4,
        minH: 3,
    });

    registerWidget({
        type: 'weather-forecast',
        label: 'Weather Forecast',
        description: 'External weather forecast from configured provider',
        component: WeatherForecastWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 4 },
        minW: 4,
        minH: 3,
    });

    registerWidget({
        type: 'sensor-health-pie',
        label: 'Sensor Health',
        description: 'Pie chart showing sensor health status distribution',
        component: SensorHealthPieWidget,
        defaultConfig: {},
        defaultLayout: { w: 4, h: 4 },
        minW: 3,
        minH: 3,
    });

    registerWidget({
        type: 'sensor-type-pie',
        label: 'Sensor Types',
        description: 'Pie chart showing sensor type distribution',
        component: SensorTypePieWidget,
        defaultConfig: {},
        defaultLayout: { w: 4, h: 4 },
        minW: 3,
        minH: 3,
    });

    registerWidget({
        type: 'health-timeline',
        label: 'Health Timeline',
        description: 'Sensor health status history chart',
        component: HealthTimelineWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 4 },
        minW: 4,
        minH: 3,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
        ],
    });

    registerWidget({
        type: 'reading-stats',
        label: 'Reading Statistics',
        description: 'Total readings per sensor data grid',
        component: ReadingStatsWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 5 },
        minW: 4,
        minH: 3,
    });

    registerWidget({
        type: 'notifications-feed',
        label: 'Notifications',
        description: 'Recent notifications feed',
        component: NotificationsFeedWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 5 },
        minW: 4,
        minH: 3,
    });
}
