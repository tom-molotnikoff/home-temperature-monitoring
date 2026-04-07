import type {Sensor, DriverInfo} from "../types/types.ts";
import {Formik, Form, type FormikProps} from 'formik';
import { Button, CardContent, Box, Stack, TextField, Typography, Alert, MenuItem } from '@mui/material';
import {useSensorForm} from "../hooks/useSensorForm.ts";
import {useDrivers} from "../hooks/useDrivers.ts";
import * as Yup from 'yup';
import type {AuthUser} from "../providers/AuthContext.tsx";
import {hasPerm} from "../tools/Utils.ts";
import {useMemo} from "react";

interface SensorFormProps {
  sensor?: Sensor;
  mode?: 'create' | 'edit';
  onSuccess?: (sensor: Sensor | null) => void;
  user: AuthUser;
}

export type SensorFormValues = {
  name: string;
  sensorDriver: string;
  config: Record<string, string>;
};

function buildValidationSchema(selectedDriver: DriverInfo | undefined) {
  const shape: Record<string, Yup.Schema> = {
    name: Yup.string().required('Name is required'),
    sensorDriver: Yup.string().required('Sensor driver is required'),
  };
  if (selectedDriver) {
    const configShape: Record<string, Yup.StringSchema> = {};
    for (const field of selectedDriver.config_fields) {
      let s = Yup.string();
      if (field.required) {
        s = s.required(`${field.label} is required`);
      }
      configShape[field.key] = s;
    }
    shape.config = Yup.object().shape(configShape);
  }
  return Yup.object().shape(shape);
}

function SensorForm ({ sensor, mode = 'edit', onSuccess, user } : SensorFormProps) {
  const { drivers } = useDrivers();
  const {
    initialValues,
    onSubmit,
    isSubmitting,
    successMessage,
    errorMessage,
    advancedErrorMessage,
    setSuccessMessage,
    setErrorMessage,
    setAdvancedErrorMessage,
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
          validationSchema={buildValidationSchema(drivers.find(d => d.type === initialValues.sensorDriver))}
        >
          {(formik: FormikProps<SensorFormValues>) => {
            const { isSubmitting: formikSubmitting, dirty, errors, touched, values, setFieldValue } = formik;

            const selectedDriver = useMemo(
              () => drivers.find(d => d.type === values.sensorDriver),
              [drivers, values.sensorDriver]
            );

            return (
              <Form>
                <Stack spacing={2}>
                  <TextField
                    name="name"
                    label="Name"
                    variant="outlined"
                    fullWidth
                    size="small"
                    value={values.name}
                    onChange={formik.handleChange}
                    disabled={fieldsDisabled}
                    error={!!(errors.name && touched.name)}
                    helperText={touched.name && errors.name}
                  />

                  <TextField
                    name="sensorDriver"
                    label="Sensor Driver"
                    variant="outlined"
                    fullWidth
                    select
                    size="small"
                    disabled={fieldsDisabled}
                    value={values.sensorDriver}
                    onChange={(e) => {
                      const newDriver = e.target.value;
                      void setFieldValue('sensorDriver', newDriver);
                      // Reset config to defaults for the new driver
                      const driver = drivers.find(d => d.type === newDriver);
                      if (driver) {
                        const newConfig: Record<string, string> = {};
                        for (const f of driver.config_fields) {
                          newConfig[f.key] = f.default ?? '';
                        }
                        void setFieldValue('config', newConfig);
                      } else {
                        void setFieldValue('config', {});
                      }
                    }}
                    error={!!(errors.sensorDriver && touched.sensorDriver)}
                    helperText={touched.sensorDriver && errors.sensorDriver}
                  >
                    {drivers.map((driver) => (
                      <MenuItem key={driver.type} value={driver.type}>
                        {driver.display_name}
                      </MenuItem>
                    ))}
                  </TextField>

                  {selectedDriver?.config_fields.map((field) => {
                    const configErrors = errors.config as Record<string, string> | undefined;
                    const configTouched = touched.config as Record<string, boolean> | undefined;
                    return (
                      <TextField
                        key={field.key}
                        name={`config.${field.key}`}
                        label={field.label}
                        helperText={
                          (configTouched?.[field.key] && configErrors?.[field.key])
                            ? configErrors[field.key]
                            : field.description
                        }
                        error={!!(configTouched?.[field.key] && configErrors?.[field.key])}
                        type={field.sensitive ? 'password' : 'text'}
                        variant="outlined"
                        fullWidth
                        size="small"
                        required={field.required}
                        value={values.config?.[field.key] ?? field.default ?? ''}
                        onChange={formik.handleChange}
                        disabled={fieldsDisabled}
                      />
                    );
                  })}

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
            <Alert severity="success" onClose={() => setSuccessMessage(null)}>
              {successMessage}
            </Alert>
          </Box>
        )}
        {errorMessage && (
          <Box mt={2}>
            <Alert severity="error" onClose={() => { setErrorMessage(null); setAdvancedErrorMessage(null); }}>
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

