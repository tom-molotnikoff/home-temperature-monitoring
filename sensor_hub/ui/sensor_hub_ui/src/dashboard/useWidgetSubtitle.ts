import { useSensorContext } from '../hooks/useSensorContext';
import { useProperties } from '../hooks/useProperties';

export function useWidgetSubtitle(type: string, config: Record<string, unknown>): string | null {
    const { sensors } = useSensorContext();
    const properties = useProperties();

    if (type === 'weather-forecast') {
        const name = properties["weather.location.name"];
        return typeof name === 'string' && name ? name : null;
    }

    if (typeof config.sensorId === 'number') {
        const sensor = sensors.find(s => s.id === config.sensorId);
        return sensor?.name ?? null;
    }

    if (Array.isArray(config.sensorIds) && config.sensorIds.length > 0) {
        const names = (config.sensorIds as number[])
            .map(id => sensors.find(s => s.id === id)?.name)
            .filter(Boolean) as string[];
        return names.length > 3 ? `${names.length} sensors` : names.join(', ');
    }

    return null;
}
