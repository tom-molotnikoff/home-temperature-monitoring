import { useState, useCallback } from 'react';
import type { Sensor } from '../types/types.ts';
import { SensorsApi } from '../api/Sensors';
import type {ApiError} from "../api/Client.ts";
import type { FormikHelpers } from 'formik';
import { logger } from '../tools/logger';
import type { SensorFormValues } from '../forms/SensorForm';
import { bestUnit, hoursToUnit, unitToHours } from '../tools/retention';

export interface UseSensorFormOpts {
  mode?: 'create' | 'edit';
  initialSensor?: Sensor | null;
  onSuccess?: (sensor: Sensor | null) => void;
}

function retentionInitialValues(sensor: Sensor | null): Pick<SensorFormValues, 'retentionEnabled' | 'retentionValue' | 'retentionUnit'> {
  if (sensor?.retentionHours != null) {
    const u = bestUnit(sensor.retentionHours);
    return { retentionEnabled: true, retentionValue: String(hoursToUnit(sensor.retentionHours, u)), retentionUnit: u };
  }
  return { retentionEnabled: false, retentionValue: '', retentionUnit: 'days' };
}

export function useSensorForm({ mode = 'edit', initialSensor = null, onSuccess }: UseSensorFormOpts) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [advancedErrorMessage, setAdvancedErrorMessage] = useState<string | null>(null);

  const initialValues: SensorFormValues = {
    name: initialSensor?.name ?? '',
    sensorDriver: initialSensor?.sensorDriver ?? '',
    config: initialSensor?.config ?? {},
    ...retentionInitialValues(initialSensor),
  };

  const handleErrors = (err: unknown) => {
    setErrorMessage(null);
    setAdvancedErrorMessage(null);
    if (err && typeof err === 'object' && 'message' in err) {

      setErrorMessage(String((err as ApiError).message));
      if ((err as ApiError).error) setAdvancedErrorMessage(JSON.stringify((err as ApiError).error));
      return;
    }
    if (err instanceof Error) {
      setErrorMessage(err.message);
      return;
    }
    try {
      setErrorMessage(JSON.stringify(err));
    } catch {
      setErrorMessage('Unknown error');
    }
  };

  const onSubmit = useCallback(
    async (values: SensorFormValues, actions: FormikHelpers<SensorFormValues>) => {
      setIsSubmitting(true);
      setSuccessMessage(null);
      setErrorMessage(null);
      setAdvancedErrorMessage(null);

      try {
        if (mode === 'create') {
          await SensorsApi.add({
            name: values.name,
            sensor_driver: values.sensorDriver,
            config: values.config,
          });
        } else {
          if (!initialSensor || initialSensor.id == null) {
            throw new Error('Missing sensor id for update');
          }
          const retentionHours = values.retentionEnabled
            ? unitToHours(parseFloat(values.retentionValue), values.retentionUnit)
            : null;
          await SensorsApi.update(Number(initialSensor.id), {
            name: values.name,
            sensor_driver: values.sensorDriver,
            config: values.config,
            retention_hours: retentionHours,
          });
        }

        let newSensor: Sensor | null;
        try {
          newSensor = await SensorsApi.getByName(values.name);
        } catch (fetchErr) {
          logger.debug('Failed to fetch sensor after create/update:', fetchErr);
          newSensor = null;
        }

        setSuccessMessage(mode === 'create' ? 'Sensor created successfully!' : 'Sensor updated successfully!');
        setTimeout(() => setSuccessMessage(null), 3000);
        if (onSuccess) onSuccess(newSensor);

        if (actions && typeof actions.resetForm === 'function') {
          if (mode === 'create') {
            actions.resetForm({ values: { name: '', sensorDriver: '', config: {}, retentionEnabled: false, retentionValue: '', retentionUnit: 'days' } });
          } else {
            actions.resetForm({
              values: {
                name: newSensor?.name ?? values.name,
                sensorDriver: newSensor?.sensorDriver ?? values.sensorDriver,
                config: newSensor?.config ?? values.config,
                ...retentionInitialValues(newSensor),
              },
            });
          }
        }
      } catch (err) {
        handleErrors(err);
      } finally {
        setIsSubmitting(false);
        if (actions && typeof actions.setSubmitting === 'function') actions.setSubmitting(false);
      }
    },
    [mode, onSuccess, initialSensor]
  );

  return {
    initialValues,
    onSubmit,
    isSubmitting,
    successMessage,
    errorMessage,
    advancedErrorMessage,
    setSuccessMessage,
    setErrorMessage,
    setAdvancedErrorMessage,
  };
}