import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Box, Snackbar, Switch } from '@mui/material';
import { alpha, useTheme } from '@mui/material/styles';
import type { WidgetProps } from '../types';
import type { Capability, CommandStatusMessage } from '../../gen/aliases';
import { apiClient } from '../../gen/client';
import { useCurrentReadings } from '../../hooks/useCurrentReadings';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';
import NeedsConfiguration from '../NeedsConfiguration';
import { useReportWidgetUpdate } from '../WidgetUpdateContext';

function resolveBinaryCapability(
  capabilities: Capability[] | undefined,
  property: string,
): Capability | undefined {
  return capabilities?.find((capability) => capability.type === 'binary' && capability.property === property);
}

export default function SensorToggleWidget({ config }: WidgetProps) {
  const theme = useTheme();
  const { sensors } = useSensorContext();
  const { user } = useAuth();
  const reportUpdate = useReportWidgetUpdate();

  const sensorId = config.sensorId as number | undefined;
  const property = config.property as string | undefined;
  const sensor = sensorId ? sensors.find((candidate) => candidate.id === sensorId) : undefined;
  const capability = sensor && property ? resolveBinaryCapability(sensor.capabilities, property) : undefined;
  const valueOn = capability?.value_on ?? 'ON';
  const valueOff = capability?.value_off ?? 'OFF';
  const [optimisticValue, setOptimisticValue] = useState<string | null>(null);
  const pendingCommandRef = useRef<{ id: number; previousValue: string | null } | null>(null);
  const [snackbarMessage, setSnackbarMessage] = useState<string | null>(null);

  const handleCommandStatus = useCallback((message: CommandStatusMessage) => {
    const pendingCommand = pendingCommandRef.current;
    if (!pendingCommand || !sensor || !property) return;
    if (message.id !== pendingCommand.id || message.sensor_id !== sensor.id || message.property !== property) return;

    if (message.status === 'failed' || message.status === 'timed_out') {
      setOptimisticValue(pendingCommand.previousValue);
      setSnackbarMessage(message.status === 'timed_out' ? 'Command timed out' : 'Command failed');
      reportUpdate(new Date());
    }

    pendingCommandRef.current = null;
  }, [property, reportUpdate, sensor]);

  const readings = useCurrentReadings({ onDataUpdate: reportUpdate, onCommandStatus: handleCommandStatus });
  const reading = sensor && property ? readings[sensor.name]?.[property] : undefined;

  const effectiveValue = optimisticValue ?? reading?.text_state ?? null;
  const checked = effectiveValue === valueOn;
  const canControl = hasPerm(user, 'control_sensors');

  const visualState = useMemo(() => ({
    trackBackground: checked
      ? alpha(theme.palette.primary.main, canControl ? 0.95 : 0.55)
      : alpha(theme.palette.text.secondary, canControl ? 0.35 : 0.2),
    thumbBackground: theme.palette.common.white,
    onOpacity: checked ? 1 : 0.35,
    offOpacity: checked ? 0.35 : 1,
  }), [canControl, checked, theme.palette.common.white, theme.palette.primary.main, theme.palette.text.secondary]);

  useEffect(() => {
    if (optimisticValue != null && reading?.text_state === optimisticValue) {
      setOptimisticValue(null);
    }
  }, [optimisticValue, reading?.text_state]);

  if (!sensor || !property || !capability) {
    return <NeedsConfiguration message="Select a controllable sensor and binary property" />;
  }

  const handleToggle = async () => {
    if (!canControl) return;

    const previousValue = reading?.text_state ?? null;
    const nextValue = checked ? valueOff : valueOn;
    setOptimisticValue(nextValue);
    reportUpdate(new Date());

    const { data, error } = await apiClient.POST('/sensors/{id}/command', {
      params: { path: { id: sensor.id } },
      body: { property, value: nextValue },
    });

    if (error) {
      setOptimisticValue(previousValue);
      setSnackbarMessage('Failed to send command');
      return;
    }

    if (data) {
      pendingCommandRef.current = { id: data.id, previousValue };
    }
  };

  return (
    <>
      <Box sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2 }}>
        <Switch
          checked={checked}
          disabled={!canControl}
          disableRipple={!canControl}
          onChange={() => {
            void handleToggle();
          }}
          slotProps={{ input: { 'aria-label': `Toggle ${sensor.name} ${property}` } }}
          sx={{
            width: 110,
            height: 64,
            p: 0,
            '& .MuiSwitch-switchBase': {
              p: '4px',
              transitionDuration: '220ms',
              '&.Mui-checked': {
                transform: 'translateX(46px)',
                color: theme.palette.common.white,
                '& + .MuiSwitch-track': {
                  backgroundColor: visualState.trackBackground,
                  opacity: 1,
                },
                '&.Mui-disabled + .MuiSwitch-track': {
                  opacity: 1,
                },
              },
              '&.Mui-disabled': {
                color: theme.palette.common.white,
              },
            },
            '& .MuiSwitch-thumb': {
              boxSizing: 'border-box',
              width: 56,
              height: 56,
              backgroundColor: visualState.thumbBackground,
              boxShadow: checked
                ? `0 0 12px ${alpha(theme.palette.primary.main, canControl ? 0.35 : 0.2)}`
                : undefined,
            },
            '& .MuiSwitch-track': {
              borderRadius: 32,
              opacity: 1,
              backgroundColor: visualState.trackBackground,
              position: 'relative',
              '&::before, &::after': {
                position: 'absolute',
                top: '50%',
                transform: 'translateY(-50%)',
                fontSize: '0.82rem',
                fontWeight: 700,
                letterSpacing: '0.08em',
                color: checked ? theme.palette.common.white : theme.palette.text.primary,
              },
              '&::before': {
                content: '"ON"',
                left: 14,
                opacity: visualState.onOpacity,
              },
              '&::after': {
                content: '"OFF"',
                right: 12,
                opacity: visualState.offOpacity,
              },
            },
          }}
        />
      </Box>
      <Snackbar
        open={snackbarMessage != null}
        autoHideDuration={3000}
        onClose={() => setSnackbarMessage(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert severity="error" onClose={() => setSnackbarMessage(null)} sx={{ width: '100%' }}>
          {snackbarMessage}
        </Alert>
      </Snackbar>
    </>
  );
}
