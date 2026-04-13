import { useState, useEffect } from 'react';
import {
  Box,
  CardContent,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  Alert,
  Typography,
  CircularProgress,
  MenuItem,
} from '@mui/material';
import type { Sensor } from '../types/types';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import { SensorsApi } from '../api/Sensors';
import { useAuth } from '../providers/AuthContext';
import { hasPerm } from '../tools/Utils';
import { useProperties } from '../hooks/useProperties';

interface SensorRetentionCardProps {
  sensor: Sensor;
}

type RetentionUnit = 'hours' | 'days' | 'weeks';

const unitMultipliers: Record<RetentionUnit, number> = {
  hours: 1,
  days: 24,
  weeks: 168,
};

function hoursToUnit(hours: number, unit: RetentionUnit): number {
  return Math.round((hours / unitMultipliers[unit]) * 100) / 100;
}

function unitToHours(value: number, unit: RetentionUnit): number {
  return Math.round(value * unitMultipliers[unit]);
}

function formatRetention(hours: number): string {
  if (hours >= 168 && hours % 168 === 0) return `${hours / 168} week${hours / 168 !== 1 ? 's' : ''}`;
  if (hours >= 24 && hours % 24 === 0) return `${hours / 24} day${hours / 24 !== 1 ? 's' : ''}`;
  return `${hours} hour${hours !== 1 ? 's' : ''}`;
}

function SensorRetentionCard({ sensor }: SensorRetentionCardProps) {
  const { user } = useAuth();
  const properties = useProperties();
  const globalRetentionDays = parseInt(properties['sensor.data.retention.days'] || '90', 10);
  const globalRetentionHours = globalRetentionDays * 24;
  const effectiveHours = sensor.retentionHours ?? globalRetentionHours;

  const [useCustom, setUseCustom] = useState(sensor.retentionHours !== null);
  const [unit, setUnit] = useState<RetentionUnit>('days');
  const [value, setValue] = useState('');
  const [saving, setSaving] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    const hasCustom = sensor.retentionHours !== null;
    setUseCustom(hasCustom);
    if (hasCustom && sensor.retentionHours !== null) {
      const h = sensor.retentionHours;
      if (h >= 168 && h % 168 === 0) {
        setUnit('weeks');
        setValue(String(h / 168));
      } else if (h >= 24 && h % 24 === 0) {
        setUnit('days');
        setValue(String(h / 24));
      } else {
        setUnit('hours');
        setValue(String(h));
      }
    } else {
      setUnit('days');
      setValue('');
    }
  }, [sensor.retentionHours]);

  const fieldsDisabled = !user || !hasPerm(user, 'manage_sensors');

  const handleSave = async () => {
    setSuccessMessage(null);
    setErrorMessage(null);
    setSaving(true);
    try {
      const retentionHours = useCustom ? unitToHours(parseFloat(value), unit) : null;
      if (useCustom && (!retentionHours || retentionHours < 1)) {
        setErrorMessage('Retention must be at least 1 hour');
        setSaving(false);
        return;
      }
      await SensorsApi.update(sensor.id, { retention_hours: retentionHours });
      setSuccessMessage(retentionHours ? `Retention set to ${formatRetention(retentionHours)}` : 'Reverted to global default');
    } catch {
      setErrorMessage('Failed to update retention');
    } finally {
      setSaving(false);
    }
  };

  const hasChanges = (() => {
    if (useCustom !== (sensor.retentionHours !== null)) return true;
    if (useCustom && value !== '') {
      const newHours = unitToHours(parseFloat(value), unit);
      return newHours !== sensor.retentionHours;
    }
    return false;
  })();

  return (
    <LayoutCard variant="secondary" changes={{ alignItems: 'center', height: '100%', width: '100%' }}>
      <CardContent sx={{ width: '100%', padding: 3, maxWidth: 650 }}>
        <TypographyH2>Data Retention</TypographyH2>
        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Global default: {formatRetention(globalRetentionHours)}. Effective: {formatRetention(effectiveHours)}.
        </Typography>

        {successMessage && (
          <Alert severity="success" onClose={() => setSuccessMessage(null)} sx={{ mb: 2 }}>
            {successMessage}
          </Alert>
        )}
        {errorMessage && (
          <Alert severity="error" onClose={() => setErrorMessage(null)} sx={{ mb: 2 }}>
            {errorMessage}
          </Alert>
        )}

        <FormControlLabel
          control={
            <Switch
              checked={useCustom}
              onChange={(e) => {
                setUseCustom(e.target.checked);
                if (!e.target.checked) setValue('');
              }}
              disabled={fieldsDisabled}
            />
          }
          label="Override global retention for this sensor"
          sx={{ mb: 2 }}
        />

        {useCustom && (
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'flex-start', mb: 2 }}>
            <TextField
              label="Retention"
              type="number"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              disabled={fieldsDisabled}
              slotProps={{ htmlInput: { min: 1, step: 1 } }}
              size="small"
              sx={{ width: 140 }}
            />
            <TextField
              select
              label="Unit"
              value={unit}
              onChange={(e) => {
                const newUnit = e.target.value as RetentionUnit;
                if (value) {
                  const hours = unitToHours(parseFloat(value), unit);
                  setValue(String(hoursToUnit(hours, newUnit)));
                }
                setUnit(newUnit);
              }}
              disabled={fieldsDisabled}
              size="small"
              sx={{ width: 120 }}
            >
              <MenuItem value="hours">Hours</MenuItem>
              <MenuItem value="days">Days</MenuItem>
              <MenuItem value="weeks">Weeks</MenuItem>
            </TextField>
          </Box>
        )}

        <Box sx={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Button
            variant="contained"
            onClick={handleSave}
            disabled={fieldsDisabled || saving || !hasChanges}
            startIcon={saving ? <CircularProgress color="inherit" size={18} /> : undefined}
          >
            {saving ? 'Saving...' : 'Save'}
          </Button>
        </Box>
      </CardContent>
    </LayoutCard>
  );
}

export default SensorRetentionCard;
