import { registerWidget, registerAlias } from '../WidgetRegistry';
import ReadingsChartWidget from './ReadingsChartWidget';
import LiveReadingsTableWidget from './LiveReadingsTableWidget';
import WeatherForecastWidget from './WeatherForecastWidget';
import SensorHealthPieWidget from './SensorHealthPieWidget';
import SensorTypePieWidget from './SensorTypePieWidget';
import HealthTimelineWidget from './HealthTimelineWidget';
import ReadingStatsWidget from './ReadingStatsWidget';
import NotificationsFeedWidget from './NotificationsFeedWidget';
import MarkdownNoteWidget from './MarkdownNoteWidget';
import CurrentReadingWidget from './CurrentReadingWidget';
import MinMaxAvgWidget from './MinMaxAvgWidget';
import GaugeWidget from './GaugeWidget';
import ComparisonChartWidget from './ComparisonChartWidget';
import GroupSummaryWidget from './GroupSummaryWidget';
import AlertSummaryWidget from './AlertSummaryWidget';
import UptimeWidget from './UptimeWidget';
import HeatmapWidget from './HeatmapWidget';
import SensorDetailWidget from './SensorDetailWidget';

export function registerAllWidgets(): void {
    registerWidget({
        type: 'readings-chart',
        label: 'Readings Chart',
        description: 'Line chart for any measurement type with configurable date range',
        component: ReadingsChartWidget,
        defaultConfig: {},
        defaultLayout: { w: 12, h: 4 },
        minW: 6,
        minH: 3,
        configFields: [
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
            { key: 'timeRange', label: 'Time Range', type: 'time-range' },
            { key: 'useHourlyAverages', label: 'Hourly Averages', type: 'boolean' },
            { key: 'refreshInterval', label: 'Refresh Interval (seconds)', type: 'number', defaultValue: 30 },
        ],
    });
    registerAlias('temperature-chart', 'readings-chart');

    registerWidget({
        type: 'live-readings',
        label: 'Live Readings Table',
        description: 'Real-time sensor readings data grid',
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
            { key: 'limit', label: 'History Limit', type: 'number', defaultValue: 1000 },
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

    registerWidget({
        type: 'markdown-note',
        label: 'Markdown Note',
        description: 'User-defined text block for notes or labels',
        component: MarkdownNoteWidget,
        defaultConfig: {},
        defaultLayout: { w: 4, h: 3 },
        minW: 2,
        minH: 2,
        configFields: [
            { key: 'content', label: 'Content', type: 'textarea' },
        ],
    });

    registerWidget({
        type: 'current-reading',
        label: 'Current Reading',
        description: 'Big number display for a single sensor',
        component: CurrentReadingWidget,
        defaultConfig: {},
        defaultLayout: { w: 3, h: 3 },
        minW: 2,
        minH: 2,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
        ],
    });

    registerWidget({
        type: 'min-max-avg',
        label: 'Min / Max / Avg',
        description: 'Period statistics (min, max, average) for a sensor and measurement type',
        component: MinMaxAvgWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 3 },
        minW: 4,
        minH: 2,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
            { key: 'timeRange', label: 'Time Range', type: 'time-range' },
        ],
    });

    registerWidget({
        type: 'gauge',
        label: 'Gauge',
        description: 'Visual circular gauge for a single sensor',
        component: GaugeWidget,
        defaultConfig: { min: 0, max: 40 },
        defaultLayout: { w: 3, h: 3 },
        minW: 2,
        minH: 3,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
            { key: 'min', label: 'Min Value', type: 'number', defaultValue: 0 },
            { key: 'max', label: 'Max Value', type: 'number', defaultValue: 40 },
        ],
    });

    registerWidget({
        type: 'comparison-chart',
        label: 'Comparison Chart',
        description: 'Multi-sensor overlay line chart for any measurement type',
        component: ComparisonChartWidget,
        defaultConfig: {},
        defaultLayout: { w: 12, h: 4 },
        minW: 6,
        minH: 3,
        configFields: [
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
            { key: 'sensorIds', label: 'Sensors', type: 'multi-sensor-select' },
            { key: 'timeRange', label: 'Time Range', type: 'time-range' },
            { key: 'useHourlyAverages', label: 'Hourly Averages', type: 'boolean' },
            { key: 'refreshInterval', label: 'Refresh Interval (seconds)', type: 'number', defaultValue: 30 },
        ],
    });

    registerWidget({
        type: 'group-summary',
        label: 'Group Summary',
        description: 'Average reading for a measurement type across all sensors',
        component: GroupSummaryWidget,
        defaultConfig: {},
        defaultLayout: { w: 4, h: 4 },
        minW: 3,
        minH: 3,
        configFields: [
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
        ],
    });

    registerWidget({
        type: 'alert-summary',
        label: 'Alert Summary',
        description: 'Compact list of configured alert rules',
        component: AlertSummaryWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 4 },
        minW: 4,
        minH: 3,
    });

    registerWidget({
        type: 'uptime',
        label: 'Sensor Uptime',
        description: 'Uptime percentage for a sensor',
        component: UptimeWidget,
        defaultConfig: {},
        defaultLayout: { w: 3, h: 3 },
        minW: 2,
        minH: 2,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
            { key: 'limit', label: 'History Limit', type: 'number', defaultValue: 1000 },
        ],
    });

    registerWidget({
        type: 'heatmap',
        label: 'Heatmap',
        description: 'Colour-coded 30-day grid for any measurement type',
        component: HeatmapWidget,
        defaultConfig: {},
        defaultLayout: { w: 4, h: 4 },
        minW: 3,
        minH: 3,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
            { key: 'measurementType', label: 'Measurement Type', type: 'measurement-type-select' },
            { key: 'scaleMin', label: 'Scale Min', type: 'number', defaultValue: 10 },
            { key: 'scaleMax', label: 'Scale Max', type: 'number', defaultValue: 30 },
        ],
    });

    registerWidget({
        type: 'sensor-detail',
        label: 'Sensor Detail',
        description: 'Latest readings grid for all measurement types of a sensor',
        component: SensorDetailWidget,
        defaultConfig: {},
        defaultLayout: { w: 6, h: 4 },
        minW: 4,
        minH: 3,
        configFields: [
            { key: 'sensorId', label: 'Sensor', type: 'sensor-select' },
        ],
    });
}
