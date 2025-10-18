import { useState } from "react";
import type {Sensor} from "../types/types.ts";
import {Formik, Field, Form, type FieldInputProps, type FormikHelpers} from 'formik';
import { Button, CardContent, Box, Stack, TextField, Typography, Alert } from '@mui/material';
import {API_BASE} from "../environment/Environment.ts";

interface SensorFormProps {
  sensor?: Sensor;
  mode?: 'create' | 'edit';
  onSuccess?: (sensor: Sensor) => void;
}

interface SensorFormValues {
  name: string;
  type: string;
  url: string;
}

interface CreatePayload {
  name: string;
  type: string;
  url: string;
}

interface EditPayload extends CreatePayload {
  id?: number | string;
}

class ApiError extends Error {
  causeData?: unknown;
  constructor(message: string, causeData?: unknown) {
    super(message);
    this.causeData = causeData;
  }
}

function toMessageAndCauseFromResponseBody(body: unknown): { message: string | null; cause?: unknown } {
  if (body === null || body === undefined) return { message: null };
  if (typeof body === 'string') return { message: body };
  if (typeof body === 'object') {
    const asObj = body as Record<string, unknown>;
    const message = (typeof asObj.message === 'string' ? asObj.message : null) ?? (typeof asObj.error === 'string' ? asObj.error : null);
    const cause = asObj.error ?? asObj.cause ?? null;
    return { message, cause: cause ?? undefined };
  }
  return { message: String(body) };
}

function SensorForm ({ sensor, mode = 'edit', onSuccess } : SensorFormProps) {
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [advancedErrorMessage, setAdvancedErrorMessage] = useState<string | null>(null);

  const initialValues: SensorFormValues = {
    name: sensor?.name ?? "",
    type: sensor?.type ?? "",
    url: sensor?.url ?? "",
  }

  const onFormSubmit = async (values: SensorFormValues, actions: FormikHelpers<SensorFormValues>) => {
    setSuccessMessage(null);
    setErrorMessage(null);
    setAdvancedErrorMessage(null);

    const payload: CreatePayload | EditPayload = mode === 'create'
      ? { name: values.name, type: values.type, url: values.url }
      : { id: sensor?.id, name: values.name, type: values.type, url: values.url };

    if (mode === 'edit' && (sensor == null || sensor.id == null)) {
      setErrorMessage('Cannot update sensor: missing sensor id');
      actions.setSubmitting(false);
      return;
    }

    const method = mode === 'create' ? 'POST' : 'PUT';
    const url = mode === 'create' ? `${API_BASE}/sensors/` : `${API_BASE}/sensors/${sensor?.id}`;

    try {
      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        let parsed: unknown = null;
        try {
          parsed = await response.json();
        } catch (parseErr) {
          console.log('Failed to parse error response as JSON:', parseErr);
        }

        const { message, cause } = toMessageAndCauseFromResponseBody(parsed ?? response.statusText);
        const errMsg = message ?? (response.statusText || `HTTP ${response.status}`);
        throw new ApiError(errMsg, cause);
      }

      // Success - server returns the created/updated sensor
      const data = await response.json();

      // Reset form values depending on mode
      if (mode === 'create') {
        actions.resetForm({ values: { name: '', type: '', url: '' } });
      } else {
        actions.resetForm({ values: {
          name: (data && typeof data.name === 'string') ? data.name : '',
          type: (data && typeof data.type === 'string') ? data.type : '',
          url: (data && typeof data.url === 'string') ? data.url : '',
        }});
      }

      setSuccessMessage(mode === 'create' ? 'Sensor created successfully!' : 'Sensor updated successfully!');
      setTimeout(() => setSuccessMessage(null), 3000);

      if (onSuccess) {
        onSuccess(data as Sensor);
      }
    } catch (error: unknown) {
      if (error instanceof ApiError) {
        setErrorMessage(error.message || 'API error');
        if (error.causeData !== undefined && error.causeData !== null) {
          try {
            setAdvancedErrorMessage(typeof error.causeData === 'string' ? error.causeData : JSON.stringify(error.causeData));
          } catch (jsonErr: unknown) {
            if (jsonErr instanceof Error) {
              setAdvancedErrorMessage(`Failed to serialize error cause: ${jsonErr.message}`);
            } else {
              setAdvancedErrorMessage('Failed to serialize error cause: unknown error');
            }
          }
        }
      } else if (error instanceof Error) {
        setErrorMessage(error.message || 'Unexpected error');
      } else {
        try {
          setErrorMessage(JSON.stringify(error));
        } catch (jsonErr: unknown) {
          if (jsonErr instanceof Error) {
            setErrorMessage(`Unexpected error (failed to serialize): ${jsonErr.message}`);
          } else {
            setErrorMessage('Unexpected error (failed to serialize): unknown error');
          }
        }
      }
    } finally {
      actions.setSubmitting(false);
    }
  }

  return (
      <CardContent sx={{width: "100%", padding: 3, maxWidth: 650}}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          {mode === 'create' ? 'Add Sensor' : 'Edit Sensor Details'}
        </Typography>
        <Formik
          initialValues={initialValues}
          enableReinitialize
          onSubmit={onFormSubmit}
        >
          {({ isSubmitting, dirty, handleChange }) => (
            <Form>
              <Stack spacing={2}>
                <Field name="name">
                  {({ field }: { field: FieldInputProps<string> }) => (
                    <TextField {...field} label="Name" variant="outlined" fullWidth size="small" onChange={e => { handleChange(e); setErrorMessage(null); setSuccessMessage(null); }} />
                  )}
                </Field>
                <Field name="type">
                  {({ field }: { field: FieldInputProps<string> }) => (
                    <TextField {...field} label="Type" variant="outlined" fullWidth size="small" onChange={e => { handleChange(e); setErrorMessage(null); setSuccessMessage(null); }} />
                  )}
                </Field>
                <Field name="url">
                  {({ field }: { field: FieldInputProps<string> }) => (
                    <TextField {...field} label="API URL" variant="outlined" fullWidth size="small" onChange={e => { handleChange(e); setErrorMessage(null); setSuccessMessage(null); }} />
                  )}
                </Field>
                <Box display="flex" >
                  <Button type="reset" disabled={isSubmitting} variant="outlined" color="primary" sx={{ mr: 2 }}>
                    Reset
                  </Button>
                  <Button type="submit" disabled={isSubmitting || !dirty} variant="contained" color="primary">
                    {mode === 'create' ? 'Create' : 'Submit'}
                  </Button>
                </Box>
              </Stack>
            </Form>
          )}
        </Formik>
        {successMessage && (
          <Box mt={2}>
            <Alert severity="success" onClose={() => setSuccessMessage(null)}>{successMessage}</Alert>
          </Box>
        )}
        {errorMessage && (
          <Box mt={2}>
            <Alert severity="error" onClose={() => {
              setErrorMessage(null)
              setAdvancedErrorMessage(null);
            }}>{errorMessage}
              {advancedErrorMessage && (
                <Box mt={1} sx={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace', fontSize: '0.75rem' }} >
                  {advancedErrorMessage}
                </Box>
              )}
            </Alert>
          </Box>
        )}
      </CardContent>
  );
}

export default SensorForm;

