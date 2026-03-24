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