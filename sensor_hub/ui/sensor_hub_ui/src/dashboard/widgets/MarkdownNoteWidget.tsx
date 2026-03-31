import type { WidgetProps } from '../types';
import { Box, Typography } from '@mui/material';

export default function MarkdownNoteWidget({ config }: WidgetProps) {
    const content = (config.content as string) || '';

    if (!content) {
        return (
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', p: 2 }}>
                <Typography color="text.secondary">Click settings to add content</Typography>
            </Box>
        );
    }

    return (
        <Box sx={{ p: 2, overflow: 'auto', height: '100%' }}>
            <Typography sx={{ whiteSpace: 'pre-wrap' }}>{content}</Typography>
        </Box>
    );
}
