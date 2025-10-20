import type {Sensor} from "../types/types.ts";
import LayoutCard from "../tools/LayoutCard.tsx";
import { CardContent, Chip, Typography, Box, Link, Avatar, Button} from '@mui/material';
import SensorsIcon from '@mui/icons-material/Sensors';
import { useState } from 'react';
import { Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Alert, CircularProgress } from '@mui/material';
import { useNavigate } from 'react-router';
import {SensorsApi} from "../api/Sensors.ts";
import type {ApiError} from "../api/Client.ts";

interface SensorInfoCardProps {
  sensor: Sensor
  onDelete?: (name: string) => void
  onDisable?: (name: string) => void
  onEnable?: (name: string) => void
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

function SensorInfoCard({sensor, onDelete, onDisable, onEnable}: SensorInfoCardProps) {
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [disableDialogOpen, setDisableDialogOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const navigate = useNavigate();

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
    if (loading) return; // prevent closing while request in-flight
    setDisableDialogOpen(false);
  }

  const closeDeleteDialog = () => {
    if (loading) return; // prevent closing while request in-flight
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

  const handleEnableSensor = async () => { await performSensorAction('enable'); };
  const handleConfirmDisable = async () => { await performSensorAction('disable'); };
  const handleConfirmDelete = async () => { await performSensorAction('delete'); };

  return (
    <LayoutCard variant="secondary" changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <Box display="flex" alignItems="center" gap={2} mb={2}>
        <Typography variant="h4" height={40}>
          {sensor.name}
        </Typography>
        <Avatar sx={{ bgcolor: getHealthBgColor(sensor.healthStatus), width: 40, height: 40 }}>
          <SensorsIcon />
        </Avatar>
      </Box>
      <CardContent>
        <Box display="flex" flexDirection="column" gap={2}>
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Type:</Typography>
            <Chip label={sensor.type} color="primary" size="small" />
          </Box>
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Health:</Typography>
            <Chip label={sensor.healthStatus} color={getHealthColor(sensor.healthStatus)} size="small" />
          </Box>
          {sensor.healthReason && (
            <Box display="flex" alignItems="center" gap={1}>
              <Typography variant="subtitle1" sx={{
                textWrap: "nowrap"
              }}>Health Reason:</Typography>
              <Typography variant="body2" color="text.secondary">{sensor.healthReason}</Typography>
            </Box>
          )}
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Enabled:</Typography>
            <Chip label={sensor.enabled ? 'true' : 'false'} color={sensor.enabled ? 'success' : 'error'} size="small" />
          </Box>
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">API URL:</Typography>
            <Link href={sensor.url} target="_blank" rel="noopener">{sensor.url}</Link>
          </Box>
          <Box display="flex" alignItems="center" gap={1}>
            <Button variant="contained" color="error" onClick={openDeleteDialog}>
              Delete
            </Button>
            <Button
              variant="outlined"
              color="warning"
              onClick={openDisableDialog}
              disabled={!sensor.enabled || loading}
            >
              Disable
            </Button>
            <Button
              variant="contained"
              color="success"
              onClick={handleEnableSensor}
              disabled={sensor.enabled || loading}
              startIcon={loading ? <CircularProgress size={18} /> : null}
            >
              {loading ? 'Enabling...' : 'Enable'}
            </Button>
          </Box>
        </Box>
      </CardContent>

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
