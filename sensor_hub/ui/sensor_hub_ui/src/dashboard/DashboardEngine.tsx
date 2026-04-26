import { useCallback, useMemo } from 'react';
import { ResponsiveGridLayout, useContainerWidth, type Layout, type LayoutItem } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import WidgetFrame from './WidgetFrame';
import { getWidget } from './WidgetRegistry';
import type { DashboardConfig, DashboardWidget } from '../gen/aliases';
import { GRID_BREAKPOINTS, GRID_COLS, GRID_ROW_HEIGHT } from './constants';

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
    const { width, containerRef } = useContainerWidth();

    const layouts = useMemo(() => {
        const lg: LayoutItem[] = config.widgets.map((w) => ({
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
        (layout: Layout) => {
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
        <div ref={containerRef} style={{ paddingBottom: isEditing ? 200 : 0 }}>
            <ResponsiveGridLayout
                width={width}
                layouts={layouts}
                breakpoints={GRID_BREAKPOINTS}
                cols={GRID_COLS}
                rowHeight={GRID_ROW_HEIGHT}
                dragConfig={{ enabled: isEditing, handle: '.drag-handle' }}
                resizeConfig={{ enabled: isEditing }}
                onLayoutChange={handleLayoutChange}
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
        </div>
    );
}
