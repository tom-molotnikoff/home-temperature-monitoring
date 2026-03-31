import { useState } from 'react';
import { Box, Button } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DashboardIcon from '@mui/icons-material/Dashboard';
import PageContainer from '../tools/PageContainer';
import { DashboardProvider, useDashboard } from './DashboardContext';
import DashboardEngine from './DashboardEngine';
import DashboardToolbar from './DashboardToolbar';
import WidgetPickerDialog from './WidgetPickerDialog';
import WidgetConfigDialog from './WidgetConfigDialog';
import EmptyState from '../components/EmptyState';
import { registerAllWidgets } from './widgets';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';

registerAllWidgets();

function DashboardPageInner() {
    const { user } = useAuth();
    const { config, isEditing, loading, updateWidgets, removeWidget, activeDashboard } = useDashboard();
    const [pickerOpen, setPickerOpen] = useState(false);
    const [configWidgetId, setConfigWidgetId] = useState<string | null>(null);

    if (loading) return null;

    const canManage = hasPerm(user, 'manage_dashboards');

    if (!activeDashboard) {
        return (
            <EmptyState
                icon={<DashboardIcon sx={{ fontSize: 48 }} />}
                title="No dashboards yet"
                description={canManage ? 'Create your first dashboard to get started.' : 'No dashboards are available.'}
                actionLabel={canManage ? 'Create Dashboard' : undefined}
            />
        );
    }

    if (config.widgets.length === 0 && !isEditing) {
        return (
            <>
                <DashboardToolbar onAddWidget={() => setPickerOpen(true)} />
                <EmptyState
                    icon={<DashboardIcon sx={{ fontSize: 48 }} />}
                    title="Empty dashboard"
                    description={canManage ? 'Click Edit then Add Widget to populate this dashboard.' : 'This dashboard has no widgets yet.'}
                />
            </>
        );
    }

    return (
        <>
            <DashboardToolbar onAddWidget={() => setPickerOpen(true)} />

            {config.widgets.length === 0 && isEditing ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
                    <Button variant="outlined" startIcon={<AddIcon />} onClick={() => setPickerOpen(true)}>
                        Add your first widget
                    </Button>
                </Box>
            ) : (
                <DashboardEngine
                    config={config}
                    isEditing={isEditing}
                    onLayoutChange={updateWidgets}
                    onRemoveWidget={removeWidget}
                    onConfigureWidget={(id) => setConfigWidgetId(id)}
                />
            )}

            <WidgetPickerDialog open={pickerOpen} onClose={() => setPickerOpen(false)} />
            <WidgetConfigDialog open={!!configWidgetId} widgetId={configWidgetId} onClose={() => setConfigWidgetId(null)} />
        </>
    );
}

export default function DashboardPage() {
    const { user } = useAuth();

    return (
        <PageContainer titleText="Dashboard" loading={user === undefined}>
            <DashboardProvider>
                <DashboardPageInner />
            </DashboardProvider>
        </PageContainer>
    );
}
