# Empty State UI Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Differentiate "loading" from "no data" states across the UI, show helpful empty states with CTAs when no sensors/alerts exist, and fix the alert rules null crash.

**Architecture:** Add a `loaded` boolean to the sensor context (set `true` on first WebSocket message, even if empty). Create a reusable `EmptyState` component (MUI icon + message + optional CTA button). Wire it into every sensor-dependent component. Fix the `alertRules` null crash in the alerts page and add an empty state there too. The weather card is independent and stays unchanged.

**Tech Stack:** React 19, TypeScript, MUI v7, Recharts, Formik, Vite

---

## Task List (Outline)

1. **Add `loaded` flag to sensor context** — `useSensors`, `SensorContextType`, `SensorContext`, `useSensorContext`
2. **Create reusable `EmptyState` component** — new file
3. **Fix `TemperatureGraph` empty state** — replace blank `<></>` with EmptyState
4. **Fix `IndoorTemperatureDataCard`** — hide controls + show EmptyState when no sensors
5. **Fix `CurrentTemperatures`** — use `loaded` flag, show EmptyState when empty
6. **Fix `SensorsDataGrid`** — use `loaded` flag, show EmptyState when empty
7. **Fix `TemperatureSensorsCard`** — pass `loaded` to child, show EmptyState
8. **Fix `AllSensorsCard`** — same pattern
9. **Fix `SensorHealthCard` / `SensorHealthPieChart`** — show EmptyState when no sensors
10. **Fix `SensorTypeCard` / `SensorTypePieChart`** — show EmptyState when no sensors
11. **Fix `TotalReadingsForEachSensorCard`** — use `loaded`, show EmptyState
12. **Fix `AlertRulesCard` null crash + empty state** — guard against null, add EmptyState CTA
13. **Verify all changes** — build, lint, visual check

---

### Task 1: Add `loaded` flag to sensor context

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/hooks/useSensors.ts`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/providers/SensorContextType.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/providers/SensorContext.tsx`

**Step 1:** Edit `useSensors.ts` — add a `loaded` state, set it `true` on first `onmessage`, return `{ sensors, loaded }`.

- Add `const [loaded, setLoaded] = useState(false);`
- Inside `ws.onmessage`, after the existing logic, call `setLoaded(true)` (React deduplicates identical state updates).
- Reset `loaded` to `false` at top of `useEffect` before ws setup so a reconnect correctly re-transitions.
- Return `{ sensors, loaded }` instead of just `sensors`.

**Step 2:** Edit `SensorContextType.tsx` — add `loaded: boolean` to context type.

```ts
type SensorContextValueType = {
  sensors: Sensor[];
  loaded: boolean;
};

export const SensorContext = createContext<SensorContextValueType>({ sensors: [], loaded: false });
```

**Step 3:** Edit `SensorContext.tsx` (the provider) — destructure `loaded` from `useSensors` and pass it.

```tsx
export function SensorContextProvider({ children, type }: SensorContextProviderProps) {
  const { sensors, loaded } = useSensors({ type });

  return (
    <SensorContext.Provider value={{ sensors, loaded }}>
      {children}
    </SensorContext.Provider>
  );
}
```

**Step 4:** No changes needed to `useSensorContext.ts` — it already returns the full context value. Consumers just destructure `loaded` alongside `sensors`.

**Step 5:** Verify build compiles: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

---

### Task 2: Create reusable `EmptyState` component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/EmptyState.tsx`

**Step 1:** Create the component with these props:

```tsx
interface EmptyStateProps {
  icon: React.ReactNode;          // MUI icon element, e.g. <SensorsOffOutlined sx={{ fontSize: 48 }} />
  title: string;                   // Primary message
  description?: string;            // Secondary guidance text
  actionLabel?: string;            // CTA button text
  actionHref?: string;             // Navigate to this route on click
  onAction?: () => void;           // Or call this function on click
  minHeight?: number | string;     // Minimum height of the container (default: 200)
}
```

Layout: centered flex column, muted icon colour (`text.disabled`), title in `body1` bold, description in `body2` `text.secondary`, button is MUI `variant="outlined"`.

**Step 2:** Verify it compiles: `npx tsc --noEmit`

---

### Task 3: Fix `TemperatureGraph` empty state

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/TemperatureGraph.tsx`

**Step 1:** Replace the empty fragment `<></>` with two distinct empty states:

- `sensors.length === 0` → EmptyState with "No sensors configured" + CTA to add sensor
- `sensors.length > 0 && chartData empty` → EmptyState with "No readings in selected date range"
- Otherwise → existing chart

Use `ShowChartOutlinedIcon`. Pass `minHeight={compact ? 250 : 350}` to match chart height.

**Step 2:** Verify build.

---

### Task 4: Fix `IndoorTemperatureDataCard`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/IndoorTemperatureDataCard.tsx`

**Step 1:** Use `loaded` from `useSensorContext`. When `sensors.length === 0`, hide the `DateRangePicker` and `HourlyAveragesToggle` — they're useless without sensors. The `TemperatureGraph` handles its own empty state (Task 3).

```tsx
const { sensors, loaded } = useSensorContext();

// Conditionally render controls:
{sensors.length > 0 && (
  <>
    <DateRangePicker />
    <HourlyAveragesToggle ... />
  </>
)}
```

**Step 2:** Verify build.

---

### Task 5: Fix `CurrentTemperatures`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/CurrentTemperatures.tsx`

**Step 1:** Import `useSensorContext` and `EmptyState`. Replace the local `isLoading` logic.

Current bug: `isLoading` starts `true` and only becomes `false` when `Object.keys(currentTemperatures).length > 0` — with zero sensors this never fires.

Fix: Remove the local `isLoading` state and its `useEffect`. Derive loading from context: `const isLoading = !loaded;`.

When `loaded && rows.length === 0`, show EmptyState (with `ThermostatOutlinedIcon`) instead of the DataGrid. Otherwise show the DataGrid with `loading={isLoading}`.

**Step 2:** Verify build.

---

### Task 6: Fix `SensorsDataGrid`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorsDataGrid.tsx`

**Step 1:** Import `useSensorContext` and `EmptyState`. Replace the local loading logic (same bug as CurrentTemperatures).

Remove the local `isLoading` state and its `useEffect`. Derive: `const isLoading = !loaded;`.

When `loaded && rows.length === 0`, show EmptyState instead of DataGrid. Message varies by permission:
- Has `manage_sensors`: "Use the Add Sensor form to register your first sensor." + CTA
- Lacks permission: "No sensors have been added yet. Ask an administrator to add sensors."

Use `SensorsOffOutlinedIcon`.

**Step 2:** Verify build.

---

### Task 7: Fix `TemperatureSensorsCard`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/TemperatureSensorsCard.tsx`

**Step 1:** No code changes needed here — it already passes sensors to `SensorsDataGrid`, which now handles empty state (Task 6). The `loaded` flag is consumed from context inside `SensorsDataGrid`.

Just verify it works correctly via build.

---

### Task 8: Fix `AllSensorsCard`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/AllSensorsCard.tsx`

**Step 1:** Same as Task 7 — no changes needed. `SensorsDataGrid` handles it.

Verify via build.

---

### Task 9: Fix `SensorHealthCard` + `SensorHealthPieChart`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorHealthCard.tsx`

**Step 1:** Use `loaded` from `useSensorContext`. When `loaded && sensors.length === 0`, show EmptyState with `MonitorHeartOutlined` icon and "No sensor health data" message instead of the pie chart.

When `!loaded`, show a `CircularProgress` spinner.

**Step 2:** Verify build.

---

### Task 10: Fix `SensorTypeCard` + `SensorTypePieChart`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/SensorTypeCard.tsx`

**Step 1:** Same pattern as Task 9. Use `loaded` from context. When empty, show EmptyState with `CategoryOutlined` icon and "No sensors to categorise".

**Step 2:** Verify build.

---

### Task 11: Fix `TotalReadingsForEachSensorCard`

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/TotalReadingsForEachSensorCard.tsx`

**Step 1:** Import `useSensorContext` and `EmptyState`. Use `loaded` from sensor context.

Current bug: same `isLoading` pattern — never becomes `false` with zero data.

Fix: Derive `isLoading` from `loaded`: `const isLoading = !loaded;`. Remove local `isLoading` state and `useEffect`.

When `loaded && rows.length === 0`, show EmptyState with `BarChartOutlined` icon and "No reading data yet".

**Step 2:** Verify build.

---

### Task 12: Fix `AlertRulesCard` null crash + empty state

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/AlertRulesCard.tsx`
- Modify: `sensor_hub/ui/sensor_hub_ui/src/components/AlertRuleDataGrid.tsx`

**Step 1:** Fix the crash in `AlertRulesCard.tsx` — the `listAlertRules()` API can return `null` when no rules exist (Go nil slice serialises as JSON `null`).

In the `load` function, guard against null:
```ts
const rules = await listAlertRules();
setAlertRules(rules ?? []);
```

**Step 2:** In `AlertRuleDataGrid.tsx`, add a defensive guard: `const rows = (alertRules ?? []).map(...)`.

**Step 3:** Add a loading state to `AlertRulesCard`. Track `loaded` with a local `useState` (alerts aren't from sensor context). Set it `true` after the first fetch completes (success or error). When `!loaded`, show the DataGrid area with a loading indicator.

**Step 4:** When `loaded && alertRules.length === 0` on desktop (non-mobile), show an EmptyState with `NotificationsNoneOutlined` icon, "No alert rules configured", and a CTA button "Create Alert Rule" that opens the create dialog. This replaces only the DataGrid area, keeping the header/button visible.

The mobile view already handles this (line 66 of AlertRulesCard: "No alert rules configured" Typography) — just enhance it to use the EmptyState component for consistency.

**Step 5:** Verify build.

---

### Task 13: Verify all changes

**Step 1:** Run full TypeScript check: `cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit`

**Step 2:** Run ESLint: `cd sensor_hub/ui/sensor_hub_ui && npx eslint src/`

**Step 3:** Run existing tests: `cd sensor_hub/ui/sensor_hub_ui && npx vitest run` (or whatever test runner is configured)

**Step 4:** Run Vite build: `cd sensor_hub/ui/sensor_hub_ui && npm run build`

**Step 5:** Fix any issues found.

---

