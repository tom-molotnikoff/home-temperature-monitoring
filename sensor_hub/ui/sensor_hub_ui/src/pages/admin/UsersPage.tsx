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
  Box, Grid,
  Menu,
  MenuItem,
  Select,
  InputLabel,
  FormControl,
  Typography
} from '@mui/material';
import { listUsers, createUser, deleteUser, setMustChange, setUserRoles } from '../../api/Users';
import { listRoles, type Role } from '../../api/Roles';
import type { User } from '../../api/Users';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from "../../tools/LayoutCard.tsx";
import {useAuth} from "../../providers/AuthContext.tsx";
import {hasPerm} from "../../tools/Utils.ts";

export default function UsersPage(){
  const [users, setUsers] = useState<User[]>([]);
  const [openCreate, setOpenCreate] = useState(false);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('user');
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);

  const [menuAnchorEl, setMenuAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedRow, setSelectedRow] = useState<User | null>(null);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<User | null>(null);
  const [openEdit, setOpenEdit] = useState(false);
  const [openDeleteConfirm, setOpenDeleteConfirm] = useState(false);
  const { user } = useAuth();

  const load = async () => {
    try{
      const u = await listUsers();
      setUsers(u);
    }catch(e){
      console.error(e);
    }
  }

  useEffect(()=>{
    (async () => {
      try {
        const u = await listUsers();
        setUsers(u);
        const r = await listRoles();
        setAvailableRoles(r || []);
      } catch(err) {
        console.error('initial load failed', err);
      }
    })();
  }, []);

  useEffect(() => {
    if (availableRoles && availableRoles.length > 0) {
      const found = availableRoles.find(x => x.name === role);
      if (!found) setRole(availableRoles[0].name);
    }
  }, [availableRoles]);

  const handleRowClick = (params: GridRowParams, event: React.MouseEvent) => {
    const id = typeof params.id === 'number' ? params.id : Number(params.id);
    const found = users.find(u => u.id === id);
    if (found) setSelectedRow(found);
    else setSelectedRow(params.row as User);
    setMenuAnchorEl(event.currentTarget as HTMLElement);
  };

  const handleMenuClose = () => { setMenuAnchorEl(null); };

  const handleOpenEdit = () => {
    console.debug('[UsersPage] handleOpenEdit selectedRow=', selectedRow);
    setEditUser(selectedRow);
    setRole((selectedRow?.roles && selectedRow.roles.length>0)? selectedRow.roles[0] : 'user');
    setOpenEdit(true);
    handleMenuClose();
  };

  const handleOpenDelete = () => { setDeleteTarget(selectedRow); setOpenDeleteConfirm(true); handleMenuClose(); };

  const handleForceChange = async () => {
    if (!selectedRow) return;
    await setMustChange(selectedRow.id, true);
    await load();
    handleMenuClose();
  }

  const confirmDelete = async () => {
    const target = deleteTarget || selectedRow;
    if (!target) return;
    try {
      console.debug('[UsersPage] confirmDelete target=', target);
      await deleteUser(target.id);
      setOpenDeleteConfirm(false);
      setDeleteTarget(null);
      await load();
    } catch (err) {
      console.error('[UsersPage] delete failed', err);
    }
  }

  const saveEdit = async () => {
    const target = editUser || selectedRow;
    if (!target) return;
    try {
      console.debug('[UsersPage] saveEdit target=', target, 'role=', role);
      await setUserRoles(target.id, [role]);
      setOpenEdit(false);
      setEditUser(null);
      await load();
    } catch (err) {
      console.error('[UsersPage] saveEdit failed', err);
    }
  }

  const onCreate = async () => {
    try{
      await createUser({ username, email, password, roles: [role] });
      setOpenCreate(false);
      setUsername(''); setEmail(''); setPassword(''); setRole('user');
      await load();
    }catch(e){
      console.error(e);
    }
  }

  const columns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 80 },
    { field: 'username', headerName: 'Username', flex: 1 },
    { field: 'email', headerName: 'Email', flex: 1 },
    { field: 'rolesDisplay', headerName: 'Roles', flex: 1 },
    { field: 'must_change_password', headerName: 'Must change password', width: 200 },
  ];

  const rows = users.map(u => ({ ...u, rolesDisplay: (u.roles || []).join(', ') }));

  if (user === undefined ) {
    return (
      <PageContainer titleText="Users">
        <Box sx={{flexGrow: 1, width: '100%'}}>
          <Grid
            container
            spacing={2}
            alignItems="stretch"
            sx={{ minHeight: "100%" }}
          >
            <Grid size={12}>
              Loading...
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

  const fieldsDisabled = !(hasPerm(user, "manage_users"));

  return (
    <PageContainer titleText="Users">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%", width: "100%" }}>
          <Box sx={{ width: '100%' }}>
            <LayoutCard variant="secondary" changes={{alignItems: "stretch", height: "100%", width: "100%"}}>
              <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{width: '100%'}}>
                <Typography variant="h4">Users</Typography>
                <Box>
                  <Button variant="contained" onClick={()=>setOpenCreate(true)} disabled={fieldsDisabled}>Create user</Button>
                </Box>
              </Box>
              <div style={{ height: 400, width: '100%' }}>
                <DataGrid rows={rows} columns={columns} pageSizeOptions={[5,10,25]} initialState={{pagination:{paginationModel:{pageSize:5}}}} onRowClick={handleRowClick} />
              </div>

              {(hasPerm(user, "manage_users")) &&
                <Menu anchorEl={menuAnchorEl} open={Boolean(menuAnchorEl)} onClose={handleMenuClose}>
                  <MenuItem onClick={handleOpenEdit}>Edit</MenuItem>
                  <MenuItem onClick={handleOpenDelete}>Delete</MenuItem>
                  <MenuItem onClick={handleForceChange}>Force change password</MenuItem>
                </Menu>
              }


            </LayoutCard>
          </Box>
        </Grid>
      </Box>

      <Dialog open={openCreate} onClose={()=>setOpenCreate(false)}>
        <DialogTitle>Create user</DialogTitle>
        <DialogContent>
          <TextField fullWidth label="Username" value={username} onChange={(e)=>setUsername(e.target.value)} sx={{mt:1}} />
          <TextField fullWidth label="Email" value={email} onChange={(e)=>setEmail(e.target.value)} sx={{mt:1}} />
          <TextField fullWidth label="Password" type="password" value={password} onChange={(e)=>setPassword(e.target.value)} sx={{mt:1}} />
          <FormControl fullWidth sx={{mt:1}}>
            <InputLabel id="role-select-label">Role</InputLabel>
            <Select labelId="role-select-label" value={role} label="Role" onChange={(e)=>setRole(e.target.value as string)}>
              {availableRoles.map(r => (<MenuItem key={r.name} value={r.name}>{r.name}</MenuItem>))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={()=>setOpenCreate(false)}>Cancel</Button>
          <Button variant="contained" onClick={onCreate}>Create</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openEdit} onClose={()=>setOpenEdit(false)}>
        <DialogTitle>Edit user</DialogTitle>
        <DialogContent>
          <TextField fullWidth label="Username" value={editUser?.username ?? ''} disabled sx={{mt:1}} />
          <FormControl fullWidth sx={{mt:2}}>
            <InputLabel id="edit-role-select-label">Role</InputLabel>
            <Select labelId="edit-role-select-label" value={role} label="Role" onChange={(e)=>setRole(e.target.value as string)}>
              {availableRoles.map(r => (<MenuItem key={r.name} value={r.name}>{r.name}</MenuItem>))}
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={()=>setOpenEdit(false)}>Cancel</Button>
          <Button variant="contained" onClick={saveEdit}>Save</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={openDeleteConfirm} onClose={()=>setOpenDeleteConfirm(false)}>
        <DialogTitle>Delete user</DialogTitle>
        <DialogContent>
          Are you sure you want to delete user <strong>{deleteTarget?.username ?? selectedRow?.username}</strong>?
        </DialogContent>
        <DialogActions>
          <Button onClick={()=>setOpenDeleteConfirm(false)}>Cancel</Button>
          <Button variant="contained" color="error" onClick={confirmDelete}>Delete</Button>
        </DialogActions>
      </Dialog>

    </PageContainer>
   )
 }
