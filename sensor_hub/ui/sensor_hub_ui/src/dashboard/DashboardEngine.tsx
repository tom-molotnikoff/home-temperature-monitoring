import { useCallback, useMemo } from 'react';
import { Responsive, WidthProvider, type Layout } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import WidgetFrame from './WidgetFrame';
import { getWidget } from './WidgetRegistry';
import type { DashboardConfig, DashboardWidget } from '../types/dashboard';
import { GRID_BREAKPOINTS, GRID_COLS, GRID_ROW_HEIGHT } from '../types/dashboard';

const ResponsiveGridLayout = WidthProvider(Responsive);

interface DashboardEngineProps {
    config: DashboardConfig;
    isEditing: boolean;
    onLayoutChange: (widgets: DashboardWidget[]) => void;
    onRemoveWidget: (id: string) => void;
    onConfigureWidget: (id: string) => void;
}

export default function DashboardEngine({
    config,
    isEditing,
    onLayoutChange,
    onRemoveWidget,
    onConfigureWidget,
}: DashboardEngineProps) {
    const layouts = useMemo(() => {
        const lg = config.widgets.map((w) => ({
            i: w.id,
            x: w.layout.x,
            y: w.layout.y,
            w: w.layout.w,
            h: w.layout.h,
            minW: getWidget(w.type)?.minW,
            minH: getWidget(w.type)?.minH,
            maxW: getWidget(w.type)?.maxW,
            maxH: getWidget(w.type)?.maxH,
        }));
        return { lg };
    }, [config.widgets]);

    const handleLayoutChange = useCallback(
        (layout: Layout[]) => {
            const updated = config.widgets.map((widget) => {
                const item = layout.find((l) => l.i === widget.id);
                if (!item) return widget;
                return {
                    ...widget,
                    layout: { x: item.x, y: item.y, w: item.w, h: item.h },
                };
            });
            onLayoutChange(updated);
        },
        [config.widgets, onLayoutChange],
    );

    return (
        <ResponsiveGridLayout
            layouts={layouts}
            breakpoints={GRID_BREAKPOINTS}
            cols={GRID_COLS}
            rowHeight={GRID_ROW_HEIGHT}
            isDraggable={isEditing}
            isResizable={isEditing}
            draggableHandle=".drag-handle"
            onLayoutChange={handleLayoutChange}
            compactType="vertical"
            margin={[16, 16]}
        >
            {config.widgets.map((widget) => (
                <div key={widget.id}>
                    <WidgetFrame
                        widget={widget}
                        isEditing={isEditing}
                        onRemove={onRemoveWidget}
                        onConfigure={onConfigureWidget}
                    />
                </div>
            ))}
        </ResponsiveGridLayout>
    );
}
