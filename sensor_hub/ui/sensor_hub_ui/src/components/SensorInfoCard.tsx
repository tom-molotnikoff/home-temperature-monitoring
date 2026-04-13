import type {Sensor} from "../types/types.ts";
import LayoutCard from "../tools/LayoutCard.tsx";
import { Chip, Typography, Box, Avatar, Button} from '@mui/material';
import SensorsIcon from '@mui/icons-material/Sensors';
import { useState, useEffect } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Alert, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router';
import {SensorsApi} from "../api/Sensors.ts";
import { MeasurementTypesApi } from '../api/Sensors';
import type { MeasurementTypeInfo } from '../types/types';
import type {ApiError} from "../api/Client.ts";
import type {AuthUser} from "../providers/AuthContext.tsx";
import {hasPerm} from "../tools/Utils.ts";
import {TypographyH2} from "../tools/Typography.tsx";
import {useProperties} from "../hooks/useProperties.ts";
import {formatRetention} from "../tools/retention.ts";

interface SensorInfoCardProps {
  sensor: Sensor
  onDelete?: (name: string) => void
  onDisable?: (name: string) => void
  onEnable?: (name: string) => void
  user: AuthUser;
}

function getHealthColor(status: Sensor['healthStatus']) {
  switch (status) {
    case 'good': return 'success';
    case 'bad': return 'error';
    case 'unknown': return 'warning';
    default: return 'default';
  }
}

function getHealthBgColor(status: Sensor['healthStatus']) {
  switch (status) {
    case 'good': return 'success.main';
    case 'bad': return 'error.main';
    case 'unknown': return 'warning.main';
    default: return 'grey.400';
  }
}

function SensorInfoCard({sensor, onDelete, onDisable, onEnable, user}: SensorInfoCardProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [disableDialogOpen, setDisableDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [measurementTypes, setMeasurementTypes] = useState<MeasurementTypeInfo[]>([]);
  const navigate = useNavigate();
  const properties = useProperties();

  const globalRetentionDays = parseInt(properties['sensor.data.retention.days'] || '90', 10);
  const globalRetentionHours = globalRetentionDays * 24;
  const effectiveHours = sensor.retentionHours ?? globalRetentionHours;

  useEffect(() => {
    MeasurementTypesApi.getForSensor(sensor.id)
      .then((types) => setMeasurementTypes(types))
      .catch(() => setMeasurementTypes([]));
  }, [sensor.id]);

  const openDeleteDialog = () => {
    setErrorMessage(null);
    setSuccessMessage(null);
    setDeleteDialogOpen(true);
  };

  const openDisableDialog = () => {
    setErrorMessage(null);
    setSuccessMessage(null);
    setDisableDialogOpen(true);
  }

  const closeDisableDialog = () => {
    if (loading) return;
    setDisableDialogOpen(false);
  }

  const closeDeleteDialog = () => {
    if (loading) return;
    setDeleteDialogOpen(false);
  };

  const performSensorAction = async (action: 'enable' | 'disable' | 'delete') => {
    setLoading(true);
    setErrorMessage(null);

    try {
      switch (action) {
        case 'enable':
          await SensorsApi.enableByName(sensor.name);
          setSuccessMessage('Sensor enabled');
          if (onEnable) onEnable(sensor.name);
          break;

        case 'disable':
          await SensorsApi.disableByName(sensor.name);
          setSuccessMessage('Sensor disabled');
          setDisableDialogOpen(false);
          if (onDisable) onDisable(sensor.name);
          break;

        case 'delete':
          await SensorsApi.delete(sensor.name);
          setSuccessMessage('Sensor deleted');
          setDeleteDialogOpen(false);
          if (onDelete) onDelete(sensor.name);
          navigate('/sensors-overview');
          break;
      }

      setTimeout(() => setSuccessMessage(null), 1500);
    } catch (err: unknown) {
      const msg = (err && typeof err === 'object' && 'message' in err) ? (err as ApiError).message : String(err);
      setErrorMessage(msg);
      return;
    } finally {
      setLoading(false);
    }
  };
  const fieldsDisabled = !(hasPerm(user, "manage_sensors"));
  const handleEnableSensor = async () => { await performSensorAction('enable'); };
  const handleConfirmDisable = async () => { await performSensorAction('disable'); };
  const handleConfirmDelete = async () => { await performSensorAction('delete'); };

  return (
    <LayoutCard variant="secondary" changes={{height: "100%", width: "100%"}}>
      <Box display="flex" alignItems="center" gap={2} mb={2}>
        <TypographyH2>
          {sensor.name}
        </TypographyH2>
        <Avatar sx={{ bgcolor: getHealthBgColor(sensor.healthStatus), width: 40, height: 40 }}>
          <SensorsIcon />
        </Avatar>
      </Box>
      <Box display="flex" flexDirection="column" gap={2}>
        <Box display="flex" alignItems="center" gap={1}>
          <Typography variant="subtitle1">Driver:</Typography>
          <Chip label={sensor.sensorDriver} color="primary" size="small" />
        </Box>
        {sensor.external_id && (
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Device ID:</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'monospace' }}>{sensor.external_id}</Typography>
          </Box>
        )}
        <Box display="flex" alignItems="center" gap={1}>
          <Typography variant="subtitle1">Health:</Typography>
          <Chip label={sensor.healthStatus} color={getHealthColor(sensor.healthStatus)} size="small" />
        </Box>
        {sensor.healthReason && (
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1" sx={{ textWrap: "nowrap" }}>Health Reason:</Typography>
            <Typography variant="body2" color="text.secondary">{sensor.healthReason}</Typography>
          </Box>
        )}
        <Box display="flex" alignItems="center" gap={1}>
          <Typography variant="subtitle1">Enabled:</Typography>
          <Chip label={sensor.enabled ? 'true' : 'false'} color={sensor.enabled ? 'success' : 'error'} size="small" />
        </Box>
        <Box display="flex" alignItems="center" gap={1}>
          <Typography variant="subtitle1">Retention:</Typography>
          {sensor.retentionHours !== null
            ? <Chip label={`Custom: ${formatRetention(effectiveHours)}`} color="primary" size="small" variant="outlined" />
            : <Typography variant="body2" color="text.secondary">Global default ({formatRetention(globalRetentionHours)})</Typography>
          }
        </Box>
        {measurementTypes.length > 0 && (
          <Box>
            <Typography variant="subtitle1" sx={{ mb: 0.5 }}>Measurement Types:</Typography>
            <Box display="flex" flexWrap="wrap" gap={1}>
              {measurementTypes.map((mt) => (
                <Chip key={mt.id} label={`${mt.display_name} (${mt.unit})`} size="small" variant="outlined" />
              ))}
            </Box>
          </Box>
        )}
        {sensor.config && Object.keys(sensor.config).length > 0 && (
          <Box>
            <Typography variant="subtitle1" sx={{ mb: 0.5 }}>Configuration:</Typography>
            {Object.entries(sensor.config).map(([key, value]) => (
              <Box key={key} display="flex" alignItems="center" gap={1} sx={{ ml: 2 }}>
                <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 'bold' }}>{key}:</Typography>
                <Typography variant="body2" color="text.secondary">{value}</Typography>
              </Box>
            ))}
          </Box>
        )}
        <Box display="flex" alignItems="center" gap={1}>
          <Button
            variant="contained"
            color="error"
            onClick={openDeleteDialog}
            disabled={loading || fieldsDisabled}
          >
            Delete
          </Button>
          <Button
            variant="outlined"
            color="warning"
            onClick={openDisableDialog}
            disabled={!sensor.enabled || loading || fieldsDisabled}
          >
            Disable
          </Button>
          <Button
            variant="contained"
            color="success"
            onClick={handleEnableSensor}
            disabled={sensor.enabled || loading || fieldsDisabled}
          >
            Enable
          </Button>
        </Box>
      </Box>

      <Dialog open={disableDialogOpen} onClose={closeDisableDialog} >
        <DialogTitle>Disable sensor "{sensor.name}"?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            This action will disable the sensor, preventing it from collecting new data. Existing data will be retained. You can re-enable the sensor later if needed.
          </DialogContentText>

          {errorMessage && <Box mt={2}><Alert severity="error">{errorMessage}</Alert></Box>}
          {successMessage && <Box mt={2}><Alert severity="success">{successMessage}</Alert></Box>}
        </DialogContent>
        <DialogActions>
          <Button onClick={closeDisableDialog} disabled={loading}>Cancel</Button>
          <Button onClick={handleConfirmDisable} color="warning" variant="contained" disabled={loading} startIcon={loading ? <CircularProgress size={18} /> : null}>
            {loading ? 'Disabling...' : 'Confirm Disable'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteDialogOpen} onClose={closeDeleteDialog}>
        <DialogTitle>Delete sensor "{sensor.name}"?</DialogTitle>
        <DialogContent>
          <DialogContentText>
            This action will permanently delete the sensor from the system. This will also purge any associated sensor readings, if you want to keep the readings, consider disabling the sensor instead. Purging may take some time depending on the volume of data.
          </DialogContentText>

          {errorMessage && <Box mt={2}><Alert severity="error">{errorMessage}</Alert></Box>}
          {successMessage && <Box mt={2}><Alert severity="success">{successMessage}</Alert></Box>}
        </DialogContent>
        <DialogActions>
          <Button onClick={closeDeleteDialog} disabled={loading}>Cancel</Button>
          <Button onClick={handleConfirmDelete} color="error" variant="contained" disabled={loading} startIcon={loading ? <CircularProgress size={18} /> : null}>
            {loading ? 'Deleting...' : 'Confirm Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </LayoutCard>
  );
}

export default SensorInfoCard;
