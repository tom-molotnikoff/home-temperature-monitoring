import {
    Dialog, DialogTitle, DialogContent, List, ListItemButton,
    ListItemText, Typography,
} from '@mui/material';
import { getAllWidgets } from './WidgetRegistry';
import { useDashboard } from './DashboardContext';
import type { DashboardWidget } from '../types/dashboard';

interface WidgetPickerDialogProps {
    open: boolean;
    onClose: () => void;
}

export default function WidgetPickerDialog({ open, onClose }: WidgetPickerDialogProps) {
    const { addWidget } = useDashboard();
    const widgets = getAllWidgets();

    const handleSelect = (type: string) => {
        const definition = widgets.find((w) => w.type === type);
        if (!definition) return;

        const newWidget: DashboardWidget = {
            id: `${type}-${Date.now()}`,
            type,
            config: { ...definition.defaultConfig },
            layout: {
                x: 0,
                y: Infinity,
                w: definition.defaultLayout.w,
                h: definition.defaultLayout.h,
            },
        };

        addWidget(newWidget);
        onClose();
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
            <DialogTitle>Add Widget</DialogTitle>
            <DialogContent>
                <List>
                    {widgets.map((w) => (
                        <ListItemButton key={w.type} onClick={() => handleSelect(w.type)}>
                            <ListItemText
                                primary={w.label}
                                secondary={<Typography variant="body2" color="text.secondary">{w.description}</Typography>}
                            />
                        </ListItemButton>
                    ))}
                </List>
            </DialogContent>
        </Dialog>
    );
}
