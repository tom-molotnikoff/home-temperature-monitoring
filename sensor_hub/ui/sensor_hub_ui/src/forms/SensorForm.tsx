import type {Sensor} from "../types/types.ts";
import {Formik, Field, Form, type FieldInputProps, type FormikProps} from 'formik';
import { Button, CardContent, Box, Stack, TextField, Typography, Alert } from '@mui/material';
import {useSensorForm} from "../hooks/useSensorForm.ts";

interface SensorFormProps {
  sensor?: Sensor;
  mode?: 'create' | 'edit';
  onSuccess?: (sensor: Sensor | null) => void;
}

type SensorFormValues = {
  name: string;
  type: string;
  url: string;
};

function SensorForm ({ sensor, mode = 'edit', onSuccess } : SensorFormProps) {
  const {
    initialValues,
    onSubmit,
    isSubmitting,
    successMessage,
    errorMessage,
    advancedErrorMessage,
  } = useSensorForm({ mode, initialSensor: sensor ?? null, onSuccess });

  return (
      <CardContent sx={{width: "100%", padding: 3, maxWidth: 650}}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          {mode === 'create' ? 'Add Sensor' : 'Edit Sensor Details'}
        </Typography>
        <Formik<SensorFormValues> initialValues={initialValues} enableReinitialize onSubmit={onSubmit}>
          {(formik: FormikProps<SensorFormValues>) => {
            const { isSubmitting: formikSubmitting, dirty, handleChange } = formik;
            return (
              <Form>
                <Stack spacing={2}>
                  <Field name="name">
                    {({ field }: { field: FieldInputProps<string> }) => (
                      <TextField
                        {...field}
                        label="Name"
                        variant="outlined"
                        fullWidth
                        size="small"
                        onChange={handleChange}
                      />
                    )}
                  </Field>

                  <Field name="type">
                    {({ field }: { field: FieldInputProps<string> }) => (
                      <TextField
                        {...field}
                        label="Type"
                        variant="outlined"
                        fullWidth
                        size="small"
                        onChange={handleChange}
                      />
                    )}
                  </Field>

                  <Field name="url">
                    {({ field }: { field: FieldInputProps<string> }) => (
                      <TextField
                        {...field}
                        label="API URL"
                        variant="outlined"
                        fullWidth
                        size="small"
                        onChange={handleChange}
                      />
                    )}
                  </Field>

                  <Box display="flex">
                    <Button
                      type="reset"
                      disabled={isSubmitting || formikSubmitting}
                      variant="outlined"
                      color="primary"
                      sx={{ mr: 2 }}
                    >
                      Reset
                    </Button>
                    <Button
                      type="submit"
                      disabled={isSubmitting || formikSubmitting || !dirty}
                      variant="contained"
                      color="primary"
                    >
                      {mode === 'create' ? 'Create' : 'Submit'}
                    </Button>
                  </Box>
                </Stack>
              </Form>
            );
          }}
        </Formik>
        {successMessage && (
          <Box mt={2}>
            <Alert severity="success" onClose={() =>{}}>
              {successMessage}
            </Alert>
          </Box>
        )}
        {errorMessage && (
          <Box mt={2}>
            <Alert severity="error" onClose={() => {}}>
              {errorMessage}
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

