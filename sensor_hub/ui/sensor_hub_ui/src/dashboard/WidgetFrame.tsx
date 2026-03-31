import { Paper, IconButton, Box, Typography } from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import SettingsIcon from '@mui/icons-material/Settings';
import DragIndicatorIcon from '@mui/icons-material/DragIndicator';
import { getWidget } from './WidgetRegistry';
import type { WidgetProps } from './types';
import type { DashboardWidget } from '../types/dashboard';

interface WidgetFrameProps {
    widget: DashboardWidget;
    isEditing: boolean;
    onRemove: (id: string) => void;
    onConfigure: (id: string) => void;
}

export default function WidgetFrame({ widget, isEditing, onRemove, onConfigure }: WidgetFrameProps) {
    const definition = getWidget(widget.type);
    if (!definition) {
        return (
            <Paper sx={{ p: 2, height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography color="error">Unknown widget: {widget.type}</Typography>
            </Paper>
        );
    }

    const Component = definition.component;
    const hasConfig = definition.configFields && definition.configFields.length > 0;
    const widgetProps: WidgetProps = {
        id: widget.id,
        config: widget.config,
        isEditing,
    };

    return (
        <Paper
            elevation={isEditing ? 3 : 1}
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                overflow: 'hidden',
                border: isEditing ? '1px dashed' : '1px solid',
                borderColor: isEditing ? 'primary.main' : 'divider',
                borderRadius: 2,
                position: 'relative',
                userSelect: isEditing ? 'none' : 'auto',
            }}
        >
            {isEditing && (
                <Box
                    className="drag-handle"
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        px: 1,
                        py: 0.5,
                        bgcolor: 'action.hover',
                        borderBottom: '1px solid',
                        borderColor: 'divider',
                        cursor: 'grab',
                        flexShrink: 0,
                    }}
                >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                        <DragIndicatorIcon fontSize="small" color="action" />
                        <Typography variant="caption" color="text.secondary">{definition.label}</Typography>
                    </Box>
                    <Box>
                        {hasConfig && (
                            <IconButton size="small" onClick={() => onConfigure(widget.id)}>
                                <SettingsIcon fontSize="small" />
                            </IconButton>
                        )}
                        <IconButton size="small" onClick={() => onRemove(widget.id)}>
                            <CloseIcon fontSize="small" />
                        </IconButton>
                    </Box>
                </Box>
            )}
            <Box sx={{
                flex: 1,
                minHeight: 0,
                overflow: 'hidden',
                p: isEditing ? 1 : 0,
                '& > *': { height: '100%', width: '100%' },
            }}>
                <Component {...widgetProps} />
            </Box>
        </Paper>
    );
}
