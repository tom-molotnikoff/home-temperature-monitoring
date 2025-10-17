import { useState } from "react";
import type {Sensor} from "../types/types.ts";
import {Formik, Field, Form, type FieldInputProps, type FormikHelpers} from 'formik';
import { Button, CardContent, Box, Stack, TextField, Typography, Alert } from '@mui/material';

interface SensorFormProps {
  sensor: Sensor
}

interface SensorFormValues {
  name: string;
  type: string;
  url: string;
}

function SensorForm ({ sensor } : SensorFormProps) {
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [advancedErrorMessage, setAdvancedErrorMessage] = useState<string | null>(null);
  const initialValues: SensorFormValues = {
    name: sensor.name ?? "",
    type: sensor.type ?? "",
    url: sensor.url ?? "",
  }

  const onFormSubmit = (values: SensorFormValues, actions: FormikHelpers<SensorFormValues>) => {
      setSuccessMessage(null);
      setErrorMessage(null);
      setAdvancedErrorMessage(null);
      const newSensorDetails = {
        ...sensor,
        name: values.name,
        type: values.type,
        url: values.url,
      };

      fetch(`http://localhost:8080/sensors/${sensor.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newSensorDetails),
      })
        .then(async response => {
          if (!response.ok) {
            let errorText;
            let advancedErrorMessageText = null;
            try {
              const errorData = await response.json();
              errorText = errorData.message || JSON.stringify(errorData);
              advancedErrorMessageText = errorData.error;
            } catch {
              errorText = response.statusText;
            }
            throw new Error(errorText, { cause: advancedErrorMessageText } );
          }
          return response.json();
        })
        .then(data => {
          actions.resetForm({ values: {
              name: data.name ?? "",
              type: data.type ?? "",
              url: data.url ?? "",
            }});
          setSuccessMessage('Sensor updated successfully!');
          setTimeout(() => setSuccessMessage(null), 3000);
        })
        .catch((error) => {
          setErrorMessage(error.message || 'Failed to update sensor');
          if (error.cause) {
            setAdvancedErrorMessage(String(error.cause));
          }
        })
        .finally(() => {
          actions.setSubmitting(false);
        });
    }
  return (
      <CardContent sx={{width: "100%", padding: 3, maxWidth: 650}}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          Edit Sensor Details
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
                    Submit
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