import type { WidgetDefinition } from './types';

const registry = new Map<string, WidgetDefinition>();

export function registerWidget(definition: WidgetDefinition): void {
    registry.set(definition.type, definition);
}

export function getWidget(type: string): WidgetDefinition | undefined {
    return registry.get(type);
}

export function getAllWidgets(): WidgetDefinition[] {
    return Array.from(registry.values());
}

export function getWidgetComponent(type: string) {
    return registry.get(type)?.component;
}
