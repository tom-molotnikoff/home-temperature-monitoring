import { useEffect, useState } from 'react';
import { DataGrid } from '@mui/x-data-grid';
import type { GridColDef } from '@mui/x-data-grid';
import {Box, Grid, IconButton, Tooltip, Button, Typography} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import CheckIcon from '@mui/icons-material/Check';
import { get, del } from '../../api/Client';
import PageContainer from '../../tools/PageContainer';
import LayoutCard from '../../tools/LayoutCard';
import { useIsMobile } from '../../hooks/useMobile';

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

export default function SessionsPage(){
  const [sessions, setSessions] = useState<Session[]>([]);
  const isMobile = useIsMobile();

  const load = async () => {
    try{
      const s = await get<Session[]>('/auth/sessions');
      setSessions(s);
    }catch(e){ console.error(e); }
  }

  useEffect(()=>{ load() }, []);

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
          { params.row.current ? <Tooltip title="Current session"><CheckIcon color="success"/></Tooltip> : null }
          <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
            <span>
              <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async ()=>{ await revoke(params.row.id as number); }}>
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
          { params.row.current ? <Tooltip title="Current session"><CheckIcon color="success" fontSize="small"/></Tooltip> : null }
          <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
            <span>
              <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async ()=>{ await revoke(params.row.id as number); }}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </span>
          </Tooltip>
        </div>
      )
    }
  ];

  const columns = isMobile ? mobileColumns : allColumns;

  const revoke = async (id: number) => {
    try{
      await del(`/auth/sessions/${id}`);
      await load();
    }catch(e){ console.error(e); }
  }

  return (
    <PageContainer titleText="Active sessions">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: "100%", width: "100%" }}>
          <Box sx={{ width: '100%' }}>
            <LayoutCard variant="secondary" changes={{alignItems: "stretch", height: "100%", width: "100%"}}>
              <Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2} sx={{width: '100%'}}>
                <Typography variant="h4">Active Sessions</Typography>
                <Box>
                  <Button variant="outlined" onClick={()=>load()}>Refresh</Button>
                </Box>
              </Box>
              <div style={{ height: 400, marginTop: 10 }}>
                <DataGrid rows={sessions} columns={columns} pageSizeOptions={[5,10,25]} initialState={{pagination:{paginationModel:{pageSize:5}}}}/>
              </div>
            </LayoutCard>
          </Box>
         </Grid>
      </Box>
    </PageContainer>
  )
}
