---
id: monitor-energy-usage
title: How to monitor energy usage with a Zigbee smart plug
sidebar_position: 2
---

# How to monitor energy usage with a Zigbee smart plug

This guide shows you how to set up a Zigbee smart plug (such as the SONOFF S40 Lite) to monitor power consumption of an appliance and visualise it on a Sensor Hub dashboard.

## Before you start

You will need:

- A Zigbee smart plug already connected to Sensor Hub via Zigbee2MQTT (see [How to connect a Zigbee device](./connect-zigbee-device.md))
- The smart plug sensor approved and showing readings in Sensor Hub

## Step 1 — Confirm the plug is reporting power data

1. Navigate to **Sensors** in the web UI
2. Click on your smart plug sensor (e.g. `office-plug`)
3. You should see readings for **Power**, **Current**, **Voltage**, and **Energy**

If you only see link quality or battery readings, your plug may not support energy monitoring — check the [Zigbee2MQTT supported devices list](https://www.zigbee2mqtt.io/supported-devices/) for your model.

## Step 2 — Create an energy monitoring dashboard

1. Navigate to **Dashboards**
2. Click **New Dashboard** and name it (e.g. "Energy Monitoring")
3. Add the following widgets:

### Current power reading

Add a **Current Reading** widget:
- **Sensor**: Select your smart plug
- **Measurement Type**: `power`

This shows the live power draw in watts.

### Power gauge

Add a **Gauge** widget:
- **Sensor**: Select your smart plug
- **Measurement Type**: `power`
- **Min**: `0`
- **Max**: Set to the maximum expected wattage (e.g. `2000` for a kettle, `200` for a monitor)

### Power over time

Add a **Readings Chart** widget:
- **Measurement Type**: `power`
- **Time Range**: `24h` or `7d`

This shows power consumption trends over time for all sensors reporting power.

### Daily power heatmap

Add a **Heatmap** widget:
- **Sensor**: Select your smart plug
- **Measurement Type**: `power`
- **Scale Min**: `0`
- **Scale Max**: Set to your typical peak wattage

This gives a colour-coded 30-day view of when the appliance is drawing power.

### Cumulative energy

Add a **Current Reading** widget:
- **Sensor**: Select your smart plug
- **Measurement Type**: `energy`

This shows the total energy consumed in kWh since the plug was connected.

## Step 3 — Set up a power alert (optional)

You can create an alert rule to notify you when power exceeds a threshold:

1. Navigate to **Alerts**
2. Click **Add Rule**
3. Configure:
   - **Sensor**: Your smart plug
   - **Measurement Type**: `power`
   - **Condition**: Greater than
   - **Threshold**: Your desired limit (e.g. `1500` for 1.5kW)

You will receive a notification in the Sensor Hub UI (and by email if SMTP is configured) when the threshold is exceeded.

## Step 4 — Compare multiple plugs

If you have multiple smart plugs monitoring different appliances, use the **Comparison Chart** widget:

- **Measurement Type**: `power`
- **Sensors**: Select all your smart plugs

This overlays the power draw of each appliance on a single chart, making it easy to spot which appliance uses the most energy.

You can also use the **Group Summary** widget with measurement type `power` to see the total average power draw across all monitored appliances.

## Tips

- **Reporting interval**: Zigbee smart plugs typically report every 30–60 seconds, or immediately when a significant change is detected. You do not need to configure a polling interval — MQTT messages arrive automatically.
- **Energy vs power**: Power (W) is the instantaneous draw. Energy (kWh) is cumulative consumption over time. Both are useful — power for spotting peaks, energy for tracking costs.
- **Cost estimation**: If your electricity tariff is 28p/kWh, multiply the energy reading (kWh) by 0.28 to estimate cost. Sensor Hub does not currently calculate costs automatically, but the raw kWh data is available via the API for external tools.
