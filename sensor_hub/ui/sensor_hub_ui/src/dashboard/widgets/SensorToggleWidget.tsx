import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Box, Snackbar } from '@mui/material';
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

const CONTROL_MAX_WIDTH = 220;
const CONTROL_HEIGHT = 72;
const CONTROL_PADDING = 4;
const THUMB_WIDTH = 104;
const THUMB_HEIGHT = 64;
const DRAG_TRAVEL = CONTROL_MAX_WIDTH - (CONTROL_PADDING * 2) - THUMB_WIDTH;

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
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
  const [dragProgress, setDragProgress] = useState<number | null>(null);
  const dragOriginXRef = useRef(0);
  const dragStartProgressRef = useRef(0);
  const dragMovedRef = useRef(false);
  const suppressClickRef = useRef(false);
  const controlRef = useRef<HTMLDivElement | null>(null);

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
  const progress = dragProgress ?? (checked ? 1 : 0);
  const crossedMidpoint = progress >= 0.5;
  const thumbLeft = CONTROL_PADDING + (progress * DRAG_TRAVEL);
  const isDragging = dragProgress !== null;

  const visualState = useMemo(() => ({
    trackBackground: crossedMidpoint
      ? alpha(theme.palette.primary.main, canControl ? 0.95 : 0.55)
      : alpha(theme.palette.text.secondary, canControl ? 0.35 : 0.2),
    thumbBackground: theme.palette.common.white,
    onOpacity: 0.35 + (progress * 0.65),
    offOpacity: 0.35 + ((1 - progress) * 0.65),
  }), [canControl, crossedMidpoint, progress, theme.palette.common.white, theme.palette.primary.main, theme.palette.text.secondary]);

  useEffect(() => {
    if (optimisticValue != null && reading?.text_state === optimisticValue) {
      setOptimisticValue(null);
    }
  }, [optimisticValue, reading?.text_state]);

  if (!sensor || !property || !capability) {
    return <NeedsConfiguration message="Select a controllable sensor and binary property" />;
  }

  const commitCheckedState = async (nextChecked: boolean) => {
    if (!canControl) return;

    const previousValue = reading?.text_state ?? null;
    const nextValue = nextChecked ? valueOn : valueOff;
    if (nextValue === effectiveValue) return;

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

  const handlePointerDown = (event: React.PointerEvent<HTMLDivElement>) => {
    if (!canControl) return;

    dragOriginXRef.current = event.clientX;
    dragStartProgressRef.current = checked ? 1 : 0;
    dragMovedRef.current = false;
    setDragProgress(dragStartProgressRef.current);

    if (typeof event.currentTarget.setPointerCapture === 'function') {
      event.currentTarget.setPointerCapture(event.pointerId);
    }
  };

  const handlePointerMove = (event: React.PointerEvent<HTMLDivElement>) => {
    if (dragProgress === null) return;

    const delta = event.clientX - dragOriginXRef.current;
    if (Math.abs(delta) > 4) {
      dragMovedRef.current = true;
    }
    setDragProgress(clamp(dragStartProgressRef.current + (delta / DRAG_TRAVEL), 0, 1));
  };

  const finishDrag = (event: React.PointerEvent<HTMLDivElement>) => {
    if (dragProgress === null) return;

    const finalProgress = dragProgress;
    const dragged = dragMovedRef.current;
    setDragProgress(null);

    if (typeof event.currentTarget.releasePointerCapture === 'function') {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }

    if (!dragged) return;

    suppressClickRef.current = true;
    void commitCheckedState(finalProgress >= 0.5);
  };

  const handleClick = () => {
    if (!canControl) return;
    if (suppressClickRef.current) {
      suppressClickRef.current = false;
      return;
    }
    void commitCheckedState(!checked);
  };

  return (
    <>
      <Box sx={{ height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', p: 2 }}>
        <Box
          ref={controlRef}
          data-testid="sensor-toggle-control"
          role="checkbox"
          aria-checked={checked}
          aria-disabled={!canControl}
          aria-label={`Toggle ${sensor.name} ${property}`}
          tabIndex={canControl ? 0 : -1}
          onClick={handleClick}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={finishDrag}
          onPointerCancel={finishDrag}
          onKeyDown={(event) => {
            if (!canControl) return;
            if (event.key === ' ' || event.key === 'Enter') {
              event.preventDefault();
              void commitCheckedState(!checked);
            }
          }}
          sx={{
            position: 'relative',
            width: '100%',
            maxWidth: `${CONTROL_MAX_WIDTH}px`,
            height: `${CONTROL_HEIGHT}px`,
            borderRadius: `${CONTROL_HEIGHT / 2}px`,
            backgroundColor: visualState.trackBackground,
            boxShadow: crossedMidpoint
              ? `inset 0 0 0 1px ${alpha(theme.palette.common.white, 0.14)}, 0 10px 24px ${alpha(theme.palette.primary.main, canControl ? 0.2 : 0.08)}`
              : `inset 0 0 0 1px ${alpha(theme.palette.text.primary, 0.08)}`,
            cursor: canControl ? (isDragging ? 'grabbing' : 'pointer') : 'default',
            userSelect: 'none',
            touchAction: 'none',
            outline: 'none',
            transition: isDragging
              ? 'background-color 90ms linear, box-shadow 90ms linear'
              : 'background-color 180ms ease, box-shadow 180ms ease',
            '&::before': {
              content: '""',
              position: 'absolute',
              left: '50%',
              top: 16,
              bottom: 16,
              width: 2,
              transform: 'translateX(-50%)',
              borderRadius: 999,
              backgroundColor: crossedMidpoint
                ? alpha(theme.palette.common.white, 0.28)
                : alpha(theme.palette.text.primary, 0.12),
            },
          }}
        >
          <Box
            sx={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              px: 2.5,
              fontSize: '0.9rem',
              fontWeight: 800,
              letterSpacing: '0.1em',
              color: crossedMidpoint ? theme.palette.common.white : theme.palette.text.primary,
            }}
          >
            <Box component="span" sx={{ opacity: visualState.onOpacity }}>ON</Box>
            <Box component="span" sx={{ opacity: visualState.offOpacity }}>OFF</Box>
          </Box>
          <Box
            sx={{
              position: 'absolute',
              top: CONTROL_PADDING,
              left: 0,
              width: `${THUMB_WIDTH}px`,
              height: `${THUMB_HEIGHT}px`,
              borderRadius: `${THUMB_HEIGHT / 2}px`,
              transform: `translateX(${thumbLeft}px) scale(${isDragging ? 0.985 : 1})`,
              backgroundColor: visualState.thumbBackground,
              boxShadow: isDragging
                ? `0 6px 14px ${alpha(theme.palette.common.black, 0.18)}`
                : `0 10px 24px ${alpha(theme.palette.common.black, crossedMidpoint ? 0.18 : 0.12)}`,
              transition: isDragging
                ? 'none'
                : 'transform 240ms cubic-bezier(0.2, 0.9, 0.25, 1.25), box-shadow 200ms ease',
              '&::before': {
                content: '""',
                position: 'absolute',
                inset: 16,
                borderRadius: 999,
                background: crossedMidpoint
                  ? `linear-gradient(90deg, ${alpha(theme.palette.primary.main, 0.28)}, ${alpha(theme.palette.primary.light, 0.12)})`
                  : `linear-gradient(90deg, ${alpha(theme.palette.text.secondary, 0.16)}, ${alpha(theme.palette.text.secondary, 0.06)})`,
              },
            }}
          />
        </Box>
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
