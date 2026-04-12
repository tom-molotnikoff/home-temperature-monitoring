import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';
import { useCurrentReadings } from '../../hooks/useCurrentReadings';

export default function GroupSummaryWidget(_props: WidgetProps) {
    const readings = useCurrentReadings();

    // Flatten nested map: for each sensor, take all measurement-type readings
    const allReadings = Object.values(readings).flatMap(byType => Object.values(byType));

    if (allReadings.length === 0) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Typography color="text.secondary">No sensor readings available</Typography>
            </Box>
        );
    }

    const nums = allReadings.filter(r => r.numeric_value !== null).map(r => r.numeric_value!);
    const avg = nums.length > 0 ? nums.reduce((sum, t) => sum + t, 0) / nums.length : 0;
    const unit = allReadings[0]?.unit ?? '';

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', p: 2 }}>
            <Typography variant="subtitle1" sx={{ fontWeight: 'bold' }}>Group Average</Typography>
            <Box sx={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <Typography variant="h2" sx={{ fontWeight: 'bold', textAlign: 'center' }}>
                    {avg.toFixed(1)}{unit}
                </Typography>
            </Box>
            <Box sx={{ maxHeight: 120, overflow: 'auto' }}>
                {Object.entries(readings).map(([name, byType]) => {
                    const firstReading = Object.values(byType)[0];
                    return (
                        <Box key={name} sx={{ display: 'flex', justifyContent: 'space-between', py: 0.25 }}>
                            <Typography variant="caption" color="text.secondary">{name}</Typography>
                            <Typography variant="caption">{firstReading?.numeric_value?.toFixed(1) ?? '—'}{firstReading?.unit ?? ''}</Typography>
                        </Box>
                    );
                })}
            </Box>
        </Box>
    );
}
