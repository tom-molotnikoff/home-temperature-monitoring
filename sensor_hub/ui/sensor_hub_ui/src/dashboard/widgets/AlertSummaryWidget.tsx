import type { WidgetProps } from '../types';
import type { AlertRule } from '../../api/Alerts';
import { useState, useEffect } from 'react';
import { Box, Typography, List, ListItem, ListItemText, Chip } from '@mui/material';
import { listAlertRules } from '../../api/Alerts';

export default function AlertSummaryWidget(_props: WidgetProps) {
    const [rules, setRules] = useState<AlertRule[]>([]);
    const [loaded, setLoaded] = useState(false);

    useEffect(() => {
        listAlertRules().then((data) => {
            setRules(data ?? []);
            setLoaded(true);
        }).catch(() => {
            setLoaded(true);
        });
    }, []);

    if (!loaded) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">Loading…</Typography>
            </Box>
        );
    }

    if (rules.length === 0) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">No alert rules configured</Typography>
            </Box>
        );
    }

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', overflow: 'hidden' }}>
            <Box sx={{ overflow: 'auto', flex: 1, minHeight: 0 }}>
                <List dense>
                    {rules.map((rule) => (
                        <ListItem key={rule.ID}>
                            <ListItemText
                                primary={rule.SensorName}
                                secondary={`${rule.AlertType} — threshold: ${rule.HighThreshold ?? rule.LowThreshold ?? '—'}${rule.LastAlertSentAt ? ` · last: ${new Date(rule.LastAlertSentAt).toLocaleDateString()}` : ''}`}
                            />
                            <Chip
                                label={rule.Enabled ? 'Enabled' : 'Disabled'}
                                size="small"
                                color={rule.Enabled ? 'success' : 'default'}
                            />
                        </ListItem>
                    ))}
                </List>
            </Box>
        </Box>
    );
}
