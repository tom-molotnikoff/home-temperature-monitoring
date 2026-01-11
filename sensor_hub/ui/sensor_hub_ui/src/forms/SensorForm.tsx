import type {Sensor} from "../types/types.ts";
import {Formik, Field, Form, type FieldInputProps, type FormikProps} from 'formik';
import { Button, CardContent, Box, Stack, TextField, Typography, Alert, MenuItem } from '@mui/material';
import {useSensorForm} from "../hooks/useSensorForm.ts";
import {SensorTypes} from "../types/types.ts";
import * as Yup from 'yup';
import type {AuthUser} from "../providers/AuthContext.tsx";
import {hasPerm} from "../tools/Utils.ts";

interface SensorFormProps {
  sensor?: Sensor;
  mode?: 'create' | 'edit';
  onSuccess?: (sensor: Sensor | null) => void;
  user: AuthUser;
}

type SensorFormValues = {
  name: string;
  type: string;
  url: string;
};

const FormValidationSchema = Yup.object().shape({
  name: Yup.string().required('Name is required'),
  type: Yup.string().oneOf(SensorTypes, 'Invalid sensor type').required('Type is required'),
  url: Yup.string().required('API URL is required'),
})

function SensorForm ({ sensor, mode = 'edit', onSuccess, user } : SensorFormProps) {
  const {
    initialValues,
    onSubmit,
    isSubmitting,
    successMessage,
    errorMessage,
    advancedErrorMessage,
  } = useSensorForm({ mode, initialSensor: sensor ?? null, onSuccess });

  const fieldsDisabled = !(hasPerm(user, "manage_sensors"));

  return (
      <CardContent sx={{width: "100%", padding: 3, maxWidth: 650}}>
        <Typography variant="h6" sx={{ mb: 2 }}>
          {mode === 'create' ? 'Add Sensor' : 'Edit Sensor Details'}
        </Typography>
        <Formik<SensorFormValues>
          initialValues={initialValues}
          enableReinitialize onSubmit={onSubmit}
          validationSchema={FormValidationSchema}
        >
          {(formik: FormikProps<SensorFormValues>) => {
            const { isSubmitting: formikSubmitting, dirty, handleChange, errors, touched } = formik;
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
                        disabled={fieldsDisabled}
                      />
                    )}
                  </Field>

                  {errors.name && touched.name && (
                    <Typography variant="body2" color="error">
                      {errors.name}
                    </Typography>
                  )}

                  <Field name="type">
                    {({ field }: { field: FieldInputProps<string> }) => (
                      <TextField
                        {...field}
                        label="Type"
                        variant="outlined"
                        fullWidth
                        select
                        size="small"
                        disabled={fieldsDisabled}
                        onChange={handleChange}
                      >
                        {SensorTypes.map((type) => (
                          <MenuItem key={type} value={type}>
                            {type}
                          </MenuItem>
                        ))}
                      </TextField>
                    )}
                  </Field>

                  {errors.type && touched.type && (
                    <Typography variant="body2" color="error">
                      {errors.type}
                    </Typography>
                  )}

                  <Field name="url">
                    {({ field }: { field: FieldInputProps<string> }) => (
                      <TextField
                        {...field}
                        label="API URL"
                        variant="outlined"
                        fullWidth
                        size="small"
                        onChange={handleChange}
                        disabled={fieldsDisabled}
                      />
                    )}
                  </Field>

                  {errors.url && touched.url && (
                    <Typography variant="body2" color="error">
                      {errors.url}
                    </Typography>
                  )}

                  <Box display="flex">
                    <Button
                      type="reset"
                      disabled={isSubmitting || formikSubmitting || fieldsDisabled}
                      variant="outlined"
                      color="primary"
                      sx={{ mr: 2 }}
                    >
                      Reset
                    </Button>
                    <Button
                      type="submit"
                      disabled={isSubmitting || formikSubmitting || !dirty || fieldsDisabled}
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

