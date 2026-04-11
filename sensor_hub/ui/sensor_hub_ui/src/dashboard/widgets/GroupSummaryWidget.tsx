import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';
import { useCurrentReadings } from '../../hooks/useCurrentReadings';

export default function GroupSummaryWidget(_props: WidgetProps) {
    const readings = useCurrentReadings();
    const entries = Object.entries(readings);

    if (entries.length === 0) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">No sensor readings available</Typography>
            </Box>
        );
    }

    const nums = entries.map(([, r]) => r.numeric_value ?? 0);
    const avg = nums.reduce((sum, t) => sum + t, 0) / nums.length;
    const unit = entries[0]?.[1]?.unit ?? '';

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 'bold' }}>Group Average</Typography>
            <Box sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography variant="h2" sx={{ fontWeight: 'bold', textAlign: 'center' }}>
                    {avg.toFixed(1)}{unit}
                </Typography>
            </Box>
            <Box sx={{ maxHeight: 120, overflow: 'auto' }}>
                {entries.map(([name, reading]) => (
                    <Box key={name} sx={{ display: 'flex', justifyContent: 'space-between', py: 0.25 }}>
                        <Typography variant="caption" color="text.secondary">{name}</Typography>
                        <Typography variant="caption">{reading.numeric_value?.toFixed(1) ?? '—'}{reading.unit ?? ''}</Typography>
                    </Box>
                ))}
            </Box>
        </Box>
    );
}
