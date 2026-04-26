# Ubiquitous Language

Canonical terminology for the home-temperature-monitoring project.
Use these terms consistently in code, docs, issues, and conversation.

---

## Sensors

| Term | Definition | Aliases to avoid |
|------|------------|-----------------|
| **Sensor** | A named, managed entity in the system that produces readings | device *(reserve for the physical hardware, not the system entity)* |
| **Sensor Driver** | A named plugin that knows how to communicate with a particular class of sensor | adapter, integration, connector |
| **Pull-based Sensor** | A sensor whose readings are collected by the hub on a schedule by calling its driver | poll-based sensor, polling sensor |
| **Push-based Sensor** | A sensor that publishes readings to an MQTT broker; the hub receives them passively | passive sensor |
| **External ID** | The immutable hardware identifier for a push-based sensor, set at auto-discovery and unchanged even if the sensor is renamed | device ID, hardware ID |
| **Sensor Status** | The lifecycle state of a sensor | state, lifecycle |
| **Health Status** | A sensor's current operational health | status *(too generic)* |
| **Health Reason** | A short human-readable explanation of the current health state | health message, error reason |
| **Health Check** | A periodic evaluation that updates a sensor's health status and records a history entry | health probe, liveness check |
| **Retention** | The period for which a sensor's readings are kept; configurable globally or overridden per sensor | TTL, expiry |

## Readings

| Term | Definition | Aliases to avoid |
|------|------------|-----------------|
| **Reading** | A single measurement produced by a sensor at a specific point in time | data point, sample, measurement |
| **Measurement Type** | The physical quantity being measured (e.g. temperature, humidity) | metric, category, sensor type |
| **Collection** | The act of triggering the hub to fetch new readings from one or all pull-based sensors | trigger, fetch, poll |
| **Aggregated Readings** | Readings summarised into time buckets when the queried time span is large | rolled-up readings, bucketed data |
| **Aggregation Function** | The statistic applied within each time bucket when aggregating | reducer, statistic |

## MQTT

| Term | Definition | Aliases to avoid |
|------|------------|-----------------|
| **Broker** | An MQTT message broker the hub connects to | server, MQTT server |
| **Subscription** | A topic pattern on a broker bound to a driver type; determines which messages produce readings | topic rule, binding |
| **Topic Pattern** | An MQTT topic glob that a subscription matches against incoming messages | topic filter, topic mask |
| **Auto-discovery** | The process by which a push-based device's first message creates a new sensor awaiting approval | device discovery, auto-registration |

## Dashboards

| Term | Definition | Aliases to avoid |
|------|------------|-----------------|
| **Dashboard** | A named, user-owned grid of widgets | layout, board |
| **Widget** | A single visualisation panel on a dashboard | card, panel, tile |
| **Default Dashboard** | The dashboard shown automatically when a user logs in | home dashboard, landing dashboard |
| **Shared Dashboard** | A dashboard granted read-only access to another user | public dashboard |

## Users & Access Control

| Term | Definition | Aliases to avoid |
|------|------------|-----------------|
| **User** | An authenticated identity in the system | account, login |
| **Role** | A named group of permissions assigned to users | group, profile |
| **Permission** | A named capability that controls access to a specific operation | right, privilege |

---

## Relationships

- A **Sensor** is always associated with exactly one **Sensor Driver**.
- A **Pull-based Sensor** produces **Readings** via **Collection**; a **Push-based Sensor** produces **Readings** via an MQTT **Subscription**.
- A **Push-based Sensor** gains an **External ID** at **Auto-discovery** and enters a pending **Sensor Status** until approved.
- **Aggregated Readings** are derived from **Readings** at query time; no pre-computed aggregate data is stored.
- A **Dashboard** is owned by one **User** and contains zero or more **Widgets**.
- A **Permission** belongs to one or more **Roles**; a **User** may have multiple **Roles**.

---

## Example dialogue

> **Dev:** "A new Zigbee device sent its first message. What happens?"
>
> **Domain expert:** "**Auto-discovery** fires — the hub creates a new **Sensor** with a pending **Sensor Status** and assigns it an **External ID** from the device's hardware identifier. It won't produce **Readings** until approved."
>
> **Dev:** "Which **Sensor Driver** handles it?"
>
> **Domain expert:** "Whatever driver the matching **Subscription** declares. The **Subscription** maps a **Topic Pattern** to a **Sensor Driver**, so all messages on that topic are parsed by that driver."
>
> **Dev:** "If the sensor goes quiet after approval, does its **Health Status** change?"
>
> **Domain expert:** "Yes — the next **Health Check** will record a new **Health Reason** explaining why. The full history of those checks is kept against the sensor."
>
> **Dev:** "And if someone queries a year's worth of **Readings**, do we return them all?"
>
> **Domain expert:** "No — the response will be **Aggregated Readings** with an **Aggregation Function** selected automatically based on the **Measurement Type** and span. The caller can override it or request raw readings."

---

## Flagged ambiguities

- **"device" vs "sensor"** — used interchangeably in some places. Prefer **Sensor** for the system entity; reserve **device** only when referring to physical hardware independently of its system representation (e.g. during **Auto-discovery**).
- **"config"** — overloaded across the codebase. Qualify it: sensor driver configuration is **Sensor Config**; dashboard-level settings are a **Dashboard**'s configuration; application-level settings are **Properties**.
- **"status"** — overloaded. **Sensor Status** is the lifecycle state; **Health Status** is the operational health. Never use bare "status" for either.
