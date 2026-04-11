import type { WidgetDefinition } from './types';

const registry = new Map<string, WidgetDefinition>();
const aliases = new Map<string, string>();

export function registerWidget(definition: WidgetDefinition): void {
    registry.set(definition.type, definition);
}

export function registerAlias(oldType: string, newType: string): void {
    aliases.set(oldType, newType);
}

export function getWidget(type: string): WidgetDefinition | undefined {
    return registry.get(type) ?? registry.get(aliases.get(type) ?? '');
}

export function getAllWidgets(): WidgetDefinition[] {
    return Array.from(registry.values());
}

export function getWidgetComponent(type: string) {
    return registry.get(type)?.component ?? registry.get(aliases.get(type) ?? '')?.component;
}
