import { useState, useEffect } from 'react';
import {
    Dialog, DialogTitle, DialogContent, DialogActions,
    Button, TextField, FormControlLabel, Switch,
    MenuItem, Select, InputLabel, FormControl, Checkbox, ListItemText,
} from '@mui/material';
import { DatePicker } from '@mui/x-date-pickers';
import { DateTime } from 'luxon';
import { getWidget } from './WidgetRegistry';
import { useDashboard } from './DashboardContext';
import { useSensorContext } from '../hooks/useSensorContext';
import { useMeasurementTypes } from '../hooks/useMeasurementTypes';
import { TIME_RANGE_PRESETS } from './timeRange';

interface WidgetConfigDialogProps {
    open: boolean;
    widgetId: string | null;
    onClose: () => void;
}

export default function WidgetConfigDialog({ open, widgetId, onClose }: WidgetConfigDialogProps) {
    const { config, updateWidgetConfig } = useDashboard();
    const { sensors } = useSensorContext();
    const { measurementTypes } = useMeasurementTypes();
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
            <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                {definition.configFields.map((field) => {
                    const value = localConfig[field.key] ?? field.defaultValue ?? '';

                    switch (field.type) {
                        case 'text':
                            return (
                                <TextField
                                    key={field.key} label={field.label} fullWidth
                                    sx={{ mt: 1 }}
                                    value={value as string}
                                    onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.value })}
                                />
                            );
                        case 'textarea':
                            return (
                                <TextField
                                    key={field.key} label={field.label} fullWidth multiline minRows={3} maxRows={10}
                                    sx={{ mt: 1 }}
                                    value={value as string}
                                    onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.value })}
                                />
                            );
                        case 'number':
                            return (
                                <TextField
                                    key={field.key} label={field.label} fullWidth type="number"
                                    sx={{ mt: 1 }}
                                    value={value as number}
                                    onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: Number(e.target.value) })}
                                />
                            );
                        case 'boolean':
                            return (
                                <FormControlLabel
                                    key={field.key}
                                    sx={{ mt: 1 }}
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
                                <FormControl sx={{ mt: 1 }} key={field.key} fullWidth>
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
                                <FormControl sx={{ mt: 1 }} key={field.key} fullWidth>
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
                        case 'multi-sensor-select': {
                            const selected = (Array.isArray(value) ? value : []) as number[];
                            return (
                                <FormControl sx={{ mt: 1 }} key={field.key} fullWidth>
                                    <InputLabel>{field.label}</InputLabel>
                                    <Select<number[]>
                                        multiple
                                        value={selected}
                                        label={field.label}
                                        onChange={(e) => {
                                            const val = e.target.value;
                                            setLocalConfig({ ...localConfig, [field.key]: typeof val === 'string' ? val.split(',').map(Number) : val });
                                        }}
                                        renderValue={(sel) => sensors.filter((s) => sel.includes(s.id)).map((s) => s.name).join(', ')}
                                    >
                                        {sensors.map((s) => (
                                            <MenuItem key={s.id} value={s.id}>
                                                <Checkbox checked={selected.includes(s.id)} />
                                                <ListItemText primary={s.name} />
                                            </MenuItem>
                                        ))}
                                    </Select>
                                </FormControl>
                            );
                        }
                        case 'date': {
                            const dt = typeof value === 'string' && value
                                ? DateTime.fromISO(value as string)
                                : null;
                            return (
                                <DatePicker
                                    key={field.key}
                                    label={field.label}
                                    value={dt}
                                    onChange={(newVal: DateTime | null) => {
                                        setLocalConfig({
                                            ...localConfig,
                                            [field.key]: newVal?.toISODate() ?? '',
                                        });
                                    }}
                                    slotProps={{ textField: { fullWidth: true, sx: { mt: 1 } } }}
                                />
                            );
                        }
                        case 'time-range': {
                            const rangeValue = (localConfig.timeRange as string) || '24h';
                            const isCustom = rangeValue === 'custom';
                            const customStart = typeof localConfig.customStart === 'string' && localConfig.customStart
                                ? DateTime.fromISO(localConfig.customStart) : null;
                            const customEnd = typeof localConfig.customEnd === 'string' && localConfig.customEnd
                                ? DateTime.fromISO(localConfig.customEnd) : null;
                            return (
                                <div key={field.key}>
                                    <FormControl sx={{ mt: 1 }} fullWidth>
                                        <InputLabel>{field.label}</InputLabel>
                                        <Select
                                            value={rangeValue}
                                            label={field.label}
                                            onChange={(e) => setLocalConfig({ ...localConfig, timeRange: e.target.value })}
                                        >
                                            {TIME_RANGE_PRESETS.map((p) => (
                                                <MenuItem key={p.value} value={p.value}>{p.label}</MenuItem>
                                            ))}
                                        </Select>
                                    </FormControl>
                                    {isCustom && (
                                        <>
                                            <DatePicker
                                                label="Start Date"
                                                value={customStart}
                                                onChange={(v: DateTime | null) =>
                                                    setLocalConfig({ ...localConfig, customStart: v?.toISODate() ?? '' })
                                                }
                                                slotProps={{ textField: { fullWidth: true, sx: { mt: 1 } } }}
                                            />
                                            <DatePicker
                                                label="End Date"
                                                value={customEnd}
                                                onChange={(v: DateTime | null) =>
                                                    setLocalConfig({ ...localConfig, customEnd: v?.toISODate() ?? '' })
                                                }
                                                slotProps={{ textField: { fullWidth: true, sx: { mt: 1 } } }}
                                            />
                                        </>
                                    )}
                                </div>
                            );
                        }
                        case 'measurement-type-select':
                            return (
                                <FormControl sx={{ mt: 1 }} key={field.key} fullWidth>
                                    <InputLabel>{field.label}</InputLabel>
                                    <Select
                                        value={(value as string) || ''} label={field.label}
                                        onChange={(e) => setLocalConfig({ ...localConfig, [field.key]: e.target.value })}
                                    >
                                        {measurementTypes.map((mt) => (
                                            <MenuItem key={mt.name} value={mt.name}>{mt.display_name} ({mt.unit})</MenuItem>
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
