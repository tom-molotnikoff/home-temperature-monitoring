# Multi-Card Page Component Extraction Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Extract inline card compositions from TemperatureDashboard, SensorsOverview, and SensorPage into self-contained card components, making these multi-card pages follow the same thin-shell pattern as all other pages.

**Architecture:** Each inline composition (DateContextProvider + LayoutCard + chart, SensorsDataGrid with config props) becomes a self-contained component owning its own state and data fetching. Pages retain only layout (Grid sizing with `useIsMobile()`) and permission-based visibility (`hasPerm` at page level). The `graphContainerStyle` constant duplicated across files is absorbed into each card.

**Tech Stack:** React 19, MUI 7 (Grid v2), TypeScript

---

### Task 1: Create IndoorTemperatureDataCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/IndoorTemperatureDataCard.tsx`

Extracted from TemperatureDashboard lines 40-54. Self-contained card that:
- Owns `useHourlyAverages` state (previously in page)
- Uses `useIsMobile()` internally for `compact` prop on TemperatureGraph
- Uses `useSensorContext()` to get sensors
- Wraps content in DateContextProvider + LayoutCard with graphContainerStyle

```tsx
import { useState, type CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import HourlyAveragesToggle from './HourlyAveragesToggle';
import TemperatureGraph from './TemperatureGraph';
import { useSensorContext } from '../hooks/useSensorContext';
import { useIsMobile } from '../hooks/useMobile';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

export default function IndoorTemperatureDataCard() {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const { sensors } = useSensorContext();
  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <LayoutCard variant="secondary" changes={graphContainerStyle}>
        <TypographyH2>Indoor Temperature Data</TypographyH2>
        <DateRangePicker />
        <HourlyAveragesToggle
          useHourlyAverages={useHourlyAverages}
          setUseHourlyAverages={setUseHourlyAverages}
        />
        <TemperatureGraph
          sensors={sensors}
          useHourlyAverages={useHourlyAverages}
          compact={isMobile}
        />
      </LayoutCard>
    </DateContextProvider>
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 2: Create WeatherDataCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherDataCard.tsx`

Extracted from TemperatureDashboard lines 55-63. Self-contained card that:
- Uses `useIsMobile()` internally for `compact` prop on WeatherChart
- Wraps content in DateContextProvider + LayoutCard with graphContainerStyle

```tsx
import type { CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import WeatherChart from './WeatherGraph';
import { useIsMobile } from '../hooks/useMobile';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

export default function WeatherDataCard() {
  const isMobile = useIsMobile();

  return (
    <DateContextProvider>
      <LayoutCard variant="secondary" changes={graphContainerStyle}>
        <TypographyH2>Sheffield Weather Data</TypographyH2>
        <DateRangePicker />
        <WeatherChart compact={isMobile} />
      </LayoutCard>
    </DateContextProvider>
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 3: Create TemperatureSensorsCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/TemperatureSensorsCard.tsx`

Thin wrapper around SensorsDataGrid with temperature-specific configuration. Self-contained:
- Gets sensors from `useSensorContext()`, filters to temperature type
- Gets user from `useAuth()`

```tsx
import SensorsDataGrid from './SensorsDataGrid';
import { useSensorContext } from '../hooks/useSensorContext';
import { useAuth } from '../providers/AuthContext';

export default function TemperatureSensorsCard() {
  const { sensors } = useSensorContext();
  const { user } = useAuth();

  if (!user) return null;

  const temperatureSensors = sensors.filter(s => s.type === 'Temperature');

  return (
    <SensorsDataGrid
      sensors={temperatureSensors}
      cardHeight="100%"
      showReason={false}
      showType={false}
      showEnabled={true}
      title="Temperature Sensors"
      user={user}
    />
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 4: Rewrite TemperatureDashboard as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/temperature-dashboard/TemperatureDashboard.tsx`

Replace entire file contents. Page retains only:
- `useAuth()` for loading guard and `hasPerm` checks
- `useIsMobile()` for grid sizing
- Grid layout with 4 card components

```tsx
import CurrentTemperatures from '../../components/CurrentTemperatures';
import IndoorTemperatureDataCard from '../../components/IndoorTemperatureDataCard';
import WeatherDataCard from '../../components/WeatherDataCard';
import TemperatureSensorsCard from '../../components/TemperatureSensorsCard';
import PageContainer from '../../tools/PageContainer';
import { useIsMobile } from '../../hooks/useMobile';
import { Box, Grid } from '@mui/material';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

function TemperatureDashboard() {
  const { user } = useAuth();
  const isMobile = useIsMobile();

  return (
    <PageContainer titleText="Temperature Dashboard" loading={user === undefined}>
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_readings') && (
            <>
              <Grid size={12}><IndoorTemperatureDataCard /></Grid>
              <Grid size={12}><WeatherDataCard /></Grid>
            </>
          )}
          {hasPerm(user, 'view_sensors') && (
            <Grid size={isMobile ? 12 : 6}><TemperatureSensorsCard /></Grid>
          )}
          {hasPerm(user, 'view_readings') && (
            <Grid size={isMobile ? 12 : 6}><CurrentTemperatures cardHeight="100%" /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default TemperatureDashboard;
```

**Step 1:** Replace the full file content.

**Step 2:** Touch the file to trigger HMR. Open browser at `http://localhost:3000/` and verify:
- Indoor Temperature Data chart renders with date picker and hourly averages toggle
- Sheffield Weather Data chart renders
- Temperature Sensors DataGrid renders
- Current Temperatures card renders

### Task 5: Create AllSensorsCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/AllSensorsCard.tsx`

Thin wrapper around SensorsDataGrid with overview configuration. Self-contained:
- Gets sensors from `useSensorContext()`
- Gets user from `useAuth()`

```tsx
import SensorsDataGrid from './SensorsDataGrid';
import { useSensorContext } from '../hooks/useSensorContext';
import { useAuth } from '../providers/AuthContext';

export default function AllSensorsCard() {
  const { sensors } = useSensorContext();
  const { user } = useAuth();

  if (!user) return null;

  return (
    <SensorsDataGrid
      cardHeight="500px"
      sensors={sensors}
      showReason={true}
      showType={true}
      showEnabled={true}
      user={user}
    />
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 6: Rewrite SensorsOverview as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensors-overview/SensorsOverview.tsx`

Replace entire file. Page retains:
- `useAuth()` for loading guard and `hasPerm`
- `useIsMobile()` for grid sizing
- Grid layout with 5 self-contained cards

```tsx
import PageContainer from '../../tools/PageContainer';
import { useIsMobile } from '../../hooks/useMobile';
import { Grid, Box } from '@mui/material';
import SensorHealthCard from '../../components/SensorHealthCard';
import AddNewSensor from '../../components/AddNewSensor';
import SensorTypeCard from '../../components/SensorTypeCard';
import TotalReadingsForEachSensorCard from '../../components/TotalReadingsForEachSensorCard';
import AllSensorsCard from '../../components/AllSensorsCard';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

function SensorsOverview() {
  const isMobile = useIsMobile();
  const { user } = useAuth();

  return (
    <PageContainer titleText="Sensors Overview" loading={user === undefined}>
      <Box sx={{ width: '100%', flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%', width: '100%', flexGrow: 1 }}>
          {hasPerm(user, 'manage_sensors') && (
            <Grid size={isMobile ? 12 : 4}><AddNewSensor /></Grid>
          )}
          {hasPerm(user, 'view_sensors') && (
            <>
              <Grid size={isMobile ? 12 : 4}><SensorHealthCard /></Grid>
              <Grid size={isMobile ? 12 : 4}><SensorTypeCard /></Grid>
              <Grid size={isMobile ? 12 : 8}><AllSensorsCard /></Grid>
              <Grid size={isMobile ? 12 : 4}><TotalReadingsForEachSensorCard /></Grid>
            </>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default SensorsOverview;
```

**Step 1:** Replace the full file content.

**Step 2:** Touch the file to trigger HMR. Open browser at `http://localhost:3000/sensors-overview` and verify:
- Add New Sensor form renders
- Sensor Health and Sensor Type cards render
- All Sensors DataGrid renders
- Total Readings card renders

---

### Task 7: Create SensorHealthHistoryChartCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChartCard.tsx`

Extracted from SensorPage lines 64-69. Wraps SensorHealthHistoryChart in LayoutCard with graphContainerStyle. Receives `sensor` prop (identifying which sensor to display).

```tsx
import type { CSSProperties } from 'react';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import SensorHealthHistoryChart from './SensorHealthHistoryChart';
import type { Sensor } from '../api/Sensors';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

interface SensorHealthHistoryChartCardProps {
  sensor: Sensor;
}

export default function SensorHealthHistoryChartCard({ sensor }: SensorHealthHistoryChartCardProps) {
  return (
    <LayoutCard variant="secondary" changes={graphContainerStyle}>
      <TypographyH2>Sensor Health History</TypographyH2>
      <SensorHealthHistoryChart sensor={sensor} limit={5000} />
    </LayoutCard>
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 8: Create SensorTemperatureDataCard

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/SensorTemperatureDataCard.tsx`

Extracted from SensorPage lines 70-84. Self-contained card that:
- Owns `useHourlyAverages` state (previously in page)
- Uses `useIsMobile()` internally for `compact` prop
- Only renders content if `sensor.type === "Temperature"` (otherwise renders empty LayoutCard or null)
- Wraps in DateContextProvider + LayoutCard with graphContainerStyle

```tsx
import { useState, type CSSProperties } from 'react';
import { DateContextProvider } from '../providers/DateContextProvider';
import LayoutCard from '../tools/LayoutCard';
import { TypographyH2 } from '../tools/Typography';
import DateRangePicker from './DateRangePicker';
import HourlyAveragesToggle from './HourlyAveragesToggle';
import TemperatureGraph from './TemperatureGraph';
import { useIsMobile } from '../hooks/useMobile';
import type { Sensor } from '../api/Sensors';

const graphContainerStyle: CSSProperties = {
  flex: 1,
  flexGrow: 1,
  height: '100%',
  alignItems: 'center',
};

interface SensorTemperatureDataCardProps {
  sensor: Sensor;
}

export default function SensorTemperatureDataCard({ sensor }: SensorTemperatureDataCardProps) {
  const [useHourlyAverages, setUseHourlyAverages] = useState(true);
  const isMobile = useIsMobile();

  if (sensor.type !== 'Temperature') return null;

  return (
    <DateContextProvider>
      <LayoutCard variant="secondary" changes={graphContainerStyle}>
        <TypographyH2>Indoor Temperature Data</TypographyH2>
        <DateRangePicker />
        <HourlyAveragesToggle
          useHourlyAverages={useHourlyAverages}
          setUseHourlyAverages={setUseHourlyAverages}
        />
        <TemperatureGraph
          sensors={[sensor]}
          useHourlyAverages={useHourlyAverages}
          compact={isMobile}
        />
      </LayoutCard>
    </DateContextProvider>
  );
}
```

**Step 1:** Create the file with the code above.

**Step 2:** Touch the file to trigger HMR. Verify no compile errors.

---

### Task 9: Rewrite SensorPage as thin shell

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensor/SensorPage.tsx`

Replace entire file. Page retains:
- `useSensorContext()` for sensor lookup by ID
- `useAuth()` for loading guard and `hasPerm`
- `useIsMobile()` for grid sizing
- Sensor-not-found guard
- Grid layout with 5 card components

```tsx
import { Box, Grid } from '@mui/material';
import PageContainer from '../../tools/PageContainer';
import { useSensorContext } from '../../hooks/useSensorContext';
import { useIsMobile } from '../../hooks/useMobile';
import SensorInfoCard from '../../components/SensorInfoCard';
import EditSensorDetails from '../../components/EditSensorDetails';
import SensorHealthHistory from '../../components/SensorHealthHistory';
import SensorHealthHistoryChartCard from '../../components/SensorHealthHistoryChartCard';
import SensorTemperatureDataCard from '../../components/SensorTemperatureDataCard';
import { useAuth } from '../../providers/AuthContext';
import { hasPerm } from '../../tools/Utils';

interface SensorPageProps {
  sensorId: number;
}

function SensorPage({ sensorId }: SensorPageProps) {
  const { sensors } = useSensorContext();
  const sensor = sensors.find(s => s.id === sensorId);
  const { user } = useAuth();
  const isMobile = useIsMobile();

  if (user === undefined) {
    return <PageContainer titleText="Sensor" loading />;
  }

  if (!sensor) {
    return (
      <PageContainer titleText="Sensor Not Found">
        <Box sx={{ flexGrow: 1 }}>
          <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
            <Grid size={12}>
              <h2>Sensor with ID {sensorId} not found.</h2>
            </Grid>
          </Grid>
        </Box>
      </PageContainer>
    );
  }

  return (
    <PageContainer titleText="Sensor">
      <Box sx={{ flexGrow: 1 }}>
        <Grid container spacing={2} alignItems="stretch" sx={{ minHeight: '100%' }}>
          {hasPerm(user, 'view_sensors') && (
            <>
              <Grid size={isMobile ? 12 : 6}><SensorInfoCard sensor={sensor} user={user} /></Grid>
              <Grid size={isMobile ? 12 : 6}><EditSensorDetails sensor={sensor} /></Grid>
            </>
          )}
          {hasPerm(user, 'view_readings') && (
            <>
              <Grid size={isMobile ? 12 : 6}><SensorHealthHistoryChartCard sensor={sensor} /></Grid>
              <Grid size={isMobile ? 12 : 6}><SensorTemperatureDataCard sensor={sensor} /></Grid>
            </>
          )}
          {hasPerm(user, 'view_sensors') && (
            <Grid size={isMobile ? 12 : 6}><SensorHealthHistory sensor={sensor} /></Grid>
          )}
        </Grid>
      </Box>
    </PageContainer>
  );
}

export default SensorPage;
```

**Step 1:** Replace the full file content.

**Step 2:** Touch the file to trigger HMR. Open browser at a sensor page (e.g., `http://localhost:3000/sensor/1`) and verify:
- Sensor Info card renders
- Edit Sensor Details form renders
- Sensor Health History Chart renders
- Temperature Data chart renders (only for temperature sensors)
- Sensor Health History table renders

---

### Task 10: Browser verification of all 3 pages

**Step 1:** Navigate to `http://localhost:3000/` (TemperatureDashboard). Verify all 4 cards render correctly.

**Step 2:** Navigate to `http://localhost:3000/sensors-overview`. Verify all 5 cards render correctly.

**Step 3:** Navigate to a sensor detail page. Verify all cards render correctly.

**Step 4:** Verify the hourly averages toggle works on the dashboard (state is now internal to IndoorTemperatureDataCard).
