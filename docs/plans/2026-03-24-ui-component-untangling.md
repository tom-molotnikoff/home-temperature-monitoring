# UI Component Untangling Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Fix broken dialog state management in the extracted AlertsPage components, then continue the untangling pattern by extracting UsersPage's inline dialogs into standalone components.

**Architecture:** Every page file should be a clean composition of imported components (like `TemperatureDashboard.tsx`). Dialog components receive `open: boolean` and `onClose: () => void` from the parent — no internal mirroring of open state. Parent fully controls dialog visibility. Action callbacks (`onCreated`, `onSaved`, `onDeleted`) are passed as props; dialogs call `onClose()` after completing their action.

**Tech Stack:** React 18+, TypeScript, MUI (Material UI), Vite HMR via Docker (`docker_tests` stack on `http://localhost:3000`), `playwright-cli` for browser verification.

---

## Bug Context

The extracted AlertsPage dialog components use a broken pattern:
- Parent passes `openXDialog={true}` as a prop
- Child mirrors it into internal `useState` via `useEffect`
- When child closes itself, parent's state remains `true`
- Next render cycle, `useEffect` fires again → dialog reopens

**Observed:** Click row → Edit → close → Click "Create Alert Rule" → BOTH Edit AND Create dialogs open simultaneously.

---

### Task 1: Fix AlertHistoryDialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/AlertHistoryDialog.tsx`

**Step 1: Replace the entire file with the fixed version**

```tsx
import {Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography} from "@mui/material";
import {useEffect, useState} from "react";
import type {AlertHistory, AlertRule} from "../api/Alerts";
import { getAlertHistory } from "../api/Alerts";

interface AlertHistoryDialogProps {
  open: boolean;
  onClose: () => void;
  selectedAlert: AlertRule | null;
}

export default function AlertHistoryDialog({open, onClose, selectedAlert}: AlertHistoryDialogProps) {
  const [historyLoading, setHistoryLoading] = useState(false);
  const [historyData, setHistoryData] = useState<AlertHistory[]>([]);

  useEffect(() => {
    if (!open || !selectedAlert) return;
    let cancelled = false;
    const fetchHistory = async () => {
      setHistoryLoading(true);
      try {
        const history = await getAlertHistory(selectedAlert.SensorID, 50);
        if (!cancelled) setHistoryData(history);
      } catch (e) {
        console.error('Failed to load alert history', e);
        if (!cancelled) setHistoryData([]);
      } finally {
        if (!cancelled) setHistoryLoading(false);
      }
    };
    fetchHistory();
    return () => { cancelled = true; };
  }, [open, selectedAlert]);

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Alert History - {selectedAlert?.SensorName}</DialogTitle>
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
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the change**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors in console.

---

### Task 2: Fix CreateAlertDialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/CreateAlertDialog.tsx`

**Step 1: Replace the entire file with the fixed version**

```tsx
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl, FormControlLabel,
  InputLabel,
  MenuItem,
  Select, Switch,
  TextField
} from "@mui/material";
import {createAlertRule, type CreateAlertRuleRequest} from "../api/Alerts.ts";
import {useState} from "react";
import type { AlertRule } from "../api/Alerts";
import {useSensorContext} from "../hooks/useSensorContext.ts";

interface CreateAlertDialogProps {
  open: boolean;
  onClose: () => void;
  onCreated: () => Promise<void>;
  alertRules: AlertRule[];
}

export default function CreateAlertDialog({open, onClose, onCreated, alertRules}: CreateAlertDialogProps) {
  const [createAlertType, setCreateAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [createSensorId, setCreateSensorId] = useState<number>(0);
  const [createHighThreshold, setCreateHighThreshold] = useState<string>('');
  const [createLowThreshold, setCreateLowThreshold] = useState<string>('');
  const [createTriggerStatus, setCreateTriggerStatus] = useState<string>('');
  const [createRateLimit, setCreateRateLimit] = useState<string>('1');
  const [createEnabled, setCreateEnabled] = useState<boolean>(true);
  const { sensors } = useSensorContext();

  const resetForm = () => {
    setCreateSensorId(0);
    setCreateAlertType('numeric_range');
    setCreateHighThreshold('');
    setCreateLowThreshold('');
    setCreateTriggerStatus('');
    setCreateRateLimit('1');
    setCreateEnabled(true);
  };

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
      resetForm();
      onClose();
      await onCreated();
    } catch (e) {
      console.error('Failed to create alert rule', e);
    }
  };

  const handleCancel = () => {
    resetForm();
    onClose();
  };

  const availableSensors = sensors.filter(s => !alertRules.some(r => r.SensorID === s.id));

  return (
    <Dialog open={open} onClose={handleCancel} maxWidth="sm" fullWidth>
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
        <Button onClick={handleCancel}>Cancel</Button>
        <Button variant="contained" onClick={handleCreate}>Create</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the change**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors in console.

---

### Task 3: Fix DeleteAlertDialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/DeleteAlertDialog.tsx`

**Step 1: Replace the entire file with the fixed version**

```tsx
import {Button, Dialog, DialogActions, DialogContent, DialogTitle} from "@mui/material";
import type {AlertRule} from "../api/Alerts.ts";
import { deleteAlertRule } from "../api/Alerts.ts";

interface DeleteAlertDialogProps {
  open: boolean;
  onClose: () => void;
  onDeleted: () => Promise<void>;
  selectedAlert: AlertRule | null;
}

export default function DeleteAlertDialog({open, onClose, onDeleted, selectedAlert}: DeleteAlertDialogProps) {
  const confirmDelete = async () => {
    if (!selectedAlert) return;
    try {
      await deleteAlertRule(selectedAlert.SensorID);
      onClose();
      await onDeleted();
    } catch (e) {
      console.error('Failed to delete alert rule', e);
    }
  };

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Delete Alert Rule</DialogTitle>
      <DialogContent>
        Are you sure you want to delete the alert rule for sensor{' '}
        <strong>{selectedAlert?.SensorName}</strong>?
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" color="error" onClick={confirmDelete}>
          Delete
        </Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the change**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors in console.

---

### Task 4: Fix EditAlertDialog

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/EditAlertDialog.tsx`

**Step 1: Replace the entire file with the fixed version**

```tsx
import {useEffect, useState} from "react";
import {type AlertRule, type CreateAlertRuleRequest, updateAlertRule} from "../api/Alerts.ts";
import {
  Button,
  Dialog, DialogActions,
  DialogContent, DialogTitle,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Select, Switch,
  TextField
} from "@mui/material";

interface EditAlertDialogProps {
  open: boolean;
  onClose: () => void;
  onSaved: () => Promise<void>;
  selectedAlert: AlertRule | null;
}

export default function EditAlertDialog({open, onClose, onSaved, selectedAlert}: EditAlertDialogProps) {
  const [editAlertType, setEditAlertType] = useState<'numeric_range' | 'status_based'>('numeric_range');
  const [editHighThreshold, setEditHighThreshold] = useState<string>('');
  const [editLowThreshold, setEditLowThreshold] = useState<string>('');
  const [editTriggerStatus, setEditTriggerStatus] = useState<string>('');
  const [editRateLimit, setEditRateLimit] = useState<string>('1');
  const [editEnabled, setEditEnabled] = useState<boolean>(true);

  useEffect(() => {
    if (open && selectedAlert) {
      setEditAlertType(selectedAlert.AlertType);
      setEditHighThreshold(selectedAlert.HighThreshold?.toString() || '');
      setEditLowThreshold(selectedAlert.LowThreshold?.toString() || '');
      setEditTriggerStatus(selectedAlert.TriggerStatus || '');
      setEditRateLimit(selectedAlert.RateLimitHours.toString());
      setEditEnabled(selectedAlert.Enabled);
    }
  }, [open, selectedAlert]);

  const handleEdit = async () => {
    if (!selectedAlert) return;
    try {
      const request: CreateAlertRuleRequest = {
        SensorID: selectedAlert.SensorID,
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

      await updateAlertRule(selectedAlert.SensorID, request);
      onClose();
      await onSaved();
    } catch (e) {
      console.error('Failed to update alert rule', e);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Alert Rule</DialogTitle>
      <DialogContent>
        <TextField
          fullWidth
          label="Sensor"
          value={selectedAlert?.SensorName || ''}
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
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={handleEdit}>Save</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the change**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors in console.
### Task 5: Update AlertsPage wiring

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Replace the entire file with the fixed version**

```tsx
import { useEffect, useState } from 'react';
import type { GridRowParams } from '@mui/x-data-grid';
import {
  Button,
  Box,
  Menu,
  MenuItem,
  Typography,
} from '@mui/material';
import AlertRuleCard from '../../components/AlertRuleCard';
import { listAlertRules } from '../../api/Alerts';
import type { AlertRule } from '../../api/Alerts';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import { useAuth } from "../../providers/AuthContext.tsx";
import { hasPerm } from "../../tools/Utils.ts";
import { useIsMobile } from '../../hooks/useMobile';
import AlertRuleDataGrid from "../../components/AlertRuleDataGrid.tsx";
import AlertHistoryDialog from "../../components/AlertHistoryDialog.tsx";
import DeleteAlertDialog from "../../components/DeleteAlertDialog.tsx";
import EditAlertDialog from "../../components/EditAlertDialog.tsx";
import CreateAlertDialog from "../../components/CreateAlertDialog.tsx";

export default function AlertsPage() {
  const [alertRules, setAlertRules] = useState<AlertRule[]>([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<AlertRule | null>(null);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openHistoryDialog, setOpenHistoryDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);

  const { user } = useAuth();
  const isMobile = useIsMobile();

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

  const closeMenu = () => { setMenuAnchorEl(null); };

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
              <Button variant="contained" disabled={fieldsDisabled} onClick={() => setOpenCreateDialog(true)}>
                Create Alert Rule
              </Button>
            </Box>
          </Box>
          {isMobile ? (
            <Box sx={{ width: '100%', maxHeight: 400, overflowY: 'auto' }}>
              {alertRules.length === 0 ? (
                <Typography color="text.secondary" sx={{ p: 2, textAlign: 'center' }}>
                  No alert rules configured
                </Typography>
              ) : (
                alertRules.map((rule) => (
                  <AlertRuleCard
                    key={rule.SensorID}
                    rule={rule}
                    onClick={(event) => {
                      setSelectedRow(rule);
                      setMenuAnchorEl(event.currentTarget as HTMLElement);
                    }}
                  />
                ))
              )}
            </Box>
          ) : (
            <div style={{ height: 400, width: '100%' }}>
              <AlertRuleDataGrid
                alertRules={alertRules}
                handleRowClick={handleRowClick}
              />
            </div>
          )}

          {hasPerm(user, "view_alerts") && (
            <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={closeMenu}>
              <MenuItem disabled={fieldsDisabled} onClick={() => { closeMenu(); setOpenEditDialog(true); }}>Edit</MenuItem>
              <MenuItem disabled={fieldsDisabled} onClick={() => { closeMenu(); setOpenDeleteDialog(true); }}>Delete</MenuItem>
              <MenuItem onClick={() => { closeMenu(); setOpenHistoryDialog(true); }}>View History</MenuItem>
            </Menu>
          )}
        </LayoutCard>
      </Box>

      <CreateAlertDialog
        open={openCreateDialog}
        onClose={() => setOpenCreateDialog(false)}
        onCreated={load}
        alertRules={alertRules}
      />

      <EditAlertDialog
        open={openEditDialog}
        onClose={() => setOpenEditDialog(false)}
        onSaved={load}
        selectedAlert={selectedRow}
      />

      <DeleteAlertDialog
        open={openDeleteDialog}
        onClose={() => setOpenDeleteDialog(false)}
        onDeleted={load}
        selectedAlert={selectedRow}
      />

      <AlertHistoryDialog
        open={openHistoryDialog}
        onClose={() => setOpenHistoryDialog(false)}
        selectedAlert={selectedRow}
      />
    </PageContainer>
  );
}
```

**Step 2: Wait for HMR and verify no build errors**

Run: `sleep 3 && playwright-cli snapshot`
Expected: Page renders with data grid and "Create Alert Rule" button.

---

### Task 6: Verify AlertsPage in browser

**Step 1: Open browser and navigate to alerts page**

Run: `playwright-cli open http://localhost:3000`
Then login (fill username "admin", password "admin", click Sign in).
Then: `playwright-cli goto http://localhost:3000/alerts`

**Step 2: Verify the data grid renders**

Run: `playwright-cli snapshot`
Expected: Snapshot shows `heading "Alert Rules"`, a `grid` with rows (Downstairs, Upstairs), and `button "Create Alert Rule"`.

**Step 3: Test Edit dialog opens alone**

Run: Click a data grid row, then click "Edit" in the menu.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog "Edit Alert Rule"` visible. No other dialogs.

**Step 4: Close Edit dialog and test Create dialog opens alone**

Run: `playwright-cli press Escape`
Run: Click "Create Alert Rule" button.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog` with heading "Create Alert Rule". No Edit/Delete/History dialogs.

**Step 5: Close Create dialog and test Delete dialog opens alone**

Run: `playwright-cli press Escape`
Run: Click a data grid row, then click "Delete" in the menu.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog "Delete Alert Rule"` visible.

**Step 6: Close Delete dialog and test History dialog opens alone**

Run: `playwright-cli press Escape`
Run: Click a data grid row, then click "View History" in the menu.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog` with title containing "Alert History".

**Step 7: Close browser**

Run: `playwright-cli close`
### Task 7: Extract CreateUserDialog from UsersPage

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/CreateUserDialog.tsx`

**Step 1: Create the component file**

```tsx
import {useState, useEffect} from "react";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
} from "@mui/material";
import {createUser} from "../api/Users.ts";
import {listRoles, type Role} from "../api/Roles.ts";

interface CreateUserDialogProps {
  open: boolean;
  onClose: () => void;
  onCreated: () => Promise<void>;
}

export default function CreateUserDialog({open, onClose, onCreated}: CreateUserDialogProps) {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('user');
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);

  useEffect(() => {
    if (!open) return;
    listRoles().then(r => {
      const roles = r || [];
      setAvailableRoles(roles);
      if (roles.length > 0 && !roles.find(x => x.name === role)) {
        setRole(roles[0].name);
      }
    }).catch(e => console.error('Failed to load roles', e));
  }, [open]);

  const resetForm = () => {
    setUsername('');
    setEmail('');
    setPassword('');
    setRole('user');
  };

  const handleCreate = async () => {
    try {
      await createUser({ username, email, password, roles: [role] });
      resetForm();
      onClose();
      await onCreated();
    } catch (e) {
      console.error('Failed to create user', e);
    }
  };

  const handleCancel = () => {
    resetForm();
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleCancel}>
      <DialogTitle>Create user</DialogTitle>
      <DialogContent>
        <TextField fullWidth label="Username" value={username} onChange={(e) => setUsername(e.target.value)} sx={{mt: 1}}/>
        <TextField fullWidth label="Email" value={email} onChange={(e) => setEmail(e.target.value)} sx={{mt: 1}}/>
        <TextField fullWidth label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} sx={{mt: 1}}/>
        <FormControl fullWidth sx={{mt: 1}}>
          <InputLabel id="role-select-label">Role</InputLabel>
          <Select labelId="role-select-label" value={role} label="Role" onChange={(e) => setRole(e.target.value as string)}>
            {availableRoles.map(r => (<MenuItem key={r.name} value={r.name}>{r.name}</MenuItem>))}
          </Select>
        </FormControl>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel}>Cancel</Button>
        <Button variant="contained" onClick={handleCreate}>Create</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the new file**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors.

---

### Task 8: Extract EditUserDialog from UsersPage

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/EditUserDialog.tsx`

**Step 1: Create the component file**

```tsx
import {useState, useEffect} from "react";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
} from "@mui/material";
import type {User} from "../api/Users.ts";
import {setUserRoles} from "../api/Users.ts";
import {listRoles, type Role} from "../api/Roles.ts";

interface EditUserDialogProps {
  open: boolean;
  onClose: () => void;
  onSaved: () => Promise<void>;
  selectedUser: User | null;
}

export default function EditUserDialog({open, onClose, onSaved, selectedUser}: EditUserDialogProps) {
  const [role, setRole] = useState('user');
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);

  useEffect(() => {
    if (!open) return;
    if (selectedUser?.roles && selectedUser.roles.length > 0) {
      setRole(selectedUser.roles[0]);
    } else {
      setRole('user');
    }
    listRoles().then(r => {
      setAvailableRoles(r || []);
    }).catch(e => console.error('Failed to load roles', e));
  }, [open, selectedUser]);

  const handleSave = async () => {
    if (!selectedUser) return;
    try {
      await setUserRoles(selectedUser.id, [role]);
      onClose();
      await onSaved();
    } catch (e) {
      console.error('Failed to update user roles', e);
    }
  };

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Edit user</DialogTitle>
      <DialogContent>
        <TextField fullWidth label="Username" value={selectedUser?.username ?? ''} disabled sx={{mt: 1}}/>
        <FormControl fullWidth sx={{mt: 2}}>
          <InputLabel id="edit-role-select-label">Role</InputLabel>
          <Select labelId="edit-role-select-label" value={role} label="Role" onChange={(e) => setRole(e.target.value as string)}>
            {availableRoles.map(r => (<MenuItem key={r.name} value={r.name}>{r.name}</MenuItem>))}
          </Select>
        </FormControl>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={handleSave}>Save</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the new file**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors.

---

### Task 9: Extract DeleteUserDialog from UsersPage

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/DeleteUserDialog.tsx`

**Step 1: Create the component file**

```tsx
import {Button, Dialog, DialogActions, DialogContent, DialogTitle} from "@mui/material";
import type {User} from "../api/Users.ts";
import {deleteUser} from "../api/Users.ts";

interface DeleteUserDialogProps {
  open: boolean;
  onClose: () => void;
  onDeleted: () => Promise<void>;
  selectedUser: User | null;
}

export default function DeleteUserDialog({open, onClose, onDeleted, selectedUser}: DeleteUserDialogProps) {
  const confirmDelete = async () => {
    if (!selectedUser) return;
    try {
      await deleteUser(selectedUser.id);
      onClose();
      await onDeleted();
    } catch (e) {
      console.error('Failed to delete user', e);
    }
  };

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Delete user</DialogTitle>
      <DialogContent>
        Are you sure you want to delete user <strong>{selectedUser?.username}</strong>?
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" color="error" onClick={confirmDelete}>Delete</Button>
      </DialogActions>
    </Dialog>
  );
}
```

**Step 2: Verify HMR picks up the new file**

Run: `sleep 3 && playwright-cli snapshot`
Expected: No build errors.
### Task 10: Rewrite UsersPage as clean composition

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/settings/UsersPage.tsx`

**Step 1: Replace the entire file with the clean version**

```tsx
import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef, GridRowParams } from '@mui/x-data-grid';
import {
  Button,
  Box, Grid,
  Menu,
  MenuItem,
  Typography,
} from '@mui/material';
import { listUsers, setMustChange } from '../../api/Users';
import type { User } from '../../api/Users';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import { useAuth } from "../../providers/AuthContext.tsx";
import { hasPerm } from "../../tools/Utils.ts";
import { useIsMobile } from '../../hooks/useMobile';
import CreateUserDialog from "../../components/CreateUserDialog.tsx";
import EditUserDialog from "../../components/EditUserDialog.tsx";
import DeleteUserDialog from "../../components/DeleteUserDialog.tsx";

export default function UsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<User | null>(null);
  const [openCreateDialog, setOpenCreateDialog] = useState(false);
  const [openEditDialog, setOpenEditDialog] = useState(false);
  const [openDeleteDialog, setOpenDeleteDialog] = useState(false);
  const { user } = useAuth();
  const isMobile = useIsMobile();

  const load = async () => {
    try {
      const u = await listUsers();
      setUsers(u);
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    const id = typeof params.id === 'number' ? params.id : Number(params.id);
    const found = users.find(u => u.id === id);
    if (found) setSelectedRow(found);
    else setSelectedRow(params.row as User);
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const closeMenu = () => { setMenuAnchorEl(null); };

  const handleForceChange = async () => {
    if (!selectedRow) return;
    closeMenu();
    await setMustChange(selectedRow.id, true);
    await load();
  };

  const allColumns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 80 },
    { field: 'username', headerName: 'Username', flex: 1 },
    { field: 'email', headerName: 'Email', flex: 1 },
    { field: 'rolesDisplay', headerName: 'Roles', flex: 1 },
    { field: 'must_change_password', headerName: 'Must change password', width: 200 },
  ];

  const mobileHiddenFields = ['id', 'email', 'must_change_password'];
  const columns = isMobile
    ? allColumns.filter(col => !mobileHiddenFields.includes(col.field))
    : allColumns;

  const rows = users.map(u => ({ ...u, rolesDisplay: (u.roles || []).join(', ') }));

  if (user === undefined) {
    return (
      <PageContainer titleText="Users">
        <Box sx={{ flexGrow: 1, width: '100%' }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
            <Grid size={12}>
              Loading...
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

  const fieldsDisabled = !hasPerm(user, "manage_users");

  return (
    <PageContainer titleText="Users">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%", width: "100%" }}>
          <Box sx={{ width: '100%' }}>
            <LayoutCard variant="secondary" changes={{ alignItems: "stretch", height: "100%", width: "100%" }}>
              <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{ width: '100%' }}>
                <Typography variant="h4">Users</Typography>
                <Box>
                  <Button variant="contained" onClick={() => setOpenCreateDialog(true)} disabled={fieldsDisabled}>Create user</Button>
                </Box>
              </Box>
              <div style={{ height: 400, width: '100%' }}>
                <DataGrid rows={rows} columns={columns} pageSizeOptions={[5, 10, 25]} initialState={{ pagination: { paginationModel: { pageSize: 5 } } }} onRowClick={handleRowClick}/>
              </div>

              {hasPerm(user, "manage_users") && (
                <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={closeMenu}>
                  <MenuItem onClick={() => { closeMenu(); setOpenEditDialog(true); }}>Edit</MenuItem>
                  <MenuItem onClick={() => { closeMenu(); setOpenDeleteDialog(true); }}>Delete</MenuItem>
                  <MenuItem onClick={handleForceChange}>Force change password</MenuItem>
                </Menu>
              )}
            </LayoutCard>
          </Box>
        </Grid>
      </Box>

      <CreateUserDialog
        open={openCreateDialog}
        onClose={() => setOpenCreateDialog(false)}
        onCreated={load}
      />

      <EditUserDialog
        open={openEditDialog}
        onClose={() => setOpenEditDialog(false)}
        onSaved={load}
        selectedUser={selectedRow}
      />

      <DeleteUserDialog
        open={openDeleteDialog}
        onClose={() => setOpenDeleteDialog(false)}
        onDeleted={load}
        selectedUser={selectedRow}
      />
    </PageContainer>
  );
}
```

**Step 2: Wait for HMR and verify no build errors**

Run: `sleep 3 && playwright-cli snapshot`
Expected: Page renders with data grid and "Create user" button.

---

### Task 11: Verify UsersPage in browser

**Step 1: Navigate to Users page**

Run: `playwright-cli goto http://localhost:3000/settings/users`
Run: `sleep 3 && playwright-cli snapshot`
Expected: Snapshot shows `heading "Users"`, a `grid` with user rows, and `button "Create user"`.

**Step 2: Test Edit dialog opens alone**

Run: Click a data grid row, then click "Edit" in the menu.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog "Edit user"` visible. No other dialogs.

**Step 3: Close Edit dialog and test Create dialog opens alone**

Run: `playwright-cli press Escape`
Run: Click "Create user" button.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog "Create user"` visible. No Edit/Delete dialogs.

**Step 4: Close Create dialog and test Delete dialog opens alone**

Run: `playwright-cli press Escape`
Run: Click a data grid row, then click "Delete" in the menu.
Run: `playwright-cli snapshot`
Expected: Exactly ONE `dialog "Delete user"` visible.

**Step 5: Close browser**

Run: `playwright-cli press Escape`
Run: `playwright-cli close`

---

### Task 12: Final verification — all pages render

**Step 1: Open browser and login**

Run: `playwright-cli open http://localhost:3000`
Login with admin/admin.

**Step 2: Navigate to each page and verify no errors**

Run: `playwright-cli goto http://localhost:3000/` → snapshot → verify dashboard renders.
Run: `playwright-cli goto http://localhost:3000/alerts` → snapshot → verify alerts grid renders.
Run: `playwright-cli goto http://localhost:3000/settings/users` → snapshot → verify users grid renders.
Run: `playwright-cli goto http://localhost:3000/settings/sessions` → snapshot → verify sessions grid renders.
Run: `playwright-cli goto http://localhost:3000/settings/roles` → snapshot → verify roles page renders.

**Step 3: Check console for errors on each page**

Expected: No React errors, no TypeScript compilation errors, no missing import errors.

**Step 4: Close browser**

Run: `playwright-cli close`
