# Mobile Improvements Phase 2 Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Fix remaining mobile UI issues: cramped X-axis labels, default date range, DateRangePicker widths, SensorHealthHistory controls clipping, and Alerts page DataGrid replacement.

**Architecture:** Use existing `useIsMobile()` hook to conditionally adjust chart tick frequency/rotation, date defaults, and component layouts. Replace Alerts DataGrid with card-based list on mobile.

**Tech Stack:** React, MUI, Recharts, existing `useIsMobile()` hook (breakpoint 950px)

---

## Task 1: Fix X-axis Label Cramping on TemperatureGraph

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/TemperatureGraph.tsx`

**Step 1: Update XAxis configuration for better tick frequency and mobile rotation**

Replace the current XAxis element with improved configuration:

```tsx
<XAxis
  dataKey="time"
  tickFormatter={(t) => {
    const date = new Date(t);
    return compact 
      ? date.toLocaleTimeString([], { hour: '2-digit' })
      : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }}
  interval="preserveStartEnd"
  minTickGap={compact ? 30 : 50}
  tick={{ fontSize: compact ? 10 : 12, angle: compact ? -45 : 0, textAnchor: compact ? 'end' : 'middle' }}
  height={compact ? 50 : 30}
/>
```

Key changes:
- `interval="preserveStartEnd"` - always show first and last, auto-skip middle
- `minTickGap` - minimum pixels between ticks (prevents cramping)
- `angle: -45` on mobile only
- `textAnchor: 'end'` when rotated (proper alignment)
- Increased `height` on mobile to accommodate rotated labels

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 2: Fix X-axis Label Cramping on WeatherGraph

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherGraph.tsx`

**Step 1: Update XAxis configuration matching TemperatureGraph pattern**

Replace the current XAxis element:

```tsx
<XAxis
  dataKey="time"
  tickFormatter={(t) => {
    const date = new Date(String(t));
    return compact 
      ? date.toLocaleTimeString([], { hour: '2-digit' })
      : date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }}
  interval="preserveStartEnd"
  minTickGap={compact ? 30 : 50}
  tick={{ fontSize: compact ? 10 : 12, angle: compact ? -45 : 0, textAnchor: compact ? 'end' : 'middle' }}
  height={compact ? 50 : 30}
/>
```

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 3: Default to 2 Days on Mobile in DateContextProvider

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/providers/DateContextProvider.tsx`

**Step 1: Import useIsMobile hook**

Add import at top:
```tsx
import { useIsMobile } from "../hooks/useMobile";
```

**Step 2: Use hook to set mobile-aware default**

Update the component to use a mobile-aware default:

```tsx
export function DateContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const isMobile = useIsMobile();
  
  const [startDate, setStartDate] = useState<DateTime | null>(
    DateTime.now().minus({ days: 7 }).startOf("day")
  );

  const [endDate, setEndDate] = useState<DateTime | null>(
    DateTime.now().plus({ days: 1 }).startOf("day")
  );

  // Adjust default range on mobile (only on initial mount)
  useEffect(() => {
    if (isMobile) {
      setStartDate(DateTime.now().minus({ days: 2 }).startOf("day"));
    }
  }, []); // Empty deps - only run once on mount
```

Note: We use an effect that runs once on mount to adjust for mobile, rather than changing the initial state, because `useIsMobile()` returns undefined initially during SSR/hydration.

**Step 3: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 4: Fix DateRangePicker Width on Mobile

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/DateRangePicker.tsx`

**Step 1: Add mobileChanges prop to DesktopRowMobileColumn**

The component already uses DesktopRowMobileColumn but may need explicit mobile styling. Add mobileChanges to ensure proper width:

```tsx
<DesktopRowMobileColumn
  testid="date-range-picker"
  desktopChanges={desktopLayoutStyleChanges}
  mobileChanges={mobileLayoutStyleChanges}
>
```

**Step 2: Add mobile style constant**

After the existing desktopLayoutStyleChanges, add:

```tsx
const mobileLayoutStyleChanges: CSSProperties = {
  width: "100%",
  alignItems: "stretch",
};
```

**Step 3: Make DatePickers full width on mobile**

Import useIsMobile and wrap DatePickers with responsive styling:

```tsx
import { useIsMobile } from "../hooks/useMobile";

function DateRangePicker() {
  const { startDate, setStartDate, endDate, setEndDate, invalidDate } =
    useContext(DateContext);
  const isMobile = useIsMobile();

  return (
    <CenteredFlex>
      <DesktopRowMobileColumn
        testid="date-range-picker"
        desktopChanges={desktopLayoutStyleChanges}
        mobileChanges={mobileLayoutStyleChanges}
      >
        <TestIdContainer testid="start-date-picker">
          <DatePicker
            label="Start Date"
            value={startDate}
            onChange={setStartDate}
            slotProps={{
              textField: {
                fullWidth: isMobile,
              },
            }}
          />
        </TestIdContainer>
        <TestIdContainer testid="end-date-picker">
          <DatePicker 
            label="End Date" 
            value={endDate} 
            onChange={setEndDate}
            slotProps={{
              textField: {
                fullWidth: isMobile,
              },
            }}
          />
        </TestIdContainer>
      </DesktopRowMobileColumn>
      {invalidDate && (
        <ErrorText
          message="Invalid date range"
          testid="invalid-date-range-error"
        />
      )}
    </CenteredFlex>
  );
}
```

**Step 4: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 5: Fix SensorHealthHistory Controls Clipping

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistory.tsx`

**Step 1: Import useIsMobile hook**

Add import at top:
```tsx
import { useIsMobile } from "../hooks/useMobile";
```

**Step 2: Add hook call in component**

After line 15 (const declarations), add:
```tsx
const isMobile = useIsMobile();
```

**Step 3: Update the controls div to be responsive**

Replace the controls div (lines 67-92) with responsive layout:

```tsx
<div style={{
  display: "flex",
  flexDirection: isMobile ? "column" : "row",
  justifyContent: isMobile ? "center" : "flex-end",
  alignItems: isMobile ? "stretch" : "center",
  flexGrow: 1,
  width: "100%",
  marginTop: 16,
  gap: 16
}}>
  <TextField
    label="Limit History Entries"
    type="number"
    defaultValue={5000}
    onChange={(e) => setLimitInput(e.target.value)}
    sx={{ mt: 2, width: isMobile ? "100%" : 200 }}
    fullWidth={isMobile}
  />
  <Button
    onClick={() => {
      setIsLoading(true);
      setLimit(parseInt(limitInput));
      refresh().then(() => {
        setIsLoading(false);
        setSnackbarOpen(true);
      });
    }}
    variant="outlined"
    startIcon={<RefreshIcon />}
    fullWidth={isMobile}
    sx={{
      mt: 2,
      alignSelf: 'center',
      height: "56px",
    }}
  >
    Refresh
  </Button>
</div>
```

**Step 4: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 6: Replace Alerts DataGrid with Card List on Mobile

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/alerts/AlertsPage.tsx`

**Step 1: Create AlertRuleCard component (inline in same file)**

Add this component before the main AlertsPage function:

```tsx
interface AlertRuleCardProps {
  rule: AlertRule;
  onClick: (event: React.MouseEvent) => void;
}

function AlertRuleCard({ rule, onClick }: AlertRuleCardProps) {
  return (
    <Box
      onClick={onClick}
      sx={{
        p: 2,
        mb: 1,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 1,
        cursor: 'pointer',
        '&:hover': {
          backgroundColor: 'action.hover',
        },
      }}
    >
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
        <Typography variant="subtitle1" fontWeight="bold">
          {rule.SensorName}
        </Typography>
        <Chip
          label={rule.Enabled ? 'Enabled' : 'Disabled'}
          color={rule.Enabled ? 'success' : 'default'}
          size="small"
        />
      </Box>
      <Typography variant="body2" color="text.secondary">
        {rule.AlertType === 'numeric_range' 
          ? `Range: ${rule.LowThreshold ?? '-'} to ${rule.HighThreshold ?? '-'}`
          : `Trigger: ${rule.TriggerStatus || '-'}`
        }
      </Typography>
      {rule.LastAlertSentAt && (
        <Typography variant="caption" color="text.secondary">
          Last alert: {new Date(rule.LastAlertSentAt).toLocaleDateString()}
        </Typography>
      )}
    </Box>
  );
}
```

**Step 2: Add Chip import**

Update the MUI imports to include Chip:
```tsx
import {
  Button,
  Box,
  Menu,
  MenuItem,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  InputLabel,
  FormControl,
  FormControlLabel,
  Switch,
  Chip,
} from '@mui/material';
```

**Step 3: Update the render to conditionally show cards or DataGrid**

Find the DataGrid section (around line 233-241) and wrap it with a conditional:

Replace:
```tsx
<div style={{ height: 400, width: '100%' }}>
  <DataGrid 
    rows={rows} 
    columns={columns} 
    pageSizeOptions={[5, 10, 25]} 
    initialState={{ pagination: { paginationModel: { pageSize: 10 } } }} 
    onRowClick={handleRowClick} 
  />
</div>
```

With:
```tsx
{isMobile ? (
  <Box sx={{ width: '100%', maxHeight: 400, overflowY: 'auto' }}>
    {alertRules.length === 0 ? (
      <Typography color="text.secondary" sx={{ p: 2, textAlign: 'center' }}>
        No alert rules configured
      </Typography>
    ) : (
      alertRules.map((rule) => (
        <AlertRuleCard
          key={rule.SensorID}
          rule={rule}
          onClick={(event) => {
            setSelectedRow(rule);
            setMenuAnchorEl(event.currentTarget as HTMLElement);
          }}
        />
      ))
    )}
  </Box>
) : (
  <div style={{ height: 400, width: '100%' }}>
    <DataGrid 
      rows={rows} 
      columns={columns} 
      pageSizeOptions={[5, 10, 25]} 
      initialState={{ pagination: { paginationModel: { pageSize: 10 } } }} 
      onRowClick={handleRowClick} 
    />
  </div>
)}
```

**Step 4: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 7: Final Build and Verification

**Step 1: Run full build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no TypeScript errors

**Step 2: Run linter**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run lint`
Expected: No new linting errors (pre-existing ones are acceptable)

---

## Summary of Changes

| Component | Change |
|-----------|--------|
| TemperatureGraph | `minTickGap`, `interval="preserveStartEnd"`, rotate labels 45° on mobile |
| WeatherGraph | Same as TemperatureGraph |
| DateContextProvider | Default to 2 days on mobile (via useEffect on mount) |
| DateRangePicker | Full-width DatePickers on mobile, stretch layout |
| SensorHealthHistory | Stack TextField/Button vertically on mobile, full-width |
| AlertsPage | Card-based list on mobile instead of DataGrid |
