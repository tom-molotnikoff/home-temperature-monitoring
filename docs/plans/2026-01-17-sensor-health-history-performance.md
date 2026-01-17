# SensorHealthHistory Performance & Layout Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Fix chart performance lag by rendering dots only at state transitions, and improve controls layout by moving them to a header row.

**Architecture:** Add `isTransition` flag during data mapping to identify status changes. Use Recharts' `dot` prop as a render function that conditionally renders dots. Restructure layout with responsive header row containing title and controls.

**Tech Stack:** React, Recharts, MUI, existing `useIsMobile()` hook

---

## Task 1: Add Transition Detection to Data Mapping

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx`

**Step 1: Update useMemo to calculate isTransition**

In the `mappedData` useMemo (around line 34), update the mapping to track transitions:

```tsx
return sortedByRecordedAt.map((h: SensorHealthHistory, index: number) => {
  const recorded = h.recordedAt;
  const status = h.healthStatus;
  const value = mapStatusToValue(status);
  const prevValue = index > 0 ? mapStatusToValue(sortedByRecordedAt[index - 1].healthStatus) : null;
  const isTransition = prevValue === null || prevValue !== value;
  return {
    ...h,
    recordedAt: recorded,
    healthStatus: status,
    healthValue: value,
    isTransition,
    // per-state series used to draw colored segments (null when not active so Recharts doesn't connect)
    goodVal: value === 2 ? 2 : null,
    badVal: value === 1 ? 1 : null,
    unknownVal: value === 0 ? 0 : null,
  };
});
```

Key change: Added `index` parameter, calculate `prevValue`, set `isTransition` flag.

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 2: Create Custom Dot Renderer

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx`

**Step 1: Add TransitionDot component before SensorHealthHistoryChart function**

Add this component after the imports (around line 19):

```tsx
// Custom dot that only renders at transition points
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function TransitionDot(props: any) {
  const { cx, cy, payload, stroke } = props;
  if (!payload?.isTransition) return null;
  return <circle cx={cx} cy={cy} r={4} fill={stroke} stroke={stroke} />;
}
```

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 3: Update Line Components to Use Custom Dot

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx`

**Step 1: Replace dot={true} with dot={TransitionDot}**

Find the three Line components (around lines 134-136) and update them:

```tsx
<Line type="step" dataKey="goodVal" stroke={lineColours[0]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Good" />
<Line type="step" dataKey="badVal" stroke={lineColours[1]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Bad" />
<Line type="step" dataKey="unknownVal" stroke={lineColours[2]} dot={TransitionDot} strokeWidth={4} isAnimationActive={false} name="Unknown" />
```

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 4: Create Header Layout with Controls

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx`

**Step 1: Add IconButton import**

Update the MUI imports (line 16):

```tsx
import {Alert, Button, IconButton, Snackbar, TextField} from "@mui/material";
```

**Step 2: Add header row before ResponsiveContainer**

Replace the current structure. After the empty check `{!Array.isArray(mappedData) || mappedData.length === 0 ? (<></>) : (` add a header row:

```tsx
<>
  {/* Header row: title left, controls right (desktop) or stacked (mobile) */}
  <div style={{
    display: "flex",
    flexDirection: isMobile ? "column" : "row",
    justifyContent: "space-between",
    alignItems: isMobile ? "stretch" : "center",
    width: "100%",
    marginBottom: 16,
    gap: isMobile ? 12 : 16,
  }}>
    <h3 style={{ margin: 0, fontSize: isMobile ? 18 : 20 }}>Sensor Health History</h3>
    <div style={{
      display: "flex",
      flexDirection: isMobile ? "column" : "row",
      alignItems: isMobile ? "stretch" : "center",
      gap: 8,
    }}>
      <TextField
        label="Limit"
        type="number"
        size="small"
        value={limitInput}
        onChange={(e) => setLimitInput(e.target.value)}
        sx={{ width: isMobile ? "100%" : 120 }}
        fullWidth={isMobile}
      />
      {isMobile ? (
        <Button
          onClick={() => {
            const parsed = parseInt(limitInput);
            const isNegative = Number.isFinite(parsed) && parsed < 0;
            if (isNegative) {
              setLimitInput("5000");
            }
            setLimit(Number.isFinite(parsed) ? parsed : 5000);
            refresh().then(() => {
              setSnackbarOpen(true);
            });
          }}
          variant="outlined"
          startIcon={<RefreshIcon />}
          fullWidth
        >
          Refresh
        </Button>
      ) : (
        <IconButton
          onClick={() => {
            const parsed = parseInt(limitInput);
            const isNegative = Number.isFinite(parsed) && parsed < 0;
            if (isNegative) {
              setLimitInput("5000");
            }
            setLimit(Number.isFinite(parsed) ? parsed : 5000);
            refresh().then(() => {
              setSnackbarOpen(true);
            });
          }}
          color="primary"
          size="small"
          title="Refresh"
        >
          <RefreshIcon />
        </IconButton>
      )}
    </div>
  </div>

  <ResponsiveContainer width="100%" height={400}>
```

**Step 3: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 5: Remove Old Controls Section

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx`

**Step 1: Delete the old controls div**

Remove the entire div containing TextField and Button after ResponsiveContainer (the div starting around line 142 with the old controls).

The section to remove starts with:
```tsx
<div style={{
  display: "flex",
  flexDirection: isMobile ? "column" : "row",
  justifyContent: isMobile ? "center" : "flex-end",
```

And ends with the closing `</div>` after the Button.

**Step 2: Verify build succeeds**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds

---

## Task 6: Run Linter and Final Build

**Step 1: Run linter**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run lint`
Expected: No new linting errors (pre-existing ones acceptable)

**Step 2: Run final build**

Run: `cd sensor_hub/ui/sensor_hub_ui && npm run build`
Expected: Build succeeds with no TypeScript errors

**Step 3: Commit**

```bash
git add sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthHistoryChart.tsx
git commit -m "perf: render dots only at transitions, move controls to header"
```

---

## Summary of Changes

| Change | Description |
|--------|-------------|
| Transition detection | `isTransition` flag computed during data mapping |
| Custom dot renderer | `TransitionDot` component renders only at status changes |
| Header controls | Title + compact controls in responsive header row |
| Mobile layout | Full-width stacked controls on mobile preserved |
| Performance | Reduces DOM elements from thousands to ~dozens of dots |

## React Best Practices Applied

- `rerender-memo`: The `mappedData` is already in useMemo, transition calculation is cheap O(n)
- `rendering-conditional-render`: Using ternary for mobile/desktop layout
- `js-early-exit`: TransitionDot returns null early when not a transition
