# Mobile Responsiveness Implementation Guide

This document describes the mobile responsiveness patterns used in the sensor-hub UI and how to apply them when building new components or pages.

## Overview

The UI is fully responsive with a mobile breakpoint of **950px**. Below this width, layouts switch to mobile-optimized configurations with stacked layouts, simplified charts, and touch-friendly interactions.

## Core Infrastructure

### useIsMobile Hook

Location: `src/hooks/useMobile.ts`

```typescript
import { useIsMobile } from '../hooks/useMobile';

function MyComponent() {
  const isMobile = useIsMobile();
  // isMobile is true when window width < 950px
}
```

**Important:** The hook may return `undefined` during SSR/initial render. Always handle this case when used in initial state.

### DesktopRowMobileColumn Component

Location: `src/tools/DesktopRowMobileColumn.tsx`

A flex container that displays children in a row on desktop and a column on mobile:

```tsx
<DesktopRowMobileColumn>
  <DatePicker />
  <DatePicker />
</DesktopRowMobileColumn>
```

## Common Patterns

### 1. Grid Layouts

Use MUI Grid with conditional sizing:

```tsx
import { Grid } from '@mui/material';

const isMobile = useIsMobile();

<Grid container spacing={2}>
  <Grid size={isMobile ? 12 : 6}>
    {/* Full width on mobile, half on desktop */}
  </Grid>
  <Grid size={isMobile ? 12 : 6}>
    {/* Stacks below on mobile */}
  </Grid>
</Grid>
```

### 2. Recharts Configuration

For charts using Recharts, apply these patterns:

```tsx
const isMobile = useIsMobile();

<LineChart>
  <XAxis
    dataKey="timestamp"
    interval="preserveStartEnd"  // Always show first/last labels
    minTickGap={isMobile ? 30 : 50}  // Reduce cramping
    angle={isMobile ? -45 : 0}  // Rotate labels on mobile
    textAnchor={isMobile ? 'end' : 'middle'}  // Align rotated labels
    height={isMobile ? 60 : 30}  // Extra height for rotated labels
  />
</LineChart>
```

**Key XAxis props:**
- `interval="preserveStartEnd"` - Auto-skip middle ticks, always show first/last
- `minTickGap` - Minimum pixels between tick labels
- `angle` - Rotation angle (use -45 for mobile)
- `textAnchor` - Must be `"end"` when angle is set
- `height` - Increase to 60px on mobile to accommodate rotated labels

**Note:** `angle` and `textAnchor` must be props directly on `XAxis`, NOT inside the `tick` prop object.

### 3. Compact Mode for Graphs

Components like `TemperatureGraph` and `SensorHealthHistoryChart` support a `compact` prop:

```tsx
<TemperatureGraph
  sensorId={id}
  compact={isMobile}  // Reduces height and enables mobile-optimized XAxis
/>
```

### 4. DataGrid Column Visibility

Hide non-essential columns on mobile:

```tsx
const columns: GridColDef[] = [
  { field: 'name', headerName: 'Name', flex: 1 },
  { field: 'status', headerName: 'Status', width: 100 },
  // Hide on mobile
  { 
    field: 'createdAt', 
    headerName: 'Created', 
    width: 150,
    // Column will be hidden via columnVisibilityModel
  },
];

<DataGrid
  columns={columns}
  columnVisibilityModel={{
    createdAt: !isMobile,
    updatedAt: !isMobile,
  }}
/>
```

### 5. Replacing DataGrid with Cards

For complex DataGrids that don't work well on mobile (e.g., Alerts page), replace with a card-based list:

```tsx
const isMobile = useIsMobile();

return isMobile ? (
  <Box>
    {items.map((item) => (
      <Card key={item.id} onClick={() => handleClick(item)}>
        <CardContent>
          <Typography variant="h6">{item.name}</Typography>
          <Chip label={item.status} />
        </CardContent>
      </Card>
    ))}
  </Box>
) : (
  <DataGrid rows={items} columns={columns} />
);
```

### 6. Button Groups

Stack buttons vertically on mobile:

```tsx
<Box sx={{ 
  display: 'flex', 
  flexDirection: isMobile ? 'column' : 'row',
  gap: 2,
  width: isMobile ? '100%' : 'auto',
}}>
  <Button fullWidth={isMobile}>Action 1</Button>
  <Button fullWidth={isMobile}>Action 2</Button>
</Box>
```

### 7. Date Range Pickers

The `DateRangePicker` component handles mobile automatically:
- Full-width DatePickers on mobile
- Uses `DesktopRowMobileColumn` for layout
- Passes `mobileChanges` to child DatePickers

### 8. Default Date Ranges

`DateContextProvider` defaults to 2 days on mobile (vs 7 days on desktop) to reduce graph data density:

```tsx
// In DateContextProvider.tsx
useEffect(() => {
  if (isMobile) {
    setStartDate(subDays(new Date(), 2));
  }
  // eslint-disable-next-line react-hooks/exhaustive-deps
}, []); // Only run on mount
```

## Page-Specific Implementations

### Temperature Dashboard

- Shows simplified graphs with `compact` prop on mobile
- Uses `DesktopRowMobileColumn` for date pickers
- Grid items stack vertically

### Sensors Overview

- Grid items switch from 2 columns to full width
- All cards stack vertically on mobile

### Sensor Page

- Health history chart stacks controls vertically
- Temperature graph uses compact mode
- Grid layout responsive

### Notifications Page

- **Desktop:** Full-size action buttons (Refresh, Mark All Read, Dismiss All)
- **Mobile:** Smaller buttons with `size="small"` and `flexWrap` to handle overflow
- Buttons wrap to next line if needed on narrow screens

### NotificationBell Popover

- **Desktop:** Fixed 360px width, anchored to right of bell icon
- **Mobile:** 90vw width (max 360px), centered anchor position
- Prevents popover from shifting the bell icon or clipping off-screen
- Uses responsive `anchorOrigin` and `transformOrigin` settings

### Alerts Page

- **Desktop:** DataGrid with all columns
- **Mobile:** Card-based list with context menu on tap
- Each card shows sensor name, status chip, and thresholds

### Sessions Page

- **Desktop:** Shows Username, Last Active, IP Address, User Agent, Actions
- **Mobile:** Shows Device (parsed from User-Agent), Last Active, Actions
- Uses `getShortDeviceInfo()` helper to parse User-Agent into readable device name

### Users Page

- Hides Created At, Updated At, and Last Login columns on mobile
- Shows only essential: Username, Role, Permissions, Actions

### OAuth Page

- Buttons stack vertically with full width
- Status cards stack vertically

## Testing Mobile Responsiveness

1. Open browser DevTools (F12)
2. Toggle device toolbar (Ctrl+Shift+M in Chrome)
3. Set width to < 950px (e.g., iPhone 12 Pro: 390px)
4. Navigate through all pages and verify:
   - Layouts stack correctly
   - No horizontal overflow/scrolling
   - Charts are readable
   - Buttons/inputs are tappable
   - DataGrids show appropriate columns

## Adding New Pages

When creating a new page:

1. Import `useIsMobile` hook
2. Use responsive Grid sizes (`isMobile ? 12 : 6`)
3. Hide non-essential DataGrid columns
4. Consider card-based alternatives for complex tables
5. Use `DesktopRowMobileColumn` for form inputs in a row
6. Add `compact` prop to any Recharts components
7. Stack buttons vertically with full width on mobile

## Common Gotchas

1. **Container heights:** Don't use fixed heights that assume desktop controls layout. Let containers grow naturally or calculate height including stacked controls.

2. **Recharts XAxis:** `angle` and `textAnchor` go directly on XAxis, NOT in the `tick` prop.

3. **Initial hook value:** `useIsMobile()` returns `undefined` during SSR. Use `useEffect` for initial state that depends on mobile detection.

4. **Lint warnings:** Pre-existing lint warnings exist in `WeatherGraph.tsx` (conditional hook call) and some hooks. These are not related to mobile responsiveness.
