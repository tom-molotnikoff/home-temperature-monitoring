---
id: configuration-package
title: Configuration Package
sidebar_position: 7
---

# Configuration Package

The `application_properties` package manages all runtime configuration for
Sensor Hub. It loads property files from disk, provides typed access via a
struct, and watches for external changes.

## Architecture

The package is built around a single source of truth: the
`ApplicationConfiguration` struct in `application_configuration.go`. Every
configurable setting is a field on this struct, annotated with struct tags that
define defaults, file mapping, validation, and sensitivity masking.

```
┌──────────────────────────────────┐
│  ApplicationConfiguration struct │  ← single source of truth
│  (struct tags define everything) │
└──────────┬───────────────────────┘
           │
     buildRegistry()          ← reflection at init() time
           │
           ▼
┌──────────────────────────────────┐
│         config_engine.go         │  ← generic engine
│  BuildDefaults · LoadFromMaps    │
│  ConvertToMaps · SaveToFiles     │
│  LogConfig · SensitiveKeys       │
└──────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────┐
│  application_properties_files.go │  ← file I/O + cross-field validation
│  ReadApplicationPropertiesFile   │
│  SaveConfigurationToFiles        │
│  InitialiseConfig                │
└──────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────┐
│       config_watcher.go          │  ← polls files for external changes
│       WatchConfigFiles           │
└──────────────────────────────────┘
```

## Adding a New Configuration Property

To add a new property you only need to edit **one place** — the
`ApplicationConfiguration` struct:

```go
type ApplicationConfiguration struct {
    // ... existing fields ...

    MyNewSetting int `prop:"my.new.setting" default:"42" file:"application" validate:"positive"`
}
```

The struct tags define the behaviour for this new property:

- **Default value** — written to the property file if absent on first load
- **Loading** — parsed from the correct file and typed into the struct field
- **Saving** — serialised back to the correct file when the API writes changes
- **Logging** — printed on startup and reload (unless marked `sensitive`)
- **Validation** — checked during loading according to the `validate` tag

### Struct Tags Reference

| Tag         | Required | Values                                        | Purpose                                               |
|-------------|----------|-----------------------------------------------|-------------------------------------------------------|
| `prop`      | yes      | Dotted key, e.g. `sensor.collection.interval` | Property key in the `.properties` file                |
| `default`   | yes      | String representation of the default value    | Written to file if absent; used as fallback           |
| `file`      | yes      | `application`, `smtp`, or `database`          | Which property file this setting belongs to           |
| `validate`  | no       | `positive` or `non_negative`                  | Per-field integer validation applied during loading   |
| `sensitive` | no       | `true`                                        | Masks value in API responses and log output           |

### Supported Field Types

| Go type  | Parsing                | Serialisation         |
|----------|------------------------|-----------------------|
| `int`    | `strconv.Atoi`         | `strconv.Itoa`        |
| `bool`   | `strconv.ParseBool`    | `strconv.FormatBool`  |
| `string` | used as-is             | used as-is            |

### Cross-Field Validation

Per-field rules (like `positive`) are handled automatically. If
you need a rule that has more complex validation logic, add it to
`validateApplicationProperties()` in `application_properties_files.go`. An
example is the check that `openapi.yaml.location` must be non-empty when
`sensor.discovery.skip` is `false`.

### Post-Processing Hooks

Some fields need transformation after loading that doesn't fit a tag — for
example, resolving relative OAuth file paths against the config directory. These
live in `postProcessConfig()` in `application_configuration.go`.

## File Watcher

`WatchConfigFiles(ctx)` is started during server boot and polls all three
property files for modification-time changes. When an external edit is detected,
the files are re-read and the configuration is reloaded.

### Write Collision Avoidance

The application itself writes to these same files (via `SaveConfigurationToFiles`
when the API updates a property). The watcher and the writer coordinate through
an atomic flag and a completion timestamp:

1. `SaveConfigurationToFiles` calls `markWriteInProgress()` before writing.
2. The watcher checks `IsWriteInProgress()` on every tick — if true, it
   refreshes its mod-time snapshot and skips the reload.
3. After writing completes, `clearWriteInProgress()` records the completion
   timestamp and clears the flag.
4. Even if the write finishes between watcher ticks (the common case — writes
   take ~50ms, ticks every 2s), `LastWriteCompletedAt()` provides the
   completion time. The watcher applies a 3-second cooldown from that timestamp,
   ensuring it does not react to the file system settling.

### Context Cancellation

The watcher goroutine exits when the context passed to
`WatchConfigFiles` is cancelled (e.g. on `SIGINT`/`SIGTERM`).

## Key Globals

| Symbol       | Type                          | Purpose                                               |
|--------------|-------------------------------|-------------------------------------------------------|
| `AppConfig`  | `*ApplicationConfiguration`   | The current live configuration, set by `ReloadConfig` |
| `registry`   | `[]PropertyDef`               | All property definitions, built once at init          |

## Typical Call Flow

**Startup:**
```
serve.go
  → InitialiseConfig(configDir)
      → setConfigPaths
      → ReadApplicationPropertiesFile  (read + validate + merge defaults)
      → ReadSMTPPropertiesFile
      → ReadDatabasePropertiesFile
      → ReloadConfig
          → LoadConfigurationFromMaps  (parse + per-field validate + post-process)
          → LogConfig
  → WatchConfigFiles(ctx)
```

**API property update:**
```
PATCH /api/properties
  → ServiceUpdateProperties
      → ConvertConfigurationToMaps  (current config → maps)
      → merge incoming changes
      → LoadConfigurationFromMaps   (parse + validate)
      → AppConfig = newCfg
      → SaveConfigurationToFiles    (mark write → write files → clear write)
```

**External file edit (watcher):**
```
config_watcher.go tick
  → stat files, compare mod times
  → if changed: ReadApplicationPropertiesFile, ReadSMTPPropertiesFile, ReadDatabasePropertiesFile
  → ReloadConfig
```
