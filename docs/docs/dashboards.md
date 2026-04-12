---
id: dashboards
title: Dashboards
sidebar_position: 9.5
---

# Dashboards

Dashboards let you build custom views of your Sensor Hub data by arranging
widgets on a drag-and-drop grid. Each dashboard is saved per-user and persists
across sessions.

## Creating a dashboard

Navigate to **Dashboards** in the sidebar. If no dashboards exist, a welcome
screen appears with a **Create Dashboard** button. Enter a name and click
create — you will be taken to the new dashboard automatically.

You can create multiple dashboards from the toolbar using the **+** button.
Switch between dashboards with the dropdown in the toolbar.

## Editing a dashboard

Click the **pencil icon** in the toolbar to enter edit mode. In edit mode you
can:

- **Add widgets** — click the **Add Widget** button to open the widget picker.
  Widgets are organised by category (charts, tables, status, utility). Click one
  to add it to the dashboard.
- **Move widgets** — drag a widget by its title bar to reposition it on the
  grid.
- **Resize widgets** — drag the bottom-right corner of a widget to change its
  size. Each widget has a minimum size to ensure its content remains readable.
- **Configure widgets** — click the **gear icon** on a widget to open its
  settings dialog (only visible for widgets that have configurable options).
- **Remove widgets** — click the **✕** button on a widget to remove it.

Click the **lock icon** to exit edit mode and save your layout.

## Deleting a dashboard

Click the **delete icon** in the toolbar. A confirmation dialog will appear.
Deleting the last dashboard returns you to the welcome screen.

## Available widgets

### Charts
| Widget | Description |
|--------|-------------|
| Temperature Chart | Line chart of temperature readings over time |
| Comparison Chart | Multi-sensor temperature overlay (configure which sensors) |
| Health Timeline | Sensor health status history for a single sensor |
| Gauge | Dial gauge showing a sensor's current reading (configure measurement type, min/max) |
| Heatmap | Temperature heatmap grid for a sensor over multiple days |

### Tables & Statistics
| Widget | Description |
|--------|-------------|
| Live Readings | Real-time data grid of the latest readings |
| Reading Statistics | Data grid with per-sensor reading statistics |
| Min / Max / Average | Summary card for a sensor over a configurable time window |
| Group Summary | Table of current values for a selected group of sensors |

### Status
| Widget | Description |
|--------|-------------|
| Sensor Health Pie | Pie chart of current sensor health statuses |
| Sensor Driver Pie | Pie chart showing sensor driver distribution |
| Alert Summary | Overview of active alert rules and their status |
| Uptime | Sensor uptime percentage over a configurable number of days |
| Current Reading | Large display of a single sensor's current value |

### Utility
| Widget | Description |
|--------|-------------|
| Weather Forecast | Local weather forecast with hourly detail |
| Notifications Feed | Scrollable list of recent notifications |
| Markdown Note | Free-text note with a configurable title and content |

## Widget configuration

Some widgets require or accept configuration:

- **sensorName** — which sensor to display (e.g. Health Timeline, Current
  Reading, Gauge, Uptime)
- **sensorIds** — multiple sensors to display (e.g. Comparison Chart, Group
  Summary)
- **hours / days** — time window for historical data (e.g. Min/Max/Avg, Uptime,
  Heatmap)
- **min / max** — scale range for the Gauge widget
- **title / content** — text for the Markdown Note widget

These are set in the widget settings dialog (gear icon) while in edit mode.

## Responsive layout

Dashboards use a responsive grid that adapts to your screen size. Widgets
rearrange automatically at smaller breakpoints. The layout you save is
remembered separately for large, medium, and small screens.

## Permissions

| Role | Can view dashboards | Can create / edit / delete |
|------|--------------------|-----------------------------|
| Admin | ✓ | ✓ |
| User | ✓ | ✓ |
| Viewer | ✓ | ✗ |
