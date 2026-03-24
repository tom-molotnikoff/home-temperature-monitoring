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
