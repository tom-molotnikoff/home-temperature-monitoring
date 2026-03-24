# Page Component Consistency Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Make every PageContainer page a simple composition of self-contained card components, matching the TemperatureDashboard/SensorsOverview pattern — so that moving a card between pages is a 2-3 line cut-and-paste.

**Architecture:** Each page file becomes a thin shell: `<PageContainer>` wrapping a `<Grid container>` of `<Grid>` items, each containing one self-contained card component. All state, data fetching, dialogs, menus, and event handlers move into the card components. Repeated boilerplate (loading state) is DRYed into PageContainer. Custom hooks are extracted where warranted.

**Tech Stack:** React 19, TypeScript, MUI v7, Vite HMR

---

## Pages Already Conforming (no changes needed)

- `src/pages/temperature-dashboard/TemperatureDashboard.tsx` — ✅ Gold standard
- `src/pages/sensors-overview/SensorsOverview.tsx` — ✅ Clean composition
- `src/pages/sensor/SensorPage.tsx` — ✅ Uses imported components

## Pages Requiring Refactoring

| Page | Current Lines | Target Lines | Card to Extract |
|------|--------------|--------------|-----------------|
| AlertsPage | 150 | ~25 | AlertRulesCard |
| UsersPage | 143 | ~25 | UserManagementCard |
| SessionsPage | 116 | ~15 | SessionsCard |
| RolesPage | 143 | ~20 | RolePermissionsCard |
| OAuthPage | 307 | ~15 | OAuthConfigCard |
| NotificationsPage | 230 | ~15 | NotificationsCard |
| NotificationPreferencesPage | 160 | ~15 | NotificationPreferencesCard |
| PropertiesOverview | 149 | ~20 | PropertiesCard |
| ChangePassword | 71 | ~15 | ChangePasswordCard |

## DRY: Loading State Pattern

Currently copy-pasted ~8 times across pages:
```tsx
if (user === undefined) {
  return (
    <PageContainer titleText="X">
      <Box sx={{ flexGrow: 1, width: '100%' }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%" }}>
          <Grid size={12}>Loading...</Grid>
        </Grid>
      </Box>
    </PageContainer>
  );
}
```
**Fix:** Add optional `loading` prop to `PageContainer`.

---

## Task List

1. **Enhance PageContainer** — Add `loading` prop to DRY up the loading state boilerplate
2. **Extract AlertRulesCard** — Move all alert rules state/logic/dialogs into a self-contained card
3. **Rewrite AlertsPage** — Thin shell composing AlertRulesCard
4. **Extract UserManagementCard** — Move all user management state/logic/dialogs into a self-contained card
5. **Rewrite UsersPage** — Thin shell composing UserManagementCard
6. **Extract SessionsCard** — Move session DataGrid, columns, revoke logic into a self-contained card
7. **Rewrite SessionsPage** — Thin shell composing SessionsCard
8. **Extract RolePermissionsCard** — Move roles list, permissions panel, toggle logic into a self-contained card
9. **Rewrite RolesPage** — Thin shell composing RolePermissionsCard
10. **Extract OAuthConfigCard** — Move OAuth status display, authorization flow, code dialog into a self-contained card
11. **Rewrite OAuthPage** — Thin shell composing OAuthConfigCard
12. **Extract NotificationsCard** — Move notification list, tabs, menu, bulk actions into a self-contained card
13. **Rewrite NotificationsPage** — Thin shell composing NotificationsCard
14. **Extract NotificationPreferencesCard** — Move preferences table, toggle logic into a self-contained card
15. **Rewrite NotificationPreferencesPage** — Thin shell composing NotificationPreferencesCard
16. **Extract PropertiesCard** — Move properties form, dirty tracking, save logic into a self-contained card
17. **Rewrite PropertiesOverview** — Thin shell composing PropertiesCard
18. **Extract ChangePasswordCard** — Move password form, validation, submit logic into a self-contained card
19. **Rewrite ChangePassword** — Thin shell composing ChangePasswordCard
20. **Update conforming pages** — Update TemperatureDashboard, SensorsOverview, SensorPage to use `loading` prop
21. **Browser verification** — Verify all pages render and function correctly

---

## Task Details

### Task 1: Enhance PageContainer with `loading` prop

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/tools/PageContainer.tsx`

**Step 1: Add `loading` prop to PageContainer**

Replace the full file content with:

```tsx
import { type CSSProperties } from "react";
import { Box, CircularProgress } from "@mui/material";
import LayoutCard from "./LayoutCard.tsx";
import NavigationSidebar from "../navigation/NavigationSidebar.tsx";
import TopAppBar from "../navigation/TopAppBar.tsx";

interface PageContainerProps {
  children: React.ReactNode;
  titleText: string;
  loading?: boolean;
}

function PageContainer({ children, titleText, loading = false }: PageContainerProps) {
  return (
    <>
      <TopAppBar pageTitle={titleText} />
      <NavigationSidebar />
      <LayoutCard changes={layoutCardStyleChanges}>
        {loading ? (
          <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center", flexGrow: 1, p: 4 }}>
            <CircularProgress />
          </Box>
        ) : (
          children
        )}
      </LayoutCard>
    </>
  );
}

const layoutCardStyleChanges: CSSProperties = {
  padding: "20px",
  alignItems: "stretch",
  minHeight: "calc(100vh - 65px)",
  border: "none",
  borderRadius: 0,
  width: "100%",
};

export default PageContainer;
```

**Step 2: Verify HMR picks up the change**

Run: `touch sensor_hub/ui/sensor_hub_ui/src/tools/PageContainer.tsx`

Verify in browser that existing pages still render (the `loading` prop defaults to `false`, so no page should change behavior).

---

### Task 2: Extract AlertRulesCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/AlertRulesCard.tsx`

**Step 1: Create AlertRulesCard component**

This moves ALL state, handlers, dialogs, and DataGrid/mobile rendering out of AlertsPage into a self-contained card. The card manages its own data loading, menu, selection, and dialog visibility.

```tsx
import { useEffect, useState } from 'react';
import type { GridRowParams } from '@mui/x-data-grid';
import { Button, Box, Menu, MenuItem, Typography } from '@mui/material';
import { listAlertRules } from '../api/Alerts';
import type { AlertRule } from '../api/Alerts';
import LayoutCard from '../tools/LayoutCard';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { useIsMobile } from '../hooks/useMobile';
import AlertRuleDataGrid from './AlertRuleDataGrid';
import AlertRuleCard from './AlertRuleCard';
import AlertHistoryDialog from './AlertHistoryDialog';
import DeleteAlertDialog from './DeleteAlertDialog';
import EditAlertDialog from './EditAlertDialog';
import CreateAlertDialog from './CreateAlertDialog';

export default function AlertRulesCard() {
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

  useEffect(() => { load(); }, []);

  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    const id = typeof params.id === 'number' ? params.id : Number(params.id);
    const found = alertRules.find(r => r.SensorID === id);
    if (found) setSelectedRow(found);
    else setSelectedRow(params.row as AlertRule);
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const closeMenu = () => { setMenuAnchorEl(null); };

  const fieldsDisabled = !user || !hasPerm(user, "manage_alerts");

  return (
    <>
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
            <AlertRuleDataGrid alertRules={alertRules} handleRowClick={handleRowClick} />
          </div>
        )}

        {user && hasPerm(user, "view_alerts") && (
          <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={closeMenu}>
            <MenuItem disabled={fieldsDisabled} onClick={() => { closeMenu(); setOpenEditDialog(true); }}>Edit</MenuItem>
            <MenuItem disabled={fieldsDisabled} onClick={() => { closeMenu(); setOpenDeleteDialog(true); }}>Delete</MenuItem>
            <MenuItem onClick={() => { closeMenu(); setOpenHistoryDialog(true); }}>View History</MenuItem>
          </Menu>
        )}
      </LayoutCard>

      <CreateAlertDialog open={openCreateDialog} onClose={() => setOpenCreateDialog(false)} onCreated={load} alertRules={alertRules} />
      <EditAlertDialog open={openEditDialog} onClose={() => setOpenEditDialog(false)} onSaved={load} selectedAlert={selectedRow} />
      <DeleteAlertDialog open={openDeleteDialog} onClose={() => setOpenDeleteDialog(false)} onDeleted={load} selectedAlert={selectedRow} />
      <AlertHistoryDialog open={openHistoryDialog} onClose={() => setOpenHistoryDialog(false)} selectedAlert={selectedRow} />
    </>
  );
}
```

---

### Task 3: Rewrite AlertsPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Replace AlertsPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import AlertRulesCard from '../../components/AlertRulesCard';

export default function AlertsPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Alert Rules" loading={user === undefined}>
      <AlertRulesCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to `/alerts`. Confirm:
- DataGrid renders with alert rules
- "Create Alert Rule" button works (opens dialog alone)
- Row click → context menu → Edit/Delete/View History each open their dialog alone
- Mobile view shows cards instead of DataGrid

---

### Task 4: Extract UserManagementCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/UserManagementCard.tsx`

**Step 1: Create UserManagementCard component**

Moves all user state, DataGrid, columns, menu, dialogs, and "force change password" action out of UsersPage.

```tsx
import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef, GridRowParams } from '@mui/x-data-grid';
import { Button, Box, Menu, MenuItem, Typography } from '@mui/material';
import { listUsers, setMustChange } from '../api/Users';
import type { User } from '../api/Users';
import LayoutCard from '../tools/LayoutCard';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { useIsMobile } from '../hooks/useMobile';
import CreateUserDialog from './CreateUserDialog';
import EditUserDialog from './EditUserDialog';
import DeleteUserDialog from './DeleteUserDialog';

export default function UserManagementCard() {
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

  useEffect(() => { load(); }, []);

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
  const fieldsDisabled = !user || !hasPerm(user, "manage_users");

  return (
    <>
      <LayoutCard variant="secondary" changes={{ alignItems: "stretch", height: "100%", width: "100%" }}>
        <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{ width: '100%' }}>
          <Typography variant="h4">Users</Typography>
          <Box>
            <Button variant="contained" onClick={() => setOpenCreateDialog(true)} disabled={fieldsDisabled}>Create user</Button>
          </Box>
        </Box>
        <div style={{ height: 400, width: '100%' }}>
          <DataGrid rows={rows} columns={columns} pageSizeOptions={[5, 10, 25]} initialState={{ pagination: { paginationModel: { pageSize: 5 } } }} onRowClick={handleRowClick} />
        </div>

        {user && hasPerm(user, "manage_users") && (
          <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={closeMenu}>
            <MenuItem onClick={() => { closeMenu(); setOpenEditDialog(true); }}>Edit</MenuItem>
            <MenuItem onClick={() => { closeMenu(); setOpenDeleteDialog(true); }}>Delete</MenuItem>
            <MenuItem onClick={handleForceChange}>Force change password</MenuItem>
          </Menu>
        )}
      </LayoutCard>

      <CreateUserDialog open={openCreateDialog} onClose={() => setOpenCreateDialog(false)} onCreated={load} />
      <EditUserDialog open={openEditDialog} onClose={() => setOpenEditDialog(false)} onSaved={load} selectedUser={selectedRow} />
      <DeleteUserDialog open={openDeleteDialog} onClose={() => setOpenDeleteDialog(false)} onDeleted={load} selectedUser={selectedRow} />
    </>
  );
}
```

---

### Task 5: Rewrite UsersPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/UsersPage.tsx`

**Step 1: Replace UsersPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import UserManagementCard from '../../components/UserManagementCard';

export default function UsersPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Users" loading={user === undefined}>
      <UserManagementCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to `/admin/users`. Confirm:
- DataGrid renders with users
- "Create user" button opens create dialog
- Row click → menu → Edit/Delete/Force change password all work
- Dialogs open one at a time (no duplicates)

---

### Task 6: Extract SessionsCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/SessionsCard.tsx`

**Step 1: Create SessionsCard component**

Moves DataGrid, column definitions, device detection, revoke logic out of SessionsPage.

```tsx
import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef } from '@mui/x-data-grid';
import { Box, IconButton, Tooltip, Button, Typography } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckIcon from '@mui/icons-material/Check';
import { get, del } from '../api/Client';
import LayoutCard from '../tools/LayoutCard';
import { useIsMobile } from '../hooks/useMobile';

type Session = { id: number; created_at: string; expires_at: string; last_accessed_at: string; ip_address: string; user_agent: string; current?: boolean };

const getShortDeviceInfo = (userAgent: string): string => {
  if (!userAgent) return 'Unknown';
  if (userAgent.includes('iPhone')) return 'iPhone';
  if (userAgent.includes('iPad')) return 'iPad';
  if (userAgent.includes('Android')) return 'Android';
  if (userAgent.includes('Windows')) return 'Windows';
  if (userAgent.includes('Mac')) return 'Mac';
  if (userAgent.includes('Linux')) return 'Linux';
  return userAgent.substring(0, 20) + '...';
};

export default function SessionsCard() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const isMobile = useIsMobile();

  const load = async () => {
    try {
      const s = await get<Session[]>('/auth/sessions');
      setSessions(s);
    } catch (e) { console.error(e); }
  };

  useEffect(() => { load(); }, []);

  const revoke = async (id: number) => {
    try {
      await del(`/auth/sessions/${id}`);
      await load();
    } catch (e) { console.error(e); }
  };

  const allColumns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 80 },
    { field: 'ip_address', headerName: 'IP', flex: 1 },
    { field: 'user_agent', headerName: 'User Agent', flex: 2 },
    { field: 'created_at', headerName: 'Created', width: 180 },
    { field: 'last_accessed_at', headerName: 'Last Accessed', width: 180 },
    { field: 'expires_at', headerName: 'Expires', width: 180 },
    {
      field: 'actions', headerName: ' ', width: 120, renderCell: (params) => (
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          {params.row.current ? <Tooltip title="Current session"><CheckIcon color="success" /></Tooltip> : null}
          <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
            <span>
              <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async () => { await revoke(params.row.id as number); }}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        </div>
      )
    }
  ];

  const mobileColumns: GridColDef[] = [
    {
      field: 'device',
      headerName: 'Device',
      flex: 1,
      valueGetter: (_value, row) => getShortDeviceInfo(row.user_agent),
    },
    { field: 'last_accessed_at', headerName: 'Last Active', width: 140 },
    {
      field: 'actions', headerName: ' ', width: 80, renderCell: (params) => (
        <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
          {params.row.current ? <Tooltip title="Current session"><CheckIcon color="success" fontSize="small" /></Tooltip> : null}
          <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
            <span>
              <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async () => { await revoke(params.row.id as number); }}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        </div>
      )
    }
  ];

  const columns = isMobile ? mobileColumns : allColumns;

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: "stretch", height: "100%", width: "100%" }}>
      <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{ width: '100%' }}>
        <Typography variant="h4">Active Sessions</Typography>
        <Box>
          <Button variant="outlined" onClick={() => load()}>Refresh</Button>
        </Box>
      </Box>
      <div style={{ height: 400, marginTop: 10 }}>
        <DataGrid rows={sessions} columns={columns} pageSizeOptions={[5, 10, 25]} initialState={{ pagination: { paginationModel: { pageSize: 5 } } }} />
      </div>
    </LayoutCard>
  );
}
```

---

### Task 7: Rewrite SessionsPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/account/SessionsPage.tsx`

**Step 1: Replace SessionsPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import SessionsCard from '../../components/SessionsCard';

export default function SessionsPage() {
  return (
    <PageContainer titleText="Active sessions">
      <SessionsCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to the sessions page. Confirm:
- DataGrid renders with sessions
- Refresh button works
- Current session shows check icon
- Revoke button works on non-current sessions

---

### Task 8: Extract RolePermissionsCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/RolePermissionsCard.tsx`

**Step 1: Create RolePermissionsCard component**

Moves both panels (roles list + permissions editor), all role/permission state, toggle logic, and snackbar out of RolesPage.

```tsx
import { useEffect, useState } from 'react';
import { List, ListItemButton, ListItemText, Box, Switch, FormControlLabel, Paper, Typography, Divider, Snackbar, Alert, CircularProgress } from '@mui/material';
import { get, post, del } from '../api/Client';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';

type Role = { id: number; name: string };
type Permission = { id: number; name: string; description: string };

export default function RolePermissionsCard() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [rolePermissions, setRolePermissions] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [toggling, setToggling] = useState<number[]>([]);
  const [snack, setSnack] = useState<{ open: boolean; message: string; severity: 'success' | 'error' }>({ open: false, message: '', severity: 'success' });
  const { user } = useAuth();

  const load = async () => {
    setLoading(true);
    try {
      const r = await get<Role[]>('/roles/');
      setRoles(r);
      const p = await get<Permission[] | null>('/roles/permissions');
      setPermissions(p ?? []);
    } catch (e) { console.error(e); }
    setLoading(false);
  };

  const loadRolePerms = async (roleId: number) => {
    try {
      const rp = await get<Permission[] | null>(`/roles/${roleId}/permissions`);
      setRolePermissions((rp ?? []).map(x => x.id));
    } catch (e) { console.error(e); setRolePermissions([]); }
  };

  useEffect(() => { load(); }, []);

  const onRoleSelect = (r: Role) => { setSelectedRole(r); loadRolePerms(r.id); };

  const togglePermission = async (permId: number) => {
    if (!selectedRole) return;
    const has = rolePermissions.includes(permId);
    setToggling(t => [...t, permId]);
    try {
      if (has) {
        await del(`/roles/${selectedRole.id}/permissions/${permId}`);
        setSnack({ open: true, message: 'Permission removed', severity: 'success' });
      } else {
        await post(`/roles/${selectedRole.id}/permissions`, { permission_id: permId });
        setSnack({ open: true, message: 'Permission added', severity: 'success' });
      }
      await loadRolePerms(selectedRole.id);
    } catch (e) {
      console.error(e);
      setSnack({ open: true, message: 'Failed to update permission', severity: 'error' });
    } finally {
      setToggling(t => t.filter(id => id !== permId));
    }
  };

  if (!user || !(hasPerm(user, "manage_roles") || hasPerm(user, "view_roles"))) return null;

  const disabled = !hasPerm(user, "manage_roles");

  return (
    <>
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
        <Box sx={{ width: { xs: '100%', md: '32%' } }}>
          <Paper elevation={2} sx={{ p: 2, minHeight: 300, height: '100%' }}>
            <Typography variant="h6" gutterBottom>Roles</Typography>
            <Divider sx={{ mb: 1 }} />
            {loading ? (
              <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}><CircularProgress /></Box>
            ) : (
              <List>
                {roles.map(r => (
                  <ListItemButton key={r.id} selected={selectedRole?.id === r.id} onClick={() => onRoleSelect(r)}>
                    <ListItemText primary={r.name} />
                  </ListItemButton>
                ))}
                {roles.length === 0 && <Typography variant="body2" sx={{ p: 2 }}>No roles found.</Typography>}
              </List>
            )}
          </Paper>
        </Box>

        <Box sx={{ width: { xs: '100%', md: '66%' } }}>
          <Paper elevation={2} sx={{ p: 2, minHeight: 350, height: '100%' }}>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <Typography variant="h6">Permissions</Typography>
              <Typography variant="body2" color="text.secondary">{selectedRole ? `Editing: ${selectedRole.name}` : 'Select a role to view permissions'}</Typography>
            </Box>
            <Divider sx={{ my: 1 }} />

            {!selectedRole ? (
              <Box sx={{ p: 2 }}>
                <Typography variant="body2">Select a role from the left to view and modify its permissions.</Typography>
              </Box>
            ) : (
              <Box>
                {permissions.length === 0 && <Typography variant="body2" sx={{ p: 2 }}>No permissions defined.</Typography>}
                {permissions.map(p => {
                  const busy = toggling.includes(p.id);
                  const checked = rolePermissions.includes(p.id);
                  return (
                    <Box key={p.id} sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', py: 1, borderBottom: '1px solid', borderColor: 'divider' }}>
                      <Box>
                        <Typography variant="subtitle1">{p.name}</Typography>
                        <Typography variant="body2" color="text.secondary">{p.description}</Typography>
                      </Box>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {busy ? <CircularProgress size={20} /> : (
                          <FormControlLabel disabled={disabled} control={<Switch checked={checked} onChange={() => togglePermission(p.id)} />} label="" />
                        )}
                      </Box>
                    </Box>
                  );
                })}
              </Box>
            )}
          </Paper>
        </Box>
      </Box>
      <Snackbar open={snack.open} autoHideDuration={3000} onClose={() => setSnack(s => ({ ...s, open: false }))}>
        <Alert severity={snack.severity} onClose={() => setSnack(s => ({ ...s, open: false }))} sx={{ width: '100%' }}>{snack.message}</Alert>
      </Snackbar>
    </>
  );
}
```

---

### Task 9: Rewrite RolesPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/RolesPage.tsx`

**Step 1: Replace RolesPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import RolePermissionsCard from '../../components/RolePermissionsCard';

export default function RolesPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Roles" loading={user === undefined}>
      <RolePermissionsCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to `/admin/roles`. Confirm:
- Roles list renders on the left
- Clicking a role loads its permissions on the right
- Permission toggles work (switch on/off, snackbar shows)
- Loading indicators appear during toggle operations

---

### Task 10: Extract OAuthConfigCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/OAuthConfigCard.tsx`

**Step 1: Create OAuthConfigCard component**

Moves all OAuth state, status display, authorization flow, code dialog out of OAuthPage. This is the largest extraction.

```tsx
import { useEffect, useState, useCallback } from 'react';
import { Button, Box, Typography, Alert, CircularProgress, Chip, Dialog, DialogTitle, DialogContent, DialogActions, TextField } from '@mui/material';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import RefreshIcon from '@mui/icons-material/Refresh';
import SyncIcon from '@mui/icons-material/Sync';
import { getOAuthStatus, getOAuthAuthorizeURL, submitOAuthCode, reloadOAuthConfig, type OAuthStatus } from '../api/OAuth';
import LayoutCard from '../tools/LayoutCard';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { useIsMobile } from '../hooks/useMobile';

export default function OAuthConfigCard() {
  const [status, setStatus] = useState<OAuthStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [authorizing, setAuthorizing] = useState(false);
  const [reloading, setReloading] = useState(false);
  const [codeDialogOpen, setCodeDialogOpen] = useState(false);
  const [authCode, setAuthCode] = useState('');
  const [pendingState, setPendingState] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const { user } = useAuth();
  const isMobile = useIsMobile();

  const loadStatus = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const s = await getOAuthStatus();
      setStatus(s);
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to load OAuth status');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadStatus(); }, [loadStatus]);

  const handleStartAuthorize = async () => {
    try {
      setAuthorizing(true);
      setError(null);
      const { auth_url, state } = await getOAuthAuthorizeURL();
      setPendingState(state);
      window.open(auth_url, '_blank', 'width=600,height=700');
      setCodeDialogOpen(true);
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to start authorization');
    } finally {
      setAuthorizing(false);
    }
  };

  const handleSubmitCode = async () => {
    if (!authCode.trim() || !pendingState) {
      setError('Please enter the authorization code');
      return;
    }
    try {
      setSubmitting(true);
      setError(null);
      await submitOAuthCode(authCode.trim(), pendingState);
      setSuccess('OAuth authorization successful! Token has been saved.');
      setCodeDialogOpen(false);
      setAuthCode('');
      setPendingState(null);
      await loadStatus();
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to exchange authorization code');
    } finally {
      setSubmitting(false);
    }
  };

  const handleReload = async () => {
    try {
      setReloading(true);
      setError(null);
      setSuccess(null);
      await reloadOAuthConfig();
      setSuccess('OAuth configuration reloaded from disk.');
      await loadStatus();
    } catch (err: unknown) {
      const e = err as { message?: string };
      setError(e.message || 'Failed to reload OAuth configuration');
    } finally {
      setReloading(false);
    }
  };

  const handleCloseCodeDialog = () => {
    setCodeDialogOpen(false);
    setAuthCode('');
    setPendingState(null);
  };

  const canManage = !!user && hasPerm(user, 'manage_oauth');

  return (
    <>
      <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
        <Box
          display="flex"
          flexDirection={isMobile ? 'column' : 'row'}
          alignItems={isMobile ? 'flex-start' : 'center'}
          justifyContent="space-between"
          gap={2}
          mb={2}
        >
          <Typography variant="h4">OAuth Configuration</Typography>
          <Box display="flex" flexDirection={isMobile ? 'column' : 'row'} gap={1} width={isMobile ? '100%' : 'auto'}>
            <Button variant="outlined" startIcon={<SyncIcon />} onClick={handleReload} disabled={loading || reloading || !canManage} title="Reload credentials.json from disk" fullWidth={isMobile}>
              {reloading ? 'Reloading...' : 'Reload Config'}
            </Button>
            <Button variant="outlined" startIcon={<RefreshIcon />} onClick={loadStatus} disabled={loading} fullWidth={isMobile}>
              Refresh
            </Button>
          </Box>
        </Box>

        {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>{error}</Alert>}
        {success && <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>{success}</Alert>}

        {loading && !status ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}><CircularProgress /></Box>
        ) : status ? (
          <Box>
            <Typography variant="h6" gutterBottom>Status</Typography>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 3 }}>
              <Chip icon={status.configured ? <CheckCircleIcon /> : <ErrorIcon />} label={status.configured ? 'Credentials Configured' : 'Not Configured'} color={status.configured ? 'success' : 'error'} variant="outlined" />
              {status.needs_auth && <Chip icon={<ErrorIcon />} label="Needs Authorization" color="warning" variant="outlined" />}
              {!status.needs_auth && <Chip icon={status.token_valid ? <CheckCircleIcon /> : <ErrorIcon />} label={status.token_valid ? 'Token Valid' : 'Token Invalid/Expired'} color={status.token_valid ? 'success' : 'warning'} variant="outlined" />}
              <Chip label={status.refresher_active ? 'Auto-refresh Active' : 'Auto-refresh Inactive'} color={status.refresher_active ? 'info' : 'default'} variant="outlined" />
            </Box>

            {status.token_expiry && <Typography variant="body2" color="text.secondary" gutterBottom>Token Expiry: {new Date(status.token_expiry).toLocaleString()}</Typography>}
            {status.last_refresh_at && <Typography variant="body2" color="text.secondary" gutterBottom>Last Refresh: {new Date(status.last_refresh_at).toLocaleString()}</Typography>}
            {status.last_error && <Alert severity="warning" sx={{ mt: 2, mb: 2 }}>Last Error: {status.last_error}</Alert>}

            <Box sx={{ mt: 4 }}>
              <Typography variant="h6" gutterBottom>{status.needs_auth ? 'Authorize Gmail Access' : 'Re-authorize Gmail Access'}</Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                {status.needs_auth
                  ? 'OAuth credentials are configured but no token exists. Click the button below to authorize access to Gmail for sending emails.'
                  : 'If your OAuth token has expired or you need to re-authorize, click the button below to start the Google authorization flow.'}
                {' '}This will open a new window where you can sign in with your Google account.
                After authorizing, Google will display an authorization code that you will need to copy and paste here.
              </Typography>
              <Button variant="contained" onClick={handleStartAuthorize} disabled={!canManage || authorizing || !status.configured}>
                {authorizing ? 'Opening...' : 'Authorize with Google'}
              </Button>
              {!status.configured && <Typography variant="body2" color="error" sx={{ mt: 1 }}>OAuth credentials file not found. Please configure credentials.json first.</Typography>}
            </Box>
          </Box>
        ) : null}
      </LayoutCard>

      <Dialog open={codeDialogOpen} onClose={handleCloseCodeDialog} maxWidth="sm" fullWidth>
        <DialogTitle>Enter Authorization Code</DialogTitle>
        <DialogContent>
          <Typography sx={{ mb: 2 }}>
            A new window has opened for Google authorization. After you sign in and grant access,
            Google will display an authorization code. Copy that code and paste it below.
          </Typography>
          <TextField autoFocus fullWidth label="Authorization Code" value={authCode} onChange={(e) => setAuthCode(e.target.value)} placeholder="Paste the authorization code here" disabled={submitting} sx={{ mt: 1 }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseCodeDialog} disabled={submitting}>Cancel</Button>
          <Button variant="contained" onClick={handleSubmitCode} disabled={submitting || !authCode.trim()}>
            {submitting ? 'Submitting...' : 'Submit Code'}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
```

---

### Task 11: Rewrite OAuthPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/OAuthPage.tsx`

**Step 1: Replace OAuthPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import OAuthConfigCard from '../../components/OAuthConfigCard';

export default function OAuthPage() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="OAuth Management" loading={user === undefined}>
      <OAuthConfigCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to `/admin/oauth`. Confirm:
- Status chips render correctly
- Reload Config and Refresh buttons work
- Authorize with Google button opens dialog flow

---

### Task 12: Extract NotificationsCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/NotificationsCard.tsx`

**Step 1: Create NotificationsCard component**

Moves notification list, tabs, severity helpers, menu, bulk actions out of NotificationsPage.

```tsx
import { useState } from 'react';
import { Box, Typography, Button, Chip, IconButton, Menu, MenuItem, CircularProgress, Card, CardContent, Divider, Tabs, Tab } from '@mui/material';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import InfoIcon from '@mui/icons-material/Info';
import WarningIcon from '@mui/icons-material/Warning';
import ErrorIcon from '@mui/icons-material/Error';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import LayoutCard from '../tools/LayoutCard';
import { useNotifications } from '../providers/NotificationContext';
import type { NotificationSeverity, NotificationCategory } from '../api/Notifications';
import { useIsMobile } from '../hooks/useMobile';

function getSeverityIcon(severity: NotificationSeverity) {
  switch (severity) {
    case 'info': return <InfoIcon color="info" />;
    case 'warning': return <WarningIcon color="warning" />;
    case 'error': return <ErrorIcon color="error" />;
    default: return <InfoIcon />;
  }
}

function getSeverityColor(severity: NotificationSeverity): 'info' | 'warning' | 'error' | 'default' {
  switch (severity) {
    case 'info': return 'info';
    case 'warning': return 'warning';
    case 'error': return 'error';
    default: return 'default';
  }
}

function getCategoryLabel(category: NotificationCategory): string {
  switch (category) {
    case 'threshold_alert': return 'Alert';
    case 'user_management': return 'User';
    case 'config_change': return 'Config';
    default: return category;
  }
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString();
}

export default function NotificationsCard() {
  const { notifications, loading, markAsRead, dismiss, markAllAsRead, dismissAll, refresh } = useNotifications();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedNotifId, setSelectedNotifId] = useState<number | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const isMobile = useIsMobile();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, notifId: number) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedNotifId(notifId);
  };

  const handleMenuClose = () => { setAnchorEl(null); setSelectedNotifId(null); };

  const handleMarkAsRead = async () => {
    if (selectedNotifId) await markAsRead(selectedNotifId);
    handleMenuClose();
  };

  const handleDismiss = async () => {
    if (selectedNotifId) await dismiss(selectedNotifId);
    handleMenuClose();
  };

  const filteredNotifications = tabValue === 0
    ? notifications.filter(n => !n.is_read)
    : notifications;

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
      <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} flexWrap="wrap">
        <Typography variant="h4">Notifications</Typography>
        <Box display="flex" gap={1} flexWrap="wrap">
          <Button variant="outlined" onClick={() => refresh()} size={isMobile ? 'small' : 'medium'}>Refresh</Button>
          <Button variant="outlined" onClick={markAllAsRead} size={isMobile ? 'small' : 'medium'}>Mark All Read</Button>
          <Button variant="outlined" color="warning" onClick={dismissAll} size={isMobile ? 'small' : 'medium'}>Dismiss All</Button>
        </Box>
      </Box>

      <Tabs value={tabValue} onChange={(_, v) => setTabValue(v)} sx={{ mb: 2 }}>
        <Tab label={`Unread (${notifications.filter(n => !n.is_read).length})`} />
        <Tab label={`All (${notifications.length})`} />
      </Tabs>

      {loading ? (
        <Box display="flex" justifyContent="center" p={4}><CircularProgress /></Box>
      ) : filteredNotifications.length === 0 ? (
        <Box textAlign="center" py={6}>
          <CheckCircleIcon color="disabled" sx={{ fontSize: 64, mb: 2 }} />
          <Typography color="text.secondary">{tabValue === 0 ? 'No unread notifications' : 'No notifications'}</Typography>
        </Box>
      ) : (
        <Box>
          {filteredNotifications.map((notif, index) => (
            <div key={notif.notification_id}>
              <Card sx={{ mb: 1, backgroundColor: notif.is_read ? 'transparent' : 'action.hover' }} variant="outlined">
                <CardContent sx={{ display: 'flex', alignItems: 'flex-start', gap: 2, py: 2 }}>
                  <Box sx={{ mt: 0.5 }}>{getSeverityIcon(notif.notification.severity)}</Box>
                  <Box sx={{ flexGrow: 1 }}>
                    <Box display="flex" alignItems="center" gap={1} mb={0.5}>
                      <Typography variant="subtitle1" fontWeight={notif.is_read ? 'normal' : 'bold'}>{notif.notification.title}</Typography>
                      <Chip label={getCategoryLabel(notif.notification.category)} size="small" color={getSeverityColor(notif.notification.severity)} variant="outlined" />
                      {!notif.is_read && <Chip label="New" size="small" color="primary" />}
                    </Box>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>{notif.notification.message}</Typography>
                    <Typography variant="caption" color="text.disabled">{formatDate(notif.notification.created_at)}</Typography>
                  </Box>
                  <IconButton size="small" onClick={(e) => handleMenuOpen(e, notif.notification_id)}><MoreVertIcon /></IconButton>
                </CardContent>
              </Card>
              {index < filteredNotifications.length - 1 && <Divider sx={{ my: 1 }} />}
            </div>
          ))}
        </Box>
      )}

      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        {[
          <MenuItem key="mark-read" onClick={handleMarkAsRead}>Mark as Read</MenuItem>,
          <MenuItem key="dismiss" onClick={handleDismiss}>Dismiss</MenuItem>
        ]}
      </Menu>
    </LayoutCard>
  );
}
```

---

### Task 13: Rewrite NotificationsPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/notifications/NotificationsPage.tsx`

**Step 1: Replace NotificationsPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import NotificationsCard from '../../components/NotificationsCard';

export default function NotificationsPage() {
  return (
    <PageContainer titleText="Notifications">
      <NotificationsCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to notifications page. Confirm:
- Tabs switch between Unread/All
- Notification cards render with severity icons
- Context menu → Mark as Read / Dismiss works
- Bulk actions (Refresh, Mark All Read, Dismiss All) work

---

### Task 14: Extract NotificationPreferencesCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/NotificationPreferencesCard.tsx`

**Step 1: Create NotificationPreferencesCard component**

Moves preferences table, category config, toggle logic out of NotificationPreferencesPage.

```tsx
import { useState, useEffect } from 'react';
import { Box, Typography, Switch, Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Alert } from '@mui/material';
import LayoutCard from '../tools/LayoutCard';
import { useNotifications } from '../providers/NotificationContext';
import type { NotificationCategory, ChannelPreference } from '../api/Notifications';

interface CategoryConfig {
  category: NotificationCategory;
  label: string;
  description: string;
}

const CATEGORIES: CategoryConfig[] = [
  { category: 'threshold_alert', label: 'Threshold Alerts', description: 'Notifications when sensor readings exceed configured thresholds' },
  { category: 'user_management', label: 'User Management', description: 'Notifications about user creation, deletion, and role changes' },
  { category: 'config_change', label: 'Configuration Changes', description: 'Notifications when sensors are added, updated, or removed' },
];

export default function NotificationPreferencesCard() {
  const { preferences, updatePreference } = useNotifications();
  const [localPrefs, setLocalPrefs] = useState<Record<NotificationCategory, ChannelPreference>>({} as Record<NotificationCategory, ChannelPreference>);
  const [saving, setSaving] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const prefMap: Record<NotificationCategory, ChannelPreference> = {} as Record<NotificationCategory, ChannelPreference>;
    CATEGORIES.forEach(({ category }) => {
      const existing = preferences.find(p => p.category === category);
      prefMap[category] = existing || { category, email_enabled: true, inapp_enabled: true };
    });
    setLocalPrefs(prefMap);
  }, [preferences]);

  const handleToggle = async (category: NotificationCategory, channel: 'email' | 'inapp', value: boolean) => {
    const currentPref = localPrefs[category];
    const newPref: ChannelPreference = {
      ...currentPref,
      [channel === 'email' ? 'email_enabled' : 'inapp_enabled']: value,
    };
    setLocalPrefs(prev => ({ ...prev, [category]: newPref }));
    setSaving(`${category}-${channel}`);
    setError(null);
    try {
      await updatePreference(newPref);
    } catch (err) {
      setError(`Failed to save preference: ${String(err)}`);
      setLocalPrefs(prev => ({ ...prev, [category]: currentPref }));
    } finally {
      setSaving(null);
    }
  };

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: 'stretch', height: '100%', width: '100%' }}>
      <Typography variant="h4" mb={1}>Notification Preferences</Typography>
      <Typography variant="body2" color="text.secondary" mb={3}>
        Configure which notifications you receive via email and in-app notifications.
      </Typography>

      {error && <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>{error}</Alert>}

      <TableContainer component={Paper} variant="outlined">
        <Table>
          <TableHead>
            <TableRow>
              <TableCell><strong>Category</strong></TableCell>
              <TableCell align="center"><strong>Email</strong></TableCell>
              <TableCell align="center"><strong>In-App</strong></TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {CATEGORIES.map(({ category, label, description }) => {
              const pref = localPrefs[category];
              if (!pref) return null;
              return (
                <TableRow key={category}>
                  <TableCell>
                    <Typography variant="subtitle2">{label}</Typography>
                    <Typography variant="caption" color="text.secondary">{description}</Typography>
                  </TableCell>
                  <TableCell align="center">
                    <Switch checked={pref.email_enabled} onChange={(e) => handleToggle(category, 'email', e.target.checked)} disabled={saving === `${category}-email`} />
                  </TableCell>
                  <TableCell align="center">
                    <Switch checked={pref.inapp_enabled} onChange={(e) => handleToggle(category, 'inapp', e.target.checked)} disabled={saving === `${category}-inapp`} />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </LayoutCard>
  );
}
```

---

### Task 15: Rewrite NotificationPreferencesPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/notifications/NotificationPreferencesPage.tsx`

**Step 1: Replace NotificationPreferencesPage with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useNotifications } from '../../providers/NotificationContext';
import NotificationPreferencesCard from '../../components/NotificationPreferencesCard';

export default function NotificationPreferencesPage() {
  const { loading } = useNotifications();

  return (
    <PageContainer titleText="Notification Preferences" loading={loading}>
      <NotificationPreferencesCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to notification preferences page. Confirm:
- Table renders with 3 categories × 2 channels
- Toggles switch on/off and persist

---

### Task 16: Extract PropertiesCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/PropertiesCard.tsx`

**Step 1: Create PropertiesCard component**

Moves properties form, dirty tracking, save logic, snackbar out of PropertiesOverview.

```tsx
import { useState, useEffect } from 'react';
import { Box, List, ListItem, Paper, Typography, Button, TextField, Stack, CircularProgress, Snackbar, Alert } from '@mui/material';
import { PropertiesApi } from '../api/Properties';
import type { PropertiesApiStructure } from '../types/types';
import { useIsMobile } from '../hooks/useMobile';
import { useProperties } from '../hooks/useProperties';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';

export default function PropertiesCard() {
  const properties = useProperties();
  const [editedProperties, setEditedProperties] = useState<PropertiesApiStructure | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const isMobile = useIsMobile();
  const { user } = useAuth();

  useEffect(() => {
    if (properties) setEditedProperties(properties);
  }, [properties]);

  const handleChange = (key: string, value: string) => {
    setEditedProperties(prev => {
      if (!prev) return prev;
      return { ...prev, [key]: value };
    });
  };

  const isDirty = () => {
    if (!properties && !editedProperties) return false;
    if (!properties || !editedProperties) return true;
    return JSON.stringify(properties) !== JSON.stringify(editedProperties);
  };

  const handleSave = async () => {
    if (!editedProperties) return;
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      await PropertiesApi.patchProperties(editedProperties);
      setSaved(true);
    } catch (e: unknown) {
      let msg: string;
      if (e && typeof e === 'object' && 'message' in e && typeof (e as { message?: unknown }).message === 'string') {
        msg = (e as { message: string }).message;
      } else {
        try { msg = JSON.stringify(e); } catch { msg = String(e); }
      }
      setError(msg);
    } finally {
      setSaving(false);
      setTimeout(() => setSaved(false), 2000);
    }
  };

  const fieldsDisabled = !user || !hasPerm(user, "manage_properties");

  return (
    <>
      {editedProperties && (
        <Paper sx={{ padding: 2, width: '100%', maxWidth: '100%', alignSelf: 'stretch' }}>
          <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
            <Typography variant="h4">Properties</Typography>
            <Stack direction="row" spacing={1} alignItems="center">
              {saving && <CircularProgress size={20} />}
              <Button variant="contained" color="primary" onClick={handleSave} disabled={!isDirty() || saving || fieldsDisabled}>
                Save changes
              </Button>
            </Stack>
          </Stack>

          {error && <Typography color="error">Error: {error}</Typography>}
          {!properties && !error && <Typography>Loading properties...</Typography>}

          <List sx={{ width: '100%' }}>
            {Object.entries(editedProperties).map(([key, value]) => (
              <ListItem key={key} divider sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: isMobile ? 'stretch' : 'center', width: '100%', py: 1, gap: 2 }}>
                <Box sx={{ flex: isMobile ? '0 0 auto' : '0 0 420px', minWidth: 160, pr: isMobile ? 0 : 2, wordBreak: 'break-all', width: isMobile ? '100%' : undefined }}>
                  <Typography sx={{ fontFamily: 'monospace' }}>{key}</Typography>
                </Box>
                <Box sx={{ flex: '1 1 auto', width: isMobile ? '100%' : undefined }}>
                  <TextField fullWidth value={value} onChange={(e) => handleChange(key, e.target.value)} size="small" disabled={fieldsDisabled} />
                </Box>
              </ListItem>
            ))}
          </List>
        </Paper>
      )}

      <Snackbar open={saved} autoHideDuration={2000} onClose={() => setSaved(false)} anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert severity="success" sx={{ width: '100%' }}>Properties updated successfully</Alert>
      </Snackbar>
    </>
  );
}
```

---

### Task 17: Rewrite PropertiesOverview as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/properties/PropertiesOverview.tsx`

**Step 1: Replace PropertiesOverview with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import { useAuth } from '../../providers/AuthContext';
import PropertiesCard from '../../components/PropertiesCard';

export default function PropertiesOverview() {
  const { user } = useAuth();

  return (
    <PageContainer titleText="Properties Overview" loading={user === undefined}>
      <PropertiesCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to properties page. Confirm:
- Properties list renders with editable text fields
- "Save changes" button is disabled until a value is changed
- Saving shows spinner and success snackbar

---

### Task 18: Extract ChangePasswordCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/ChangePasswordCard.tsx`

**Step 1: Create ChangePasswordCard component**

Moves password form, validation, submit logic out of ChangePassword page.

```tsx
import React, { useState } from 'react';
import { changePassword } from '../api/Users';
import { useNavigate } from 'react-router';
import { Container, Box, TextField, Button, Typography, Alert, Avatar, Paper, CircularProgress } from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';

function extractErrorMessage(err: unknown): string | null {
  if (!err) return null;
  if (typeof err === 'string') return err;
  if (typeof err === 'object' && err !== null && 'message' in err) {
    const e = err as { message?: unknown };
    if (typeof e.message === 'string') return e.message;
    return String(e.message);
  }
  return String(err);
}

export default function ChangePasswordCard() {
  const [newPassword, setNewPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (newPassword !== confirm) {
      setError('Passwords do not match');
      return;
    }
    setLoading(true);
    try {
      await changePassword(0, newPassword);
      navigate('/');
    } catch (err: unknown) {
      const message = extractErrorMessage(err) || 'Failed to change password';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="sm">
      <Paper elevation={4} sx={{ p: 4, borderRadius: 2, mt: 4 }}>
        <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
          <Avatar sx={{ bgcolor: 'primary.main' }}><LockOutlinedIcon /></Avatar>
          <Typography component="h1" variant="h5">Change password</Typography>
          {error && <Alert severity="error" sx={{ width: '100%' }}>{error}</Alert>}
          <Box component="form" onSubmit={submit} sx={{ mt: 1, width: '100%' }}>
            <TextField label="New password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} fullWidth margin="normal" disabled={loading} required />
            <TextField label="Confirm password" type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} fullWidth margin="normal" disabled={loading} required />
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
              <Button type="submit" variant="contained" disabled={loading} startIcon={loading ? <CircularProgress color="inherit" size={18} /> : undefined}>
                {loading ? 'Saving...' : 'Change password'}
              </Button>
            </Box>
          </Box>
        </Box>
      </Paper>
    </Container>
  );
}
```

---

### Task 19: Rewrite ChangePassword as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/account/ChangePassword.tsx`

**Step 1: Replace ChangePassword with composition**

```tsx
import PageContainer from '../../tools/PageContainer';
import ChangePasswordCard from '../../components/ChangePasswordCard';

export default function ChangePasswordPage() {
  return (
    <PageContainer titleText="Change password">
      <ChangePasswordCard />
    </PageContainer>
  );
}
```

**Step 2: Verify in browser**

Navigate to change password page. Confirm:
- Form renders with password fields
- Validation error shown when passwords don't match

---

### Task 20: Update conforming pages to use `loading` prop

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/temperature-dashboard/TemperatureDashboard.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensors-overview/SensorsOverview.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensor/SensorPage.tsx`

**Step 1: Update TemperatureDashboard**

Remove the `if (user === undefined)` loading block (lines 29-46). Replace first `<PageContainer>` with:
```tsx
<PageContainer titleText="Temperature Dashboard" loading={user === undefined}>
```

**Step 2: Update SensorsOverview**

Remove the `if (user === undefined)` loading block. Replace first `<PageContainer>` with:
```tsx
<PageContainer titleText="Sensors Overview" loading={user === undefined}>
```

**Step 3: Update SensorPage**

Remove the `if (user === undefined)` loading block. Replace first `<PageContainer>` with:
```tsx
<PageContainer titleText="Sensor Details" loading={user === undefined}>
```

Note: SensorPage has a second conditional check (`if (!sensor)`) — that stays in the page since it's a different condition (sensor not found, not auth loading).

---

### Task 21: Browser verification

**Step 1: Navigate all pages**

Test each page loads correctly:
- `/` (Temperature Dashboard)
- `/sensors` (Sensors Overview)
- `/sensor/:name` (Sensor Page — click any sensor)
- `/alerts` (Alert Rules)
- `/admin/users` (Users)
- `/admin/roles` (Roles)
- `/admin/oauth` (OAuth)
- `/account/sessions` (Sessions — find exact route)
- `/notifications` (Notifications — find exact route)
- `/notifications/preferences` (Notification Preferences — find exact route)
- `/properties` (Properties — find exact route)
- `/account/change-password` (Change Password)

**Step 2: Spot-check interactive features**

- Alerts: Create/Edit/Delete dialogs open individually
- Users: Create/Edit/Delete dialogs open individually
- Roles: Select role → toggle permissions
- OAuth: Refresh/Reload buttons work
- Sessions: Revoke button works

**Step 3: Check console for errors**

No new JavaScript errors should appear in the browser console.
