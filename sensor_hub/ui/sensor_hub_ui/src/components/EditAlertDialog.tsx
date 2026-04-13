import {useEffect, useState} from "react";
import {type AlertRule, type UpdateAlertRuleRequest, updateAlertRule} from "../api/Alerts.ts";
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
import { logger } from '../tools/logger';

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
      const request: UpdateAlertRuleRequest = {
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

      await updateAlertRule(selectedAlert.ID, request);
      onClose();
      await onSaved();
    } catch (e) {
      logger.error('Failed to update alert rule', e);
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