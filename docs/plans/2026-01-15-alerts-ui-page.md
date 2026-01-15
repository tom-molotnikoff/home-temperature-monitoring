# Alerts UI Page Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Create a full CRUD management UI page for sensor alert rules with DataGrid display, create/edit/delete dialogs, alert history viewing, and RBAC permission controls.

**Architecture:** React component following UsersPage.tsx pattern with Material UI DataGrid, dynamic form dialogs that adapt to alert type (numeric_range vs status_based), smart sensor filtering (only show sensors without rules when creating), and permission-based UI disabling. Integrates with existing alert management API endpoints.

**Tech Stack:** React, TypeScript, Material UI, DataGrid, existing API client patterns, AuthContext for permissions

---

## Task 1: Create Alert API Client

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/api/Alerts.ts`

**Step 1: Create the API client file**

Create the file with complete type definitions and API methods:

```typescript
import { get, post, put, del, type ApiMessage } from './Client';

export type AlertRule = {
  ID: number;
  SensorID: number;
  SensorName: string;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold: number | null;
  LowThreshold: number | null;
  TriggerStatus: string;
  Enabled: boolean;
  RateLimitHours: number;
  LastAlertSentAt: string | null;
};

export type AlertHistory = {
  id: number;
  sensor_id: number;
  alert_type: string;
  reading_value: string;
  sent_at: string;
};

export type CreateAlertRuleRequest = {
  SensorID: number;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitHours: number;
  Enabled: boolean;
};

export type UpdateAlertRuleRequest = {
  SensorID: number;
  AlertType: 'numeric_range' | 'status_based';
  HighThreshold?: number;
  LowThreshold?: number;
  TriggerStatus?: string;
  RateLimitHours: number;
  Enabled: boolean;
};

export const listAlertRules = () => get<AlertRule[]>('/alerts');

export const getAlertRule = (sensorId: number) => get<AlertRule>(`/alerts/${sensorId}`);

export const createAlertRule = (rule: CreateAlertRuleRequest) => post<ApiMessage>('/alerts', rule);

export const updateAlertRule = (sensorId: number, rule: UpdateAlertRuleRequest) => put<ApiMessage>(`/alerts/${sensorId}`, rule);

export const deleteAlertRule = (sensorId: number) => del<ApiMessage>(`/alerts/${sensorId}`);

export const getAlertHistory = (sensorId: number, limit?: number) => {
  const params = limit ? `?limit=${limit}` : '';
  return get<AlertHistory[]>(`/alerts/${sensorId}/history${params}`);
};
```

**Step 2: Verify the file structure**

Run: `ls -la sensor_hub/ui/sensor_hub_ui/src/api/Alerts.ts`
Expected: File exists with correct content

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/api/Alerts.ts
git commit -m "feat(ui): add alerts API client

- Add AlertRule, AlertHistory types matching backend API
- Add CreateAlertRuleRequest, UpdateAlertRuleRequest types
- Implement CRUD operations for alert rules
- Add alert history fetching with optional limit"
```

---

## Task 2: Create AlertsPage Component Structure

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Create the basic page structure**

Create file with imports and basic component structure:

```typescript
import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef, GridRowParams } from '@mui/x-data-grid';
import {
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Box,
  Menu,
  MenuItem,
  Select,
  InputLabel,
  FormControl,
  Typography,
  FormControlLabel,
  Switch,
} from '@mui/material';
import { listAlertRules, createAlertRule, updateAlertRule, deleteAlertRule, getAlertHistory } from '../../api/Alerts';
import type { AlertRule, AlertHistory, CreateAlertRuleRequest } from '../../api/Alerts';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import { useAuth } from "../../providers/AuthContext.tsx";
import { hasPerm } from "../../tools/Utils.ts";
import { useSensorContext } from "../../hooks/useSensorContext.ts";

export default function AlertsPage() {
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<AlertRule | null>(null);
  
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
```

**Step 2: Verify component compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds without TypeScript errors

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): add AlertsPage basic structure

- Create page with DataGrid showing alert rules
- Add columns for all alert rule properties
- Add permission checks for view_alerts and manage_alerts
- Add row click handler and context menu skeleton
- Format display values (thresholds, enabled, last alert)"
```

---

## Task 3: Add Create Alert Dialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Add state for create dialog**

After the existing state declarations (around line 30), add:

```typescript
  const [openCreate, setOpenCreate] = useState(false);
  const [createAlertType, setCreateAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [createSensorId, setCreateSensorId] = useState<number>(0);
  const [createHighThreshold, setCreateHighThreshold] = useState<string>('');
  const [createLowThreshold, setCreateLowThreshold] = useState<string>('');
  const [createTriggerStatus, setCreateTriggerStatus] = useState<string>('');
  const [createRateLimit, setCreateRateLimit] = useState<string>('1');
  const [createEnabled, setCreateEnabled] = useState<boolean>(true);
```

**Step 2: Add create handler function**

Before the return statement, add:

```typescript
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
      // Reset form
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

  // Get sensors that don't have alert rules yet
  const availableSensors = sensors.filter(s => !alertRules.some(r => r.SensorID === s.id));
```

**Step 3: Update create button**

Replace the create button with:

```typescript
              <Button variant="contained" disabled={fieldsDisabled} onClick={() => setOpenCreate(true)}>
                Create Alert Rule
              </Button>
```

**Step 4: Add create dialog**

After the closing `</LayoutCard>` tag but before the closing `</PageContainer>` tag, add:

```typescript
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
```

**Step 5: Verify it compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): add create alert rule dialog

- Add dynamic form that shows/hides fields based on alert type
- Filter sensors to only show those without existing rules
- Add validation and form state management
- Submit creates new alert rule via API"
```

---

## Task 4: Add Edit Alert Dialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Add state for edit dialog**

After the create dialog state declarations, add:

```typescript
  const [openEdit, setOpenEdit] = useState(false);
  const [editRule, setEditRule] = useState<AlertRule | null>(null);
  const [editAlertType, setEditAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [editHighThreshold, setEditHighThreshold] = useState<string>('');
  const [editLowThreshold, setEditLowThreshold] = useState<string>('');
  const [editTriggerStatus, setEditTriggerStatus] = useState<string>('');
  const [editRateLimit, setEditRateLimit] = useState<string>('1');
  const [editEnabled, setEditEnabled] = useState<boolean>(true);
```

**Step 2: Add edit handler functions**

Before the return statement, after the handleCreate function, add:

```typescript
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
```

**Step 3: Update menu Edit item**

Replace the disabled Edit menu item with:

```typescript
              <MenuItem disabled={fieldsDisabled} onClick={handleOpenEdit}>Edit</MenuItem>
```

**Step 4: Add edit dialog**

After the create dialog, add:

```typescript
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
```

**Step 5: Verify it compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): add edit alert rule dialog

- Populate form with existing rule values
- Sensor field disabled (can't change which sensor)
- Dynamic form based on alert type
- Submit updates alert rule via API"
```

---

## Task 5: Add Delete Confirmation Dialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Add state for delete dialog**

After the edit dialog state declarations, add:

```typescript
  const [openDeleteConfirm, setOpenDeleteConfirm] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<AlertRule | null>(null);
```

**Step 2: Add delete handler functions**

After the handleEdit function, add:

```typescript
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
```

**Step 3: Update menu Delete item**

Replace the disabled Delete menu item with:

```typescript
              <MenuItem disabled={fieldsDisabled} onClick={handleOpenDelete}>Delete</MenuItem>
```

**Step 4: Add delete confirmation dialog**

After the edit dialog, add:

```typescript
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
```

**Step 5: Verify it compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): add delete alert rule confirmation

- Add confirmation dialog before deletion
- Show sensor name in confirmation message
- Delete rule via API on confirmation"
```

---

## Task 6: Add Alert History Dialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Add state for history dialog**

After the delete dialog state declarations, add:

```typescript
  const [openHistory, setOpenHistory] = useState(false);
  const [historyData, setHistoryData] = useState<AlertHistory[]>([]);
  const [historyLoading, setHistoryLoading] = useState(false);
```

**Step 2: Add history handler functions**

After the confirmDelete function, add:

```typescript
  const handleOpenHistory = async () => {
    if (!selectedRow) return;
    setHistoryLoading(true);
    setOpenHistory(true);
    handleMenuClose();
    try {
      const history = await getAlertHistory(selectedRow.SensorID, 50);
      setHistoryData(history);
    } catch (e) {
      console.error('Failed to load alert history', e);
      setHistoryData([]);
    } finally {
      setHistoryLoading(false);
    }
  };
```

**Step 3: Update menu View History item**

Replace the View History menu item with:

```typescript
              <MenuItem onClick={handleOpenHistory}>View History</MenuItem>
```

**Step 4: Add history dialog**

After the delete dialog, add:

```typescript
      <Dialog open={openHistory} onClose={() => setOpenHistory(false)} maxWidth="md" fullWidth>
        <DialogTitle>Alert History - {selectedRow?.SensorName}</DialogTitle>
        <DialogContent>
          {historyLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
              Loading...
            </Box>
          ) : historyData.length === 0 ? (
            <Box sx={{ p: 2 }}>
              <Typography>No alert history found for this sensor.</Typography>
            </Box>
          ) : (
            <Box sx={{ mt: 1 }}>
              {historyData.map((h) => (
                <Box
                  key={h.id}
                  sx={{
                    p: 2,
                    mb: 1,
                    border: '1px solid',
                    borderColor: 'divider',
                    borderRadius: 1,
                  }}
                >
                  <Typography variant="body2">
                    <strong>Type:</strong> {h.alert_type}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Value:</strong> {h.reading_value}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    <strong>Sent:</strong> {new Date(h.sent_at).toLocaleString()}
                  </Typography>
                </Box>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenHistory(false)}>Close</Button>
        </DialogActions>
      </Dialog>
```

**Step 5: Verify it compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): add alert history viewer

- Fetch and display alert history for selected sensor
- Show alert type, reading value, and timestamp
- Handle loading and empty states
- Limit to 50 most recent alerts"
```

---

## Task 7: Add Route to Navigation

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx`

**Step 1: Add import and route in AppRoutes.tsx**

Add import at the top with other page imports:

```typescript
import AlertsPage from "../pages/alerts/AlertsPage.tsx";
```

Add route after the RolesPage route (around line 25):

```typescript
        <Route path="/alerts" element={<RequireAuth><AlertsPage /></RequireAuth>} />
```

**Step 2: Verify routes file compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 3: Add navigation item in NavigationSidebar.tsx**

Add icon import at top:

```typescript
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';
```

Add navigation item after the Properties item (around line 103):

```typescript
        { (hasPerm(user, 'view_alerts') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/alerts')}>
              <ListItemIcon><NotificationsActiveIcon /></ListItemIcon>
              <ListItemText primary="Alerts" />
            </ListItemButton>
          </ListItem>
        ))}
```

**Step 4: Verify sidebar compiles**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/navigation/AppRoutes.tsx sensor_hub/ui/sensor_hub_ui/src/navigation/NavigationSidebar.tsx
git commit -m "feat(ui): add alerts page to navigation

- Add /alerts route with auth requirement
- Add Alerts menu item to sidebar with view_alerts permission
- Use NotificationsActive icon for alerts menu item"
```

---

## Task 8: Manual Testing

**Files:**
- No file changes

**Step 1: Start the development server**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run dev`
Expected: Dev server starts without errors

**Step 2: Test viewing alerts page**

1. Login as user with `view_alerts` permission
2. Click "Alerts" in sidebar
3. Verify page loads with DataGrid
4. Verify Create button is disabled if no `manage_alerts` permission

Expected: Page displays, permission checks work

**Step 3: Test creating alert rule**

1. Login as user with `manage_alerts` permission
2. Click "Create Alert Rule"
3. Select sensor, set alert type to "numeric_range"
4. Set high threshold = 30, low threshold = 10
5. Set rate limit = 1 hour
6. Click Create

Expected: Alert rule created, appears in grid

**Step 4: Test editing alert rule**

1. Click on a row with an alert rule
2. Click "Edit" from menu
3. Change threshold values
4. Click Save

Expected: Alert rule updated in grid

**Step 5: Test viewing history**

1. Click on a row with an alert rule
2. Click "View History" from menu
3. Verify history dialog shows (may be empty)

Expected: Dialog opens, handles empty state gracefully

**Step 6: Test deleting alert rule**

1. Click on a row with an alert rule
2. Click "Delete" from menu
3. Confirm deletion

Expected: Alert rule removed from grid

**Step 7: Test dynamic form behavior**

1. Click "Create Alert Rule"
2. Change Alert Type from "numeric_range" to "status_based"
3. Verify High/Low Threshold fields hide and Trigger Status field shows
4. Change back to "numeric_range"
5. Verify fields switch back

Expected: Form fields dynamically show/hide based on alert type

**Step 8: Stop dev server**

Press Ctrl+C to stop the dev server

---

## Task 9: Final Commit and Documentation

**Files:**
- Modify: `docs/alerting-system.md`

**Step 1: Update documentation**

In `docs/alerting-system.md`, find the "Future Enhancements" section (around line 299) and update it:

```markdown
## Future Enhancements

Possible improvements for future iterations:
- ~~Web UI for managing alert rules~~ ✅ **Implemented** - See /alerts page
- Multiple notification channels (SMS, Slack, etc.)
- Alert templates per sensor type
- Configurable alert messages
- Alert acknowledgment system
- Grouped/batched alerts for multiple sensors
```

**Step 2: Commit documentation update**

```bash
git add docs/alerting-system.md
git commit -m "docs: mark alerts UI as implemented

Update alerting-system.md to reflect completed UI implementation"
```

**Step 3: Create final summary commit**

```bash
git log --oneline | head -8
```

Expected: Shows all commits from this implementation

---

## Summary

This plan implements a complete CRUD management UI for sensor alert rules:

✅ **API Client** - Type-safe client for all alert endpoints
✅ **DataGrid Display** - Shows all alert rule properties with proper formatting
✅ **Create Dialog** - Dynamic form for creating new rules with sensor filtering
✅ **Edit Dialog** - Update existing rules with pre-populated values
✅ **Delete Confirmation** - Safe deletion with user confirmation
✅ **Alert History** - View past alerts for each sensor
✅ **Navigation** - Integrated into sidebar with permission checks
✅ **Permissions** - Respects view_alerts and manage_alerts RBAC permissions

The implementation follows existing patterns from UsersPage.tsx and RolesPage.tsx, uses Material UI components consistently, and integrates seamlessly with the existing authentication and sensor context providers.
