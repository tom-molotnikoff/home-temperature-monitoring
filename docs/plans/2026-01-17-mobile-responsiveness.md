# Mobile Responsiveness Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Improve mobile responsiveness across all UI pages so content doesn't squish or clip on screens < 950px.

**Architecture:** Use existing `useIsMobile()` hook (breakpoint 950px) to conditionally adjust layouts. For grids, switch from side-by-side to stacked columns. For DataGrids, hide secondary columns. For charts, show simplified versions with reduced height and simplified legends.

**Tech Stack:** React, MUI Grid/DataGrid, Recharts, existing `useIsMobile()` hook

---

## Task 1: SensorPage - Make Grids Responsive

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensor/SensorPage.tsx`

**Step 1: Import useIsMobile hook**

Add import at top of file:

```tsx
import { useIsMobile } from "../../hooks/useMobile.ts";
```

**Step 2: Add isMobile hook call inside component**

After line 24 (`const sensor = sensors.find...`), add:

```tsx
const isMobile = useIsMobile();
```

**Step 3: Update all Grid size={6} to be responsive**

Replace all instances of `Grid size={6}` with `Grid size={isMobile ? 12 : 6}`.

Specific changes:
- Line 61: `<Grid size={6}>` → `<Grid size={isMobile ? 12 : 6}>`
- Line 64: `<Grid size={6}>` → `<Grid size={isMobile ? 12 : 6}>`
- Line 72: `<Grid size={6}>` → `<Grid size={isMobile ? 12 : 6}>`
- Line 78: `<Grid size={6}>` → `<Grid size={isMobile ? 12 : 6}>`
- Line 96: `<Grid size={6}>` → `<Grid size={isMobile ? 12 : 6}>`

**Step 4: Remove hardcoded width on container**

Line 59: Change `sx={{ minHeight: "100%", width: "98vw" }}` to `sx={{ minHeight: "100%" }}` to prevent overflow on mobile.

**Step 5: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

**Step 6: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/sensor/SensorPage.tsx
git commit -m "feat(ui): make SensorPage responsive on mobile"
```

---

## Task 2: SensorsOverview - Fix Remaining Hardcoded Grid Sizes

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/sensors-overview/SensorsOverview.tsx`

**Step 1: Make SensorsDataGrid and TotalReadingsCard responsive**

Lines 61-72 have `Grid size={8}` and `Grid size={4}` which are hardcoded.

Change:
```tsx
<Grid size={8}>
  <SensorsDataGrid .../>
</Grid>
<Grid size={4}>
  <TotalReadingsForEachSensorCard />
</Grid>
```

To:
```tsx
<Grid size={isMobile ? 12 : 8}>
  <SensorsDataGrid .../>
</Grid>
<Grid size={isMobile ? 12 : 4}>
  <TotalReadingsForEachSensorCard />
</Grid>
```

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/sensors-overview/SensorsOverview.tsx
git commit -m "feat(ui): make SensorsOverview fully responsive on mobile"
```

---

## Task 3: TemperatureDashboard - Show Simplified Graphs on Mobile

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/temperature-dashboard/TemperatureDashboard.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/TemperatureGraph.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherGraph.tsx`

### Step 1: Update TemperatureDashboard to show graphs on mobile

Replace lines 57-84 (the conditional that hides graphs on mobile):

FROM:
```tsx
{isMobile ? null : (
  hasPerm(user, "view_readings") &&
    <>
      <Grid size={12} sx={{width: "98vw"}}>
        ...
      </Grid>
      <Grid size={12} sx={{width: "98vw"}}>
        ...
      </Grid>
    </>
)}
```

TO:
```tsx
{hasPerm(user, "view_readings") && (
  <>
    <Grid size={12}>
      <DateContextProvider>
        <LayoutCard variant="secondary" changes={graphContainerStyle}>
          <TypographyH2>Indoor Temperature Data</TypographyH2>
          <DateRangePicker/>
          <HourlyAveragesToggle
            useHourlyAverages={useHourlyAverages}
            setUseHourlyAverages={setUseHourlyAverages}/>
          <TemperatureGraph
            sensors={sensors}
            useHourlyAverages={useHourlyAverages}
            compact={isMobile}/>
        </LayoutCard>
      </DateContextProvider>
    </Grid>
    <Grid size={12}>
      <DateContextProvider>
        <LayoutCard variant="secondary" changes={graphContainerStyle}>
          <TypographyH2>Sheffield Weather Data</TypographyH2>
          <DateRangePicker/>
          <WeatherChart compact={isMobile}/>
        </LayoutCard>
      </DateContextProvider>
    </Grid>
  </>
)}
```

Note: Removed `sx={{width: "98vw"}}` to prevent horizontal overflow, and added `compact` prop.

### Step 2: Update TemperatureGraph to support compact mode

In `TemperatureGraph.tsx`, update the component to accept and use a `compact` prop:

Update props interface (around line 23):
```tsx
const TemperatureGraph = React.memo(function TemperatureGraph({
  sensors,
  useHourlyAverages,
  compact = false,
}: {
  sensors: Sensor[];
  useHourlyAverages: boolean;
  compact?: boolean;
}) {
```

Update the chart rendering section (inside ResponsiveContainer):
- Reduce height: `height={compact ? 250 : 350}`
- Simplify X-axis labels on compact: hide tick labels or reduce tick count
- Simplify legend: use `wrapperStyle` to make it scrollable or simpler

Replace the return JSX starting at line 54:
```tsx
return (
  <div data-testid="temperature-graph" style={{ ...graphContainerStyle, height: compact ? 250 : 350 }}>
    {!Array.isArray(chartData) || chartData.length === 0 ? (
      <></>
    ) : (
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <CartesianGrid stroke="#eee" strokeDasharray="3 3" />
          <XAxis
            dataKey="time"
            tickFormatter={(t) => {
              const date = new Date(t);
              return compact 
                ? date.toLocaleTimeString([], { hour: '2-digit' })
                : date.toLocaleTimeString();
            }}
            interval={compact ? "preserveStartEnd" : 0}
            tick={{ fontSize: compact ? 10 : 12 }}
          />
          <YAxis type="number" domain={[12, 26]} tick={{ fontSize: compact ? 10 : 12 }} />
          <Tooltip />
          <Legend 
            onClick={legendClickHandler}
            wrapperStyle={compact ? { fontSize: 10 } : undefined}
          />
          {sensors.map((sensor, index) => (
            <Line
              key={sensor.name}
              type="natural"
              dataKey={sensor.name}
              stroke={lineColours[index]}
              dot={false}
              connectNulls={true}
              animationEasing="ease-in-out"
              animationDuration={800}
              hide={linesHidden[sensor.name]}
              legendType="plainline"
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    )}
  </div>
);
```

Remove the hardcoded `graphContainerStyle` const (lines 90-93) and replace with:
```tsx
const graphContainerStyle: CSSProperties = {
  width: "100%",
};
```

### Step 3: Update WeatherGraph to support compact mode

In `WeatherGraph.tsx`, update to accept a `compact` prop:

Update component signature (line 15):
```tsx
export default function WeatherChart({ compact = false }: { compact?: boolean }) {
```

Update ResponsiveContainer and chart (line 54):
```tsx
<ResponsiveContainer width="100%" height={compact ? 200 : 300}>
  <LineChart data={data}>
    <CartesianGrid stroke="#eee" />
    <XAxis
      dataKey="time"
      tickFormatter={(t) => {
        const date = new Date(String(t));
        return compact 
          ? date.toLocaleTimeString([], { hour: '2-digit' })
          : date.toLocaleTimeString();
      }}
      interval={compact ? "preserveStartEnd" : 0}
      tick={{ fontSize: compact ? 10 : 12 }}
    />
    <YAxis tick={{ fontSize: compact ? 10 : 12 }} />
    <Tooltip
      labelFormatter={(t) => new Date(String(t)).toLocaleString()}
    />
    <Legend wrapperStyle={compact ? { fontSize: 10 } : undefined} />
    ...
  </LineChart>
</ResponsiveContainer>
```

### Step 4: Verify build succeeds

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

### Step 5: Commit

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/temperature-dashboard/TemperatureDashboard.tsx
git add sensor_hub/ui/sensor_hub_ui/src/components/TemperatureGraph.tsx
git add sensor_hub/ui/sensor_hub_ui/src/components/WeatherGraph.tsx
git commit -m "feat(ui): show simplified charts on mobile with compact mode"
```

---

## Task 4: AlertsPage - Hide Secondary Columns on Mobile

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

### Step 1: Import useIsMobile hook

Add import at top of file:
```tsx
import { useIsMobile } from '../../hooks/useMobile';
```

### Step 2: Add hook call in component

After line 56 (`const { sensors } = useSensorContext();`), add:
```tsx
const isMobile = useIsMobile();
```

### Step 3: Update columns definition to hide columns on mobile

Replace the columns definition (lines 188-197) with a function that filters based on mobile:

```tsx
const allColumns: GridColDef[] = [
  { field: 'SensorName', headerName: 'Sensor', flex: 1 },
  { field: 'AlertType', headerName: 'Alert Type', width: 150 },
  { field: 'HighThreshold', headerName: 'High', width: 80 },
  { field: 'LowThreshold', headerName: 'Low', width: 80 },
  { field: 'TriggerStatus', headerName: 'Status', width: 100 },
  { field: 'RateLimitHours', headerName: 'Rate Limit (hrs)', width: 130 },
  { field: 'Enabled', headerName: 'Enabled', width: 80 },
  { field: 'LastAlertSentAt', headerName: 'Last Alert Sent', width: 180 },
];

// On mobile: show Sensor, thresholds, status, enabled
// Hide: AlertType, RateLimitHours, LastAlertSentAt
const mobileHiddenFields = ['AlertType', 'RateLimitHours', 'LastAlertSentAt'];
const columns = isMobile 
  ? allColumns.filter(col => !mobileHiddenFields.includes(col.field))
  : allColumns;
```

### Step 4: Verify build succeeds

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

### Step 5: Commit

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx
git commit -m "feat(ui): hide secondary columns on AlertsPage for mobile"
```

---

## Task 5: SessionsPage - Hide Secondary Columns and Show Short User Agent

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/account/SessionsPage.tsx`

### Step 1: Import useIsMobile hook

Add import at top of file:
```tsx
import { useIsMobile } from '../../hooks/useMobile';
```

### Step 2: Add hook call in component

After line 13 (`const [sessions, setSessions] = useState...`), add:
```tsx
const isMobile = useIsMobile();
```

### Step 3: Create a helper function to parse short device info from user agent

Add before the columns definition:
```tsx
const getShortDeviceInfo = (userAgent: string): string => {
  if (!userAgent) return 'Unknown';
  // Check for common patterns
  if (userAgent.includes('iPhone')) return 'iPhone';
  if (userAgent.includes('iPad')) return 'iPad';
  if (userAgent.includes('Android')) return 'Android';
  if (userAgent.includes('Windows')) return 'Windows';
  if (userAgent.includes('Mac')) return 'Mac';
  if (userAgent.includes('Linux')) return 'Linux';
  // Fallback: first 20 chars
  return userAgent.substring(0, 20) + '...';
};
```

### Step 4: Update columns definition with mobile variants

Replace the columns definition (lines 25-46) with:

```tsx
const allColumns: GridColDef[] = [
  { field: 'id', headerName: 'ID', width: 80 },
  { field: 'ip_address', headerName: 'IP', flex: 1 },
  { field: 'user_agent', headerName: 'User Agent', flex: 2 },
  { field: 'created_at', headerName: 'Created', width: 180 },
  { field: 'last_accessed_at', headerName: 'Last Accessed', width: 180 },
  { field: 'expires_at', headerName: 'Expires', width: 180 },
  {
    field: 'actions', headerName: ' ', width: 120, renderCell: (params) => (
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        { params.row.current ? <Tooltip title="Current session"><CheckIcon color="success"/></Tooltip> : null }
        <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
          <span>
            <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async ()=>{ await revoke(params.row.id as number); }}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          </span>
        </Tooltip>
      </div>
    )
  }
];

const mobileColumns: GridColDef[] = [
  { 
    field: 'device', 
    headerName: 'Device', 
    flex: 1,
    valueGetter: (value, row) => getShortDeviceInfo(row.user_agent),
  },
  { field: 'last_accessed_at', headerName: 'Last Active', width: 140 },
  {
    field: 'actions', headerName: ' ', width: 80, renderCell: (params) => (
      <div style={{ display: 'flex', gap: 4, alignItems: 'center' }}>
        { params.row.current ? <Tooltip title="Current session"><CheckIcon color="success" fontSize="small"/></Tooltip> : null }
        <Tooltip title={params.row.current ? 'Cannot revoke current session' : 'Revoke session'}>
          <span>
            <IconButton aria-label="revoke" size="small" disabled={params.row.current} onClick={async ()=>{ await revoke(params.row.id as number); }}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          </span>
        </Tooltip>
      </div>
    )
  }
];

const columns = isMobile ? mobileColumns : allColumns;
```

### Step 5: Verify build succeeds

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

### Step 6: Commit

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/account/SessionsPage.tsx
git commit -m "feat(ui): show short device info on SessionsPage for mobile"
```

---

## Task 6: UsersPage - Hide Secondary Columns on Mobile

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/UsersPage.tsx`

### Step 1: Import useIsMobile hook

Add import at top of file:
```tsx
import { useIsMobile } from '../../hooks/useMobile';
```

### Step 2: Add hook call in component

After line 42 (`const { user } = useAuth();`), add:
```tsx
const isMobile = useIsMobile();
```

### Step 3: Update columns definition to hide columns on mobile

Replace the columns definition (lines 139-145) with:

```tsx
const allColumns: GridColDef[] = [
  { field: 'id', headerName: 'ID', width: 80 },
  { field: 'username', headerName: 'Username', flex: 1 },
  { field: 'email', headerName: 'Email', flex: 1 },
  { field: 'rolesDisplay', headerName: 'Roles', flex: 1 },
  { field: 'must_change_password', headerName: 'Must change password', width: 200 },
];

// On mobile: show Username, Roles only
// Hide: ID, Email, must_change_password
const mobileHiddenFields = ['id', 'email', 'must_change_password'];
const columns = isMobile 
  ? allColumns.filter(col => !mobileHiddenFields.includes(col.field))
  : allColumns;
```

### Step 4: Verify build succeeds

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

### Step 5: Commit

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/admin/UsersPage.tsx
git commit -m "feat(ui): hide secondary columns on UsersPage for mobile"
```

---

## Task 7: OAuthPage - Fix Button Layout Clipping

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/admin/OAuthPage.tsx`

### Step 1: Import useIsMobile hook

Add import at top of file:
```tsx
import { useIsMobile } from '../../hooks/useMobile';
```

### Step 2: Add hook call in component

After line 37 (`const { user } = useAuth();`), add:
```tsx
const isMobile = useIsMobile();
```

### Step 3: Update header layout to stack buttons on mobile

Replace lines 139-160 (the Box containing title and buttons):

FROM:
```tsx
<Box display="flex" alignItems="center" justifyContent="space-between" gap={2} mb={2}>
  <Typography variant="h4">OAuth Configuration</Typography>
  <Box display="flex" gap={1}>
    <Button
      variant="outlined"
      startIcon={<SyncIcon />}
      onClick={handleReload}
      disabled={loading || reloading || !canManage}
      title="Reload credentials.json from disk"
    >
      {reloading ? 'Reloading...' : 'Reload Config'}
    </Button>
    <Button
      variant="outlined"
      startIcon={<RefreshIcon />}
      onClick={loadStatus}
      disabled={loading}
    >
      Refresh
    </Button>
  </Box>
</Box>
```

TO:
```tsx
<Box 
  display="flex" 
  flexDirection={isMobile ? 'column' : 'row'}
  alignItems={isMobile ? 'flex-start' : 'center'} 
  justifyContent="space-between" 
  gap={2} 
  mb={2}
>
  <Typography variant="h4">OAuth Configuration</Typography>
  <Box display="flex" flexDirection={isMobile ? 'column' : 'row'} gap={1} width={isMobile ? '100%' : 'auto'}>
    <Button
      variant="outlined"
      startIcon={<SyncIcon />}
      onClick={handleReload}
      disabled={loading || reloading || !canManage}
      title="Reload credentials.json from disk"
      fullWidth={isMobile}
    >
      {reloading ? 'Reloading...' : 'Reload Config'}
    </Button>
    <Button
      variant="outlined"
      startIcon={<RefreshIcon />}
      onClick={loadStatus}
      disabled={loading}
      fullWidth={isMobile}
    >
      Refresh
    </Button>
  </Box>
</Box>
```

### Step 4: Verify build succeeds

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no errors

### Step 5: Commit

```bash
git add sensor_hub/ui/sensor_hub_ui/src/pages/admin/OAuthPage.tsx
git commit -m "feat(ui): fix button layout clipping on OAuthPage for mobile"
```

---

## Task 8: Final Build and Integration Test

### Step 1: Run full build

```bash
cd sensor_hub/ui/sensor_hub_ui && npm run build
```

Expected: Build succeeds with no TypeScript errors

### Step 2: Run linter if available

```bash
cd sensor_hub/ui/sensor_hub_ui && npm run lint
```

Expected: No linting errors (or only pre-existing ones)

### Step 3: Final commit for any cleanup

```bash
git add -A
git commit -m "chore: mobile responsiveness cleanup" --allow-empty
```

---

## Summary of Changes

| Page | Before | After |
|------|--------|-------|
| SensorPage | Hardcoded `Grid size={6}` | Responsive `size={isMobile ? 12 : 6}` |
| SensorsOverview | Partial responsive, some hardcoded | Fully responsive grid sizes |
| TemperatureDashboard | Graphs hidden on mobile | Simplified compact graphs shown |
| TemperatureGraph | Fixed 350px height | Compact mode: 250px, smaller fonts |
| WeatherGraph | Fixed 300px height | Compact mode: 200px, smaller fonts |
| AlertsPage | All 8 columns shown | Hides AlertType, RateLimitHours, LastAlertSentAt on mobile |
| SessionsPage | All 7 columns shown | Shows Device (short UA), Last Active, Actions on mobile |
| UsersPage | All 5 columns shown | Hides ID, Email, must_change_password on mobile |
| OAuthPage | Buttons clip off screen | Buttons stack vertically on mobile |
