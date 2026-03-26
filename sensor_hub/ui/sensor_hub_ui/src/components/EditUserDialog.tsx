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
import { logger } from '../tools/logger';

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
    }).catch(e => logger.error('Failed to load roles', e));
  }, [open, selectedUser]);

  const handleSave = async () => {
    if (!selectedUser) return;
    try {
      await setUserRoles(selectedUser.id, [role]);
      onClose();
      await onSaved();
    } catch (e) {
      logger.error('Failed to update user roles', e);
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
