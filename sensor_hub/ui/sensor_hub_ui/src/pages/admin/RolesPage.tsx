import { useEffect, useState } from 'react';
import PageContainer from '../../tools/PageContainer';
import { List, ListItemButton, ListItemText, Box, Switch, FormControlLabel, Paper, Typography, Divider, Snackbar, Alert, CircularProgress } from '@mui/material';
import { get, post, del } from '../../api/Client';

type Role = { id: number; name: string };
type Permission = { id: number; name: string; description: string };

export default function RolesPage(){
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [rolePermissions, setRolePermissions] = useState<number[]>([]);
  const [loading, setLoading] = useState(false);
  const [toggling, setToggling] = useState<number[]>([]); // permission ids being toggled
  const [snack, setSnack] = useState<{open:boolean; message:string; severity:'success'|'error'}>({open:false,message:'',severity:'success'});

  const load = async () => {
    setLoading(true);
    try{
      const r = await get<Role[]>('/roles/');
      setRoles(r);
      const p = await get<Permission[] | null>('/roles/permissions');
      setPermissions(p ?? []);
    }catch(e){ console.error(e); }
    setLoading(false);
  }

  const loadRolePerms = async (roleId: number) => {
    try{
      const rp = await get<Permission[] | null>(`/roles/${roleId}/permissions`);
      setRolePermissions((rp ?? []).map(x => x.id));
    }catch(e){ console.error(e); setRolePermissions([]); }
  }

  useEffect(()=>{ load() }, []);

  const onRoleSelect = (r: Role) => { setSelectedRole(r); loadRolePerms(r.id); }

  const togglePermission = async (permId: number) => {
    if (!selectedRole) return;
    const has = rolePermissions.includes(permId);
    setToggling(t => [...t, permId]);
    try{
      if (has) {
        await del(`/roles/${selectedRole.id}/permissions/${permId}`);
        setSnack({ open: true, message: 'Permission removed', severity: 'success' });
      } else {
        await post(`/roles/${selectedRole.id}/permissions`, { permission_id: permId });
        setSnack({ open: true, message: 'Permission added', severity: 'success' });
      }
      await loadRolePerms(selectedRole.id);
    }catch(e){
      console.error(e);
      setSnack({ open: true, message: 'Failed to update permission', severity: 'error' });
    }finally{
      setToggling(t => t.filter(id => id !== permId));
    }
  }

  return (
    <PageContainer titleText="Roles">
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
        <Box sx={{ width: { xs: '100%', md: '32%' } }}>
          <Paper elevation={2} sx={{ p:2, minHeight: 300, height: '100%' }}>
            <Typography variant="h6" gutterBottom>Roles</Typography>
            <Divider sx={{ mb:1 }} />
            {loading ? (
              <Box sx={{ display:'flex', justifyContent:'center', p:2 }}><CircularProgress/></Box>
            ) : (
              <List>
                {roles.map(r => (
                  <ListItemButton key={r.id} selected={selectedRole?.id === r.id} onClick={()=>onRoleSelect(r)}>
                    <ListItemText primary={r.name} />
                  </ListItemButton>
                ))}
                {roles.length === 0 && <Typography variant="body2" sx={{ p:2 }}>No roles found.</Typography>}
              </List>
            )}
          </Paper>
        </Box>

        <Box sx={{ width: { xs: '100%', md: '66%' } }}>
          <Paper elevation={2} sx={{ p:2, minHeight: 350, height: '100%' }}>
            <Box sx={{ display:'flex', alignItems:'center', justifyContent:'space-between' }}>
              <Typography variant="h6">Permissions</Typography>
              <Typography variant="body2" color="text.secondary">{selectedRole ? `Editing: ${selectedRole.name}` : 'Select a role to view permissions'}</Typography>
            </Box>
            <Divider sx={{ my:1 }} />

            {!selectedRole ? (
              <Box sx={{ p:2 }}>
                <Typography variant="body2">Select a role from the left to view and modify its permissions.</Typography>
              </Box>
            ) : (
              <Box>
                {permissions.length === 0 && <Typography variant="body2" sx={{ p:2 }}>No permissions defined.</Typography>}
                {permissions.map(p => {
                  const busy = toggling.includes(p.id);
                  const checked = rolePermissions.includes(p.id);
                  return (
                    <Box key={p.id} sx={{ display:'flex', alignItems:'center', justifyContent:'space-between', py:1, borderBottom:'1px solid', borderColor:'divider' }}>
                      <Box>
                        <Typography variant="subtitle1">{p.name}</Typography>
                        <Typography variant="body2" color="text.secondary">{p.description}</Typography>
                      </Box>
                      <Box sx={{ display:'flex', alignItems:'center', gap:1 }}>
                        {busy ? <CircularProgress size={20} /> : (
                          <FormControlLabel control={<Switch checked={checked} onChange={()=>togglePermission(p.id)} />} label="" />
                        )}
                      </Box>
                    </Box>
                  )
                })}
              </Box>
            )}
          </Paper>
        </Box>
      </Box>

      <Snackbar open={snack.open} autoHideDuration={3000} onClose={()=>setSnack(s=>({...s,open:false}))}>
        <Alert severity={snack.severity} onClose={()=>setSnack(s=>({...s,open:false}))} sx={{ width: '100%' }}>{snack.message}</Alert>
      </Snackbar>
    </PageContainer>
  )
}
