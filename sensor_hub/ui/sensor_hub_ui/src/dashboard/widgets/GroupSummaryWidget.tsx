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

    const temps = entries.map(([, r]) => r.numeric_value ?? 0);
    const avg = temps.reduce((sum, t) => sum + t, 0) / temps.length;

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 'bold' }}>Group Average</Typography>
            <Box sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography variant="h2" sx={{ fontWeight: 'bold', textAlign: 'center' }}>
                    {avg.toFixed(1)}°
                </Typography>
            </Box>
            <Box sx={{ maxHeight: 120, overflow: 'auto' }}>
                {entries.map(([name, reading]) => (
                    <Box key={name} sx={{ display: 'flex', justifyContent: 'space-between', py: 0.25 }}>
                        <Typography variant="caption" color="text.secondary">{name}</Typography>
                        <Typography variant="caption">{reading.numeric_value?.toFixed(1) ?? '—'}°</Typography>
                    </Box>
                ))}
            </Box>
        </Box>
    );
}
