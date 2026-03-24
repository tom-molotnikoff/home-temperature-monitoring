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
