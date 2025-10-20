import { useState, useCallback } from 'react';
import type { Sensor } from '../types/types.ts';
import { SensorsApi } from '../api/Sensors';
import type {ApiError} from "../api/Client.ts";
import type { FormikHelpers } from 'formik';

export interface UseSensorFormOpts {
  mode?: 'create' | 'edit';
  initialSensor?: Sensor | null;
  onSuccess?: (sensor: Sensor | null) => void;
}

type SensorFormValues = {
  name: string;
  type: string;
  url: string;
};

export function useSensorForm({ mode = 'edit', initialSensor = null, onSuccess }: UseSensorFormOpts) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [advancedErrorMessage, setAdvancedErrorMessage] = useState<string | null>(null);

  const initialValues = {
    name: initialSensor?.name ?? '',
    type: initialSensor?.type ?? '',
    url: initialSensor?.url ?? '',
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
    async (values: { name: string; type: string; url: string }, actions: FormikHelpers<SensorFormValues>) => {
      setIsSubmitting(true);
      setSuccessMessage(null);
      setErrorMessage(null);
      setAdvancedErrorMessage(null);

      const payload = { name: values.name, type: values.type, url: values.url };

      try {
        if (mode === 'create') {
          await SensorsApi.add(payload);
        } else {
          if (!initialSensor || initialSensor.id == null) {
            throw new Error('Missing sensor id for update');
          }
          await SensorsApi.update(Number(initialSensor.id), payload); // returns ApiMessage
        }

        let newSensor: Sensor | null;
        try {
          newSensor = await SensorsApi.getByName(values.name);
        } catch (fetchErr) {
          console.log('Failed to fetch sensor after create/update:', fetchErr);
          newSensor = null;
        }

        setSuccessMessage(mode === 'create' ? 'Sensor created successfully!' : 'Sensor updated successfully!');
        if (onSuccess) onSuccess(newSensor);

        if (actions && typeof actions.resetForm === 'function') {
          if (mode === 'create') {
            actions.resetForm({ values: { name: '', type: '', url: '' } });
          } else {
            actions.resetForm({
              values: {
                name: newSensor?.name ?? values.name,
                type: newSensor?.type ?? values.type,
                url: newSensor?.url ?? values.url,
              },
            });
          }
        }
      } catch (err) {
        handleErrors(err);
      } finally {
        setIsSubmitting(false);
        if (actions && typeof actions.setSubmitting === 'function') actions.setSubmitting(false);
        if (!errorMessage && successMessage) {
          setTimeout(() => setSuccessMessage(null), 3000);
        }
      }
    },
    [mode, onSuccess, initialSensor, errorMessage, successMessage]
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