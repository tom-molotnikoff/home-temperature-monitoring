import { useState, useEffect } from 'react';
import {
    Dialog, DialogTitle, DialogContent, DialogActions,
    Button, TextField, FormControlLabel, Switch,
    MenuItem, Select, InputLabel, FormControl,
} from '@mui/material';
import { getWidget } from './WidgetRegistry';
import { useDashboard } from './DashboardContext';
import { useSensorContext } from '../hooks/useSensorContext';

interface WidgetConfigDialogProps {
    open: boolean;
    widgetId: string | null;
    onClose: () => void;
}

export default function WidgetConfigDialog({ open, widgetId, onClose }: WidgetConfigDialogProps) {
    const { config, updateWidgetConfig } = useDashboard();
    const { sensors } = useSensorContext();
    const [localConfig, setLocalConfig] = useState<Record<string, unknown>>({});

    const widget = widgetId ? config.widgets.find((w) => w.id === widgetId) : null;
    const definition = widget ? getWidget(widget.type) : null;

    useEffect(() => {
        if (widget) setLocalConfig({ ...widget.config });
    }, [widget]);

    if (!widget || !definition?.configFields?.length) return null;

    const handleSave = () => {
        if (widgetId) updateWidgetConfig(widgetId, localConfig);
        onClose();
    };

    return (
        <Dialog open={open} onClose={onClose} maxWidth="xs" fullWidth>
            <DialogTitle>Configure {definition.label}</DialogTitle>
            <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 1 }}>
                {definition.configFields.map((field) => {
                    const value = localConfig[field.key] ?? field.defaultValue ?? '';

                    switch (field.type) {
                        case 'text':
                            return (
                                <TextField
                                    key={field.key} label={field.label} fullWidth
                                    value={value as string}
                                    onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.value })}
                                />
                            );
                        case 'number':
                            return (
                                <TextField
                                    key={field.key} label={field.label} fullWidth type="number"
                                    value={value as number}
                                    onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: Number(e.target.value) })}
                                />
                            );
                        case 'boolean':
                            return (
                                <FormControlLabel
                                    key={field.key}
                                    control={
                                        <Switch
                                            checked={Boolean(value)}
                                            onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.checked })}
                                        />
                                    }
                                    label={field.label}
                                />
                            );
                        case 'select':
                            return (
                                <FormControl key={field.key} fullWidth>
                                    <InputLabel>{field.label}</InputLabel>
                                    <Select
                                        value={value as string} label={field.label}
                                        onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.value })}
                                    >
                                        {field.options?.map((opt) => (
                                            <MenuItem key={opt.value} value={opt.value}>{opt.label}</MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>
                            );
                        case 'sensor-select':
                            return (
                                <FormControl key={field.key} fullWidth>
                                    <InputLabel>{field.label}</InputLabel>
                                    <Select
                                        value={(value as number) || ''} label={field.label}
                                        onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: Number(e.target.value) })}
                                    >
                                        {sensors.map((s) => (
                                            <MenuItem key={s.id} value={s.id}>{s.name}</MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>
                            );
                        default:
                            return null;
                    }
                })}
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button variant="contained" onClick={handleSave}>Save</Button>
            </DialogActions>
        </Dialog>
    );
}
