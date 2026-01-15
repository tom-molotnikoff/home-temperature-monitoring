import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef, GridRowParams } from '@mui/x-data-grid';
import {
  Button,
  Box,
  Menu,
  MenuItem,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  InputLabel,
  FormControl,
  FormControlLabel,
  Switch,
} from '@mui/material';
import { listAlertRules, createAlertRule, updateAlertRule, deleteAlertRule } from '../../api/Alerts';
import type { AlertRule, CreateAlertRuleRequest } from '../../api/Alerts';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import { useAuth } from "../../providers/AuthContext.tsx";
import { hasPerm } from "../../tools/Utils.ts";
import { useSensorContext } from "../../hooks/useSensorContext.ts";

export default function AlertsPage() {
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [openCreate, setOpenCreate] = useState(false);
  const [createAlertType, setCreateAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [createSensorId, setCreateSensorId] = useState<number>(0);
  const [createHighThreshold, setCreateHighThreshold] = useState<string>('');
  const [createLowThreshold, setCreateLowThreshold] = useState<string>('');
  const [createTriggerStatus, setCreateTriggerStatus] = useState<string>('');
  const [createRateLimit, setCreateRateLimit] = useState<string>('1');
  const [createEnabled, setCreateEnabled] = useState<boolean>(true);
  const [openEdit, setOpenEdit] = useState(false);
  const [editRule, setEditRule] = useState<AlertRule | null>(null);
  const [editAlertType, setEditAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [editHighThreshold, setEditHighThreshold] = useState<string>('');
  const [editLowThreshold, setEditLowThreshold] = useState<string>('');
  const [editTriggerStatus, setEditTriggerStatus] = useState<string>('');
  const [editRateLimit, setEditRateLimit] = useState<string>('1');
  const [editEnabled, setEditEnabled] = useState<boolean>(true);
  const [selectedRow, setSelectedRow] = useState<AlertRule | null>(null);
  const [openDeleteConfirm, setOpenDeleteConfirm] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<AlertRule | null>(null);
  
  const { user } = useAuth();
  const { sensors } = useSensorContext();

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

  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    const id = typeof params.id === 'number' ? params.id : Number(params.id);
    const found = alertRules.find(r => r.SensorID === id);
    if (found) setSelectedRow(found);
    else setSelectedRow(params.row as AlertRule);
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const handleMenuClose = () => { setMenuAnchorEl(null); };

  const handleCreate = async () => {
    try {
      const request: CreateAlertRuleRequest = {
        SensorID: createSensorId,
        AlertType: createAlertType,
        RateLimitHours: parseInt(createRateLimit, 10),
        Enabled: createEnabled,
      };

      if (createAlertType === 'numeric_range') {
        request.HighThreshold = parseFloat(createHighThreshold);
        request.LowThreshold = parseFloat(createLowThreshold);
      } else {
        request.TriggerStatus = createTriggerStatus;
      }

      await createAlertRule(request);
      setOpenCreate(false);
      setCreateSensorId(0);
      setCreateAlertType('numeric_range');
      setCreateHighThreshold('');
      setCreateLowThreshold('');
      setCreateTriggerStatus('');
      setCreateRateLimit('1');
      setCreateEnabled(true);
      await load();
    } catch (e) {
      console.error('Failed to create alert rule', e);
    }
  };

  const availableSensors = sensors.filter(s => !alertRules.some(r => r.SensorID === s.id));

  const handleOpenEdit = () => {
    if (!selectedRow) return;
    setEditRule(selectedRow);
    setEditAlertType(selectedRow.AlertType);
    setEditHighThreshold(selectedRow.HighThreshold?.toString() || '');
    setEditLowThreshold(selectedRow.LowThreshold?.toString() || '');
    setEditTriggerStatus(selectedRow.TriggerStatus || '');
    setEditRateLimit(selectedRow.RateLimitHours.toString());
    setEditEnabled(selectedRow.Enabled);
    setOpenEdit(true);
    handleMenuClose();
  };

  const handleEdit = async () => {
    if (!editRule) return;
    try {
      const request: CreateAlertRuleRequest = {
        SensorID: editRule.SensorID,
        AlertType: editAlertType,
        RateLimitHours: parseInt(editRateLimit, 10),
        Enabled: editEnabled,
      };

      if (editAlertType === 'numeric_range') {
        request.HighThreshold = parseFloat(editHighThreshold);
        request.LowThreshold = parseFloat(editLowThreshold);
      } else {
        request.TriggerStatus = editTriggerStatus;
      }

      await updateAlertRule(editRule.SensorID, request);
      setOpenEdit(false);
      setEditRule(null);
      await load();
    } catch (e) {
      console.error('Failed to update alert rule', e);
    }
  };

  const handleOpenDelete = () => {
    setDeleteTarget(selectedRow);
    setOpenDeleteConfirm(true);
    handleMenuClose();
  };

  const confirmDelete = async () => {
    const target = deleteTarget || selectedRow;
    if (!target) return;
    try {
      await deleteAlertRule(target.SensorID);
      setOpenDeleteConfirm(false);
      setDeleteTarget(null);
      await load();
    } catch (e) {
      console.error('Failed to delete alert rule', e);
    }
  };

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
              <Button variant="contained" disabled={fieldsDisabled} onClick={() => setOpenCreate(true)}>
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
              <MenuItem disabled={fieldsDisabled} onClick={handleOpenEdit}>Edit</MenuItem>
              <MenuItem disabled={fieldsDisabled} onClick={handleOpenDelete}>Delete</MenuItem>
              <MenuItem>View History</MenuItem>
            </Menu>
          )}
        </LayoutCard>
      </Box>

      <Dialog open={openCreate} onClose={() => setOpenCreate(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create Alert Rule</DialogTitle>
        <DialogContent>
          <FormControl fullWidth sx={{ mt: 1 }}>
            <InputLabel id="create-sensor-label">Sensor</InputLabel>
            <Select
              labelId="create-sensor-label"
              value={createSensorId}
              label="Sensor"
              onChange={(e) => setCreateSensorId(Number(e.target.value))}
            >
              {availableSensors.map(s => (
                <MenuItem key={s.id} value={s.id}>{s.name}</MenuItem>
              ))}
            </Select>
          </FormControl>

          <FormControl fullWidth sx={{ mt: 2 }}>
            <InputLabel id="create-type-label">Alert Type</InputLabel>
            <Select
              labelId="create-type-label"
              value={createAlertType}
              label="Alert Type"
              onChange={(e) => setCreateAlertType(e.target.value as 'numeric_range' | 'status_based')}
            >
              <MenuItem value="numeric_range">Numeric Range</MenuItem>
              <MenuItem value="status_based">Status Based</MenuItem>
            </Select>
          </FormControl>

          {createAlertType === 'numeric_range' ? (
            <>
              <TextField
                fullWidth
                label="High Threshold"
                type="number"
                value={createHighThreshold}
                onChange={(e) => setCreateHighThreshold(e.target.value)}
                sx={{ mt: 2 }}
              />
              <TextField
                fullWidth
                label="Low Threshold"
                type="number"
                value={createLowThreshold}
                onChange={(e) => setCreateLowThreshold(e.target.value)}
                sx={{ mt: 2 }}
              />
            </>
          ) : (
            <TextField
              fullWidth
              label="Trigger Status"
              value={createTriggerStatus}
              onChange={(e) => setCreateTriggerStatus(e.target.value)}
              sx={{ mt: 2 }}
              helperText="e.g., 'open', 'closed', 'motion_detected'"
            />
          )}

          <TextField
            fullWidth
            label="Rate Limit (hours)"
            type="number"
            value={createRateLimit}
            onChange={(e) => setCreateRateLimit(e.target.value)}
            sx={{ mt: 2 }}
          />

          <FormControlLabel
            control={
              <Switch
                checked={createEnabled}
                onChange={(e) => setCreateEnabled(e.target.checked)}
              />
            }
            label="Enabled"
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenCreate(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleCreate}>Create</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openEdit} onClose={() => setOpenEdit(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Alert Rule</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Sensor"
            value={editRule?.SensorName || ''}
            disabled
            sx={{ mt: 1 }}
          />

          <FormControl fullWidth sx={{ mt: 2 }}>
            <InputLabel id="edit-type-label">Alert Type</InputLabel>
            <Select
              labelId="edit-type-label"
              value={editAlertType}
              label="Alert Type"
              onChange={(e) => setEditAlertType(e.target.value as 'numeric_range' | 'status_based')}
            >
              <MenuItem value="numeric_range">Numeric Range</MenuItem>
              <MenuItem value="status_based">Status Based</MenuItem>
            </Select>
          </FormControl>

          {editAlertType === 'numeric_range' ? (
            <>
              <TextField
                fullWidth
                label="High Threshold"
                type="number"
                value={editHighThreshold}
                onChange={(e) => setEditHighThreshold(e.target.value)}
                sx={{ mt: 2 }}
              />
              <TextField
                fullWidth
                label="Low Threshold"
                type="number"
                value={editLowThreshold}
                onChange={(e) => setEditLowThreshold(e.target.value)}
                sx={{ mt: 2 }}
              />
            </>
          ) : (
            <TextField
              fullWidth
              label="Trigger Status"
              value={editTriggerStatus}
              onChange={(e) => setEditTriggerStatus(e.target.value)}
              sx={{ mt: 2 }}
              helperText="e.g., 'open', 'closed', 'motion_detected'"
            />
          )}

          <TextField
            fullWidth
            label="Rate Limit (hours)"
            type="number"
            value={editRateLimit}
            onChange={(e) => setEditRateLimit(e.target.value)}
            sx={{ mt: 2 }}
          />

          <FormControlLabel
            control={
              <Switch
                checked={editEnabled}
                onChange={(e) => setEditEnabled(e.target.checked)}
              />
            }
            label="Enabled"
            sx={{ mt: 2 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenEdit(false)}>Cancel</Button>
          <Button variant="contained" onClick={handleEdit}>Save</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openDeleteConfirm} onClose={() => setOpenDeleteConfirm(false)}>
        <DialogTitle>Delete Alert Rule</DialogTitle>
        <DialogContent>
          Are you sure you want to delete the alert rule for sensor{' '}
          <strong>{deleteTarget?.SensorName || selectedRow?.SensorName}</strong>?
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDeleteConfirm(false)}>Cancel</Button>
          <Button variant="contained" color="error" onClick={confirmDelete}>
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </PageContainer>
  );
}
