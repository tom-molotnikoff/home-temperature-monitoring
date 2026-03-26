# Weather Forecast Redesign Implementation Plan

> **For Copilot:** REQUIRED SUB-SKILL: Use superpowers:iterative-development to implement this plan task-by-task.

**Goal:** Replace the weather line chart with a 7-day forecast card view featuring day cards with icons, plus an hourly detail toggle for today. Make location configurable via application properties.

**Architecture:** Add three weather properties to the Go backend (`weather.latitude`, `weather.longitude`, `weather.location.name`) following the existing properties pattern. Rewrite `useWeatherApi` to request both daily and hourly Open-Meteo data. Replace `WeatherDataCard` + `WeatherGraph` with new `WeatherForecastCard` containing `DayForecastCard` components and `HourlyForecastDetail`. Create a WMO code → MUI icon mapping utility.

**Tech Stack:** Go (application properties), React 19, TypeScript, MUI v7, Open-Meteo API (daily + hourly endpoints)

---

## Task List (Outline)

1. **Backend: Add weather properties** — defaults, struct fields, map conversion, load
2. **WMO icon mapping utility** — weather code → MUI icon + label
3. **Rewrite useWeatherApi hook** — daily + hourly fields, accept configurable lat/long
4. **DayForecastCard component** — individual day card with icon, temps, rain, wind
5. **HourlyForecastDetail component** — scrollable hourly breakdown for today
6. **WeatherForecastCard component** — main container, reads properties, renders day cards + hourly toggle
7. **Dashboard layout update** — swap WeatherDataCard for WeatherForecastCard, full-width row
8. **Cleanup** — delete WeatherDataCard, WeatherGraph; remove unused imports
9. **Verify** — tsc, eslint, go build, go test, npm run build

---

### Task 1: Backend — Add weather properties

**Files:**
- Modify: `sensor_hub/application_properties/application_properties_defaults.go`
- Modify: `sensor_hub/application_properties/application_configuration.go`

**Step 1: Add defaults**

In `application_properties_defaults.go`, add to `ApplicationPropertiesDefaults`:
```go
// Weather defaults
"weather.latitude":      "53.383",
"weather.longitude":     "-1.4659",
"weather.location.name": "Sheffield",
```

**Step 2: Add struct fields**

In `application_configuration.go`, add to `ApplicationConfiguration`:
```go
// Weather configuration
WeatherLatitude     string
WeatherLongitude    string
WeatherLocationName string
```

**Step 3: Add to ConvertConfigurationToMaps**

In `ConvertConfigurationToMaps`, add to `appProps`:
```go
appProps["weather.latitude"] = cfg.WeatherLatitude
appProps["weather.longitude"] = cfg.WeatherLongitude
appProps["weather.location.name"] = cfg.WeatherLocationName
```

**Step 4: Add to LoadConfigurationFromMaps**

Add before the `smtpProps` section:
```go
if v, ok := appProps["weather.latitude"]; ok {
    cfg.WeatherLatitude = v
}
if v, ok := appProps["weather.longitude"]; ok {
    cfg.WeatherLongitude = v
}
if v, ok := appProps["weather.location.name"]; ok {
    cfg.WeatherLocationName = v
}
```

**Step 5: Add to ReloadConfig log struct**

Add `WeatherLatitude`, `WeatherLongitude`, `WeatherLocationName` fields and values to the anonymous struct in `ReloadConfig`.

**Step 6: Verify**

Run: `cd sensor_hub && go build ./... && go test ./application_properties/...`

---

### Task 2: WMO icon mapping utility

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/tools/weatherIcons.ts`

Create mapping of WMO weather codes to MUI icon components and labels. Codes:
- 0-1 → WbSunnyOutlined (Clear)
- 2-3 → CloudOutlined (Cloudy)
- 45,48 → Foggy (Fog)
- 51-57 → GrainOutlined (Drizzle)
- 61-67 → WaterDropOutlined (Rain)
- 71-77 → AcUnitOutlined (Snow)
- 80-82 → WaterDropOutlined (Showers)
- 85-86 → AcUnitOutlined (Snow showers)
- 95,96,99 → ThunderstormOutlined (Storm)

Export: `getWeatherInfo(code: number): { icon: SvgIconComponent; label: string }`

**Verify:** `npx tsc --noEmit`

---

### Task 3: Rewrite useWeatherApi hook

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/hooks/useWeatherApi.ts`

Rewrite entirely. New signature:
```typescript
export function useWeatherApi(latitude: number, longitude: number):
  { data: WeatherForecastData | null; loading: boolean; error: string | null }
```

Export types: `DailyForecast`, `HourlyForecast`, `WeatherForecastData`.

Single Open-Meteo request with both `daily` and `hourly` params, `forecast_days=7`.

**Daily fields:** `weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max`

**Hourly fields:** `temperature_2m,apparent_temperature,precipitation_probability,weather_code,wind_speed_10m`

Transform response into typed data. Filter hourly to today only (client-side date comparison).

**Verify:** `npx tsc --noEmit`

---

### Task 4: DayForecastCard component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/DayForecastCard.tsx`

Props: `{ day: DailyForecast; isToday: boolean }`

Renders a compact MUI `Paper` card:
- Day name + short date at top
- Weather icon (from `getWeatherInfo`) centred
- Weather label below icon
- High/Low temps
- Rain % with WaterDrop icon
- Wind speed with Air icon
- `isToday` → primary-coloured border

**Verify:** `npx tsc --noEmit`

---

### Task 5: HourlyForecastDetail component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/HourlyForecastDetail.tsx`

Props: `{ hours: HourlyForecast[] }`

Horizontal scrollable row of compact hour chips/cells. Each shows:
- Time (HH:00)
- Small weather icon
- Temp + feels-like
- Rain %
- Wind speed

Use MUI `Box` with `overflow-x: auto` and `flex-wrap: nowrap`.

**Verify:** `npx tsc --noEmit`

---

### Task 6: WeatherForecastCard component

**Files:**
- Create: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherForecastCard.tsx`

Reads `weather.latitude`, `weather.longitude`, `weather.location.name` from `useProperties()`. If lat/long empty → show EmptyState ("Configure weather location in Settings").

Title: "Weather — {locationName}". Renders:
1. Horizontal scrollable row of 7 × `DayForecastCard`
2. Toggle button "Show hourly detail" (only for today)
3. Conditionally renders `HourlyForecastDetail` below

Loading/error states handled.

**Verify:** `npx tsc --noEmit`

---

### Task 7: Dashboard layout update

**Files:**
- Modify: `sensor_hub/ui/sensor_hub_ui/src/pages/temperature-dashboard/TemperatureDashboard.tsx`

Replace `WeatherDataCard` import with `WeatherForecastCard`. It already sits in a `<Grid size={12}>` (full-width) — keep that. Remove `WeatherDataCard` import.

**Verify:** `npx tsc --noEmit`

---

### Task 8: Cleanup

**Files:**
- Delete: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherDataCard.tsx`
- Delete: `sensor_hub/ui/sensor_hub_ui/src/components/WeatherGraph.tsx`

Search entire `src/` for remaining imports of these files. Remove any dead references. DateContext stays (used by IndoorTemperatureDataCard).

**Verify:** `npx tsc --noEmit && npx eslint .`

---

### Task 9: Full verification

Run all checks:
```bash
cd sensor_hub && go build ./... && go test ./...
cd sensor_hub/ui/sensor_hub_ui && npx tsc --noEmit && npx eslint . && npm run build
```
