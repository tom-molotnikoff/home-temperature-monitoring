import { Component, type ReactNode } from 'react';
import { Paper, Typography, Button, Box } from '@mui/material';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';

interface Props {
    children: ReactNode;
    widgetId: string;
    onRemove?: (id: string) => void;
    onConfigure?: (id: string) => void;
}

interface State {
    hasError: boolean;
}

export class WidgetErrorBoundary extends Component<Props, State> {
    state: State = { hasError: false };

    static getDerivedStateFromError(): State {
        return { hasError: true };
    }

    render() {
        if (!this.state.hasError) return this.props.children;

        return (
            <Paper sx={{ p: 2, height: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 1 }}>
                <WarningAmberIcon color="warning" sx={{ fontSize: 40 }} />
                <Typography variant="subtitle2" color="text.secondary" align="center">
                    This widget encountered an error.
                </Typography>
                <Typography variant="caption" color="text.secondary" align="center">
                    Try editing its configuration or removing it.
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                    {this.props.onConfigure && (
                        <Button size="small" variant="outlined" onClick={() => this.props.onConfigure!(this.props.widgetId)}>
                            Reconfigure
                        </Button>
                    )}
                    {this.props.onRemove && (
                        <Button size="small" color="error" onClick={() => this.props.onRemove!(this.props.widgetId)}>
                            Remove
                        </Button>
                    )}
                </Box>
            </Paper>
        );
    }
}
