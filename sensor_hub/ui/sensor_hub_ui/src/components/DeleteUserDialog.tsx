import {Button, Dialog, DialogActions, DialogContent, DialogTitle} from "@mui/material";
import type {User} from "../api/Users.ts";
import {deleteUser} from "../api/Users.ts";
import { logger } from '../tools/logger';

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
      logger.error('Failed to delete user', e);
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
