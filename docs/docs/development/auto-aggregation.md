---
id: auto-aggregation
title: Auto-Aggregation
sidebar_position: 12
---

# Auto-Aggregation

Smart auto-aggregation automatically selects an appropriate time bucket size when querying sensor readings, keeping response sizes manageable for charting regardless of the time range requested.

## Problem

Without aggregation, a 30-day query at a 5-minute collection interval produces ~8,640 data points per sensor. With multiple sensors this quickly exceeds what chart libraries like Recharts can render performantly, and the density of points exceeds what a user can meaningfully interpret on screen.

The previous approach used pre-computed `hourly_averages` and `hourly_events` tables maintained by a background periodic task. This had several drawbacks:

- Fixed 1-hour granularity — too coarse for short ranges, appropriate only for multi-day views
- Background task lag — averages weren't available until the next hourly run
- Storage overhead — duplicate data in separate tables
- Code complexity — separate API endpoint, separate repository methods, separate UI toggle

## Solution

Query-time aggregation using SQL `GROUP BY` with `strftime()` bucket expressions. The server inspects the time span of each request and selects an aggregation interval from a configurable tier list. No pre-computed tables are needed.

## Architecture

The implementation follows the existing API → Service → Repository pattern:

```
┌─────────┐     ┌──────────────────┐     ┌──────────────────────┐
│  API    │────▶│  ReadingsService │────▶│  ReadingsRepository  │
│ Handler │     │                  │     │                      │
└─────────┘     └──────────────────┘     └──────────────────────┘
     │                   │                         │
  Reads query       Resolves tier              Executes SQL
  params            + function              GROUP BY strftime()
```

### API layer (`api/readings_api.go`)

The unified `GET /readings/between` handler accepts optional `aggregation` and `aggregation_function` query parameters alongside the existing `start`, `end`, `type`, and `sensor` parameters. It delegates to the service layer and returns an `AggregatedReadingsResponse` wrapper.

### Service layer (`service/readings_service.go`)

The service is constructed with:
- A `ReadingsRepository` for data access
- A `MeasurementTypeRepository` for looking up default aggregation functions
- A slice of `AggregationTier` structs parsed from `application.properties`
- An `enabled` flag

When a request arrives, the service:

1. Calculates the time span between `start` and `end`
2. Walks the tier list (sorted by ascending threshold) to find the first tier whose threshold exceeds the span
3. If no tier matches, uses the fallback interval
4. Looks up the default aggregation function for each measurement type in the request (e.g., `avg` for temperature, `last` for binary states)
5. Passes the resolved interval and function to the repository

If `aggregation=raw` is passed or the resolved tier is `raw`, the service calls the repository's raw path — no GROUP BY is applied.

### Repository layer (`db/readings_repository.go`)

The repository has three SQL paths:

- **Raw** — `SELECT` from `readings` with no aggregation (used for short ranges or explicit `raw` override)
- **Aggregated** — `GROUP BY` using `strftime()` bucket expressions with aggregate functions (`AVG`, `MIN`, `MAX`, `SUM`, `COUNT`)
- **Last** — Uses `ROW_NUMBER() OVER (PARTITION BY ... ORDER BY r.time DESC)` to pick the last reading per bucket (used for binary/text sensors)

### Bucket expressions (`timeBucketExpression`)

The repository maps each `AggregationInterval` to a SQLite expression:

| Interval | SQL expression |
|----------|---------------|
| PT10S    | `strftime('%Y-%m-%d %H:%M:') \|\| printf('%02d', (CAST(strftime('%S', r.time) AS INTEGER) / 10) * 10)` |
| PT1M     | `strftime('%Y-%m-%d %H:%M:00', r.time)` |
| PT5M     | `strftime('%Y-%m-%d %H:') \|\| printf('%02d', (CAST(strftime('%M', r.time) AS INTEGER) / 5) * 5) \|\| ':00'` |
| PT15M    | Same pattern with `/15)*15` |
| PT30M    | Same pattern with `/30)*30` |
| PT1H     | `strftime('%Y-%m-%d %H:00:00', r.time)` |
| PT6H     | Hour-based division by 6 |
| PT12H    | Hour-based division by 12 |
| P1D      | `strftime('%Y-%m-%d 00:00:00', r.time)` |

## Configuration

Tiers are defined in `application.properties` using ISO 8601 durations:

```properties
readings.aggregation.enabled=true
readings.aggregation.tier.1=PT15M=raw
readings.aggregation.tier.2=PT1H=PT10S
readings.aggregation.tier.3=PT6H=PT1M
readings.aggregation.tier.4=P1D=PT5M
readings.aggregation.tier.5=P7D=PT15M
readings.aggregation.tier.6=P30D=PT1H
readings.aggregation.tier.fallback=P1D
```

Each tier is `THRESHOLD=INTERVAL`, meaning: "If the time span is ≤ THRESHOLD, use INTERVAL as the bucket size." The tiers are evaluated in order of ascending threshold. The fallback interval is used when the span exceeds all thresholds.

The parsing logic lives in `application_properties/aggregation_tiers.go` and converts ISO 8601 durations to Go `time.Duration` values.

### Default aggregation functions

Each measurement type can have a default aggregation function stored in the `measurement_type_aggregations` table. Migration 000016 creates this table and seeds common defaults:

| Measurement type | Default function | Supported functions |
|-----------------|-----------------|---------------------|
| temperature     | avg             | avg, count, last    |
| humidity        | avg             | avg, count, last    |
| power           | avg             | avg, count, last    |
| door            | last            | count, last         |
| motion          | last            | count, last         |

The `last` function is used for binary/text state sensors where averaging makes no sense — instead, the last reading in each time bucket is returned.

The `supported_aggregation_functions` list for each measurement type is exposed via the `GET /measurement-types` endpoint. The UI uses this to populate the aggregation function dropdown dynamically, showing only functions that the selected measurement type supports.

### Aggregation function validation

When a client requests an `aggregation_function` override via the query parameter, the service validates that the requested function is in the measurement type's `supported_aggregation_functions` list. If not, the API returns a **400 Bad Request** with a message like:

```
aggregation function "avg" is not supported for measurement type "motion"; supported functions: count, last
```

This prevents silent fallthrough to the default function and ensures clients get meaningful feedback when requesting an unsupported aggregation.

## Response format

All readings queries return an `AggregatedReadingsResponse`:

```json
{
  "aggregation_interval": "PT5M",
  "aggregation_function": "avg",
  "readings": [ ... ]
}
```

When no aggregation is applied (short range or explicit `raw`), the response uses `"aggregation_interval": "raw"` and `"aggregation_function": "none"`.

The UI uses the `aggregation_interval` and `aggregation_function` fields to display a badge informing the user about the aggregation applied to the data they're viewing.

## Types

Core types are defined in `types/aggregation.go`:

- `AggregationInterval` — String type with constants (`AggregationRaw`, `AggregationPT10S`, ..., `AggregationP1D`)
- `AggregationFunction` — String type with constants (`AggFuncAvg`, `AggFuncCount`, `AggFuncLast`, `AggFuncNone`)
- `AggregationTier` — Struct with `Threshold` (duration) and `Interval` fields
- `AggregatedReadingsResponse` — Wrapper with `AggregationInterval`, `AggregationFunction`, and `Readings` fields
- `MeasurementTypeAggregation` — Database model linking measurement types to their default function and supported functions list
- `ErrUnsupportedAggregationFunction` — Typed error returned when a client requests an aggregation function not supported by the measurement type

## Testing

### Unit tests

- `application_properties/aggregation_tiers_test.go` — Tier parsing, duration conversion, resolution logic
- `service/readings_service_test.go` — 7 tests covering raw/aggregated/override/disabled scenarios
- `api/readings_api_test.go` — HTTP handler tests for the unified endpoint

### Integration tests

Run with the `integration` build tag:

```bash
cd sensor_hub && go test -tags integration -v -timeout 300s ./integration/
```

Integration tests in `integration/readings_test.go` verify end-to-end aggregation behavior using testcontainers with real SQLite databases.
