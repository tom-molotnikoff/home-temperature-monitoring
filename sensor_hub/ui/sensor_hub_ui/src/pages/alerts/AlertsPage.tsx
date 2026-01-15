import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef, GridRowParams } from '@mui/x-data-grid';
import {
  Button,
  Box,
  Menu,
  MenuItem,
  Typography,
} from '@mui/material';
import { listAlertRules } from '../../api/Alerts';
import type { AlertRule } from '../../api/Alerts';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import { useAuth } from "../../providers/AuthContext.tsx";
import { hasPerm } from "../../tools/Utils.ts";

export default function AlertsPage() {
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  
  const { user } = useAuth();

  const load = async () => {
    try {
      const rules = await listAlertRules();
      setAlertRules(rules);
    } catch (e) {
      console.error('Failed to load alert rules', e);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleRowClick = (_params: GridRowParams, event: React.MouseEvent) => {
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const handleMenuClose = () => { setMenuAnchorEl(null); };

  const columns: GridColDef[] = [
    { field: 'SensorName', headerName: 'Sensor', flex: 1 },
    { field: 'AlertType', headerName: 'Alert Type', width: 150 },
    { field: 'HighThreshold', headerName: 'High Threshold', width: 130 },
    { field: 'LowThreshold', headerName: 'Low Threshold', width: 130 },
    { field: 'TriggerStatus', headerName: 'Trigger Status', width: 130 },
    { field: 'RateLimitHours', headerName: 'Rate Limit (hrs)', width: 130 },
    { field: 'Enabled', headerName: 'Enabled', width: 100 },
    { field: 'LastAlertSentAt', headerName: 'Last Alert Sent', width: 180 },
  ];

  const rows = alertRules.map(r => ({ 
    id: r.SensorID, 
    ...r,
    HighThreshold: r.HighThreshold ?? '-',
    LowThreshold: r.LowThreshold ?? '-',
    TriggerStatus: r.TriggerStatus || '-',
    Enabled: r.Enabled ? 'Yes' : 'No',
    LastAlertSentAt: r.LastAlertSentAt ? new Date(r.LastAlertSentAt).toLocaleString() : 'Never',
  }));

  if (user === undefined) {
    return (
      <PageContainer titleText="Alert Rules">
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
          Loading...
        </Box>
      </PageContainer>
    );
  }

  const fieldsDisabled = !hasPerm(user, "manage_alerts");

  return (
    <PageContainer titleText="Alert Rules">
      <Box sx={{ flexGrow: 1 }}>
        <LayoutCard variant="secondary" changes={{ alignItems: "stretch", height: "100%", width: "100%" }}>
          <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{ width: '100%' }}>
            <Typography variant="h4">Alert Rules</Typography>
            <Box>
              <Button variant="contained" disabled={fieldsDisabled}>
                Create Alert Rule
              </Button>
            </Box>
          </Box>
          <div style={{ height: 400, width: '100%' }}>
            <DataGrid 
              rows={rows} 
              columns={columns} 
              pageSizeOptions={[5, 10, 25]} 
              initialState={{ pagination: { paginationModel: { pageSize: 10 } } }} 
              onRowClick={handleRowClick} 
            />
          </div>

          {hasPerm(user, "view_alerts") && (
            <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={handleMenuClose}>
              <MenuItem disabled={fieldsDisabled}>Edit</MenuItem>
              <MenuItem disabled={fieldsDisabled}>Delete</MenuItem>
              <MenuItem>View History</MenuItem>
            </Menu>
          )}
        </LayoutCard>
      </Box>
    </PageContainer>
  );
}
