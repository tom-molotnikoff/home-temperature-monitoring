---
id: dashboards
title: Dashboards
sidebar_position: 9.5
---

# Dashboards

Dashboards let you build custom views of your Sensor Hub data by arranging
widgets on a drag-and-drop grid. Each dashboard is saved per-user and persists
across sessions.

## Widget configuration

Some widgets require or accept configuration:

- **sensorId** — which sensor to display (e.g. Health Timeline, Current
  Reading, Gauge, Uptime, Sensor Detail, Heatmap, Min/Max/Avg, Sensor Toggle)
- **sensorIds** — multiple sensors to display (e.g. Comparison Chart)
- **property** — which controllable binary property to switch (used by Sensor
  Toggle, defaults to `state` and is chosen from the selected sensor's
  available binary capabilities)
- **measurementType** — which measurement type to chart or display
  (e.g. `temperature`, `humidity`, `contact`, `power`)
- **timeRange** — time window for historical data: `1h`, `6h`, `24h`,
  `3d`, `7d`, `30d`, or `custom` with `customStart`/`customEnd` ISO dates
- **min / max** — scale range for the Gauge widget
- **scaleMin / scaleMax** — colour scale for the Heatmap widget
- **content** — markdown text for the Markdown Note widget

These are set in the widget settings dialog (gear icon) while in edit mode.

## Responsive layout

Dashboards use a responsive grid that adapts to your screen size. Widgets
rearrange automatically at smaller breakpoints. The layout you save is
remembered separately for large, medium, and small screens.
