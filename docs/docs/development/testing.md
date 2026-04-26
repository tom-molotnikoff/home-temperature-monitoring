# Testing

## Running Tests

Run all tests from the `sensor_hub` directory:

```bash
cd sensor_hub
go test ./...
```

Run a specific package:

```bash
go test ./api/
```

## Integration Tests

Integration tests exercise the full stack — HTTP API through to a real SQLite
database — with testcontainers-managed mock sensor Docker containers.

### Prerequisites

- Docker Engine must be running (testcontainers-go manages containers automatically)

### Running Integration Tests

```bash
cd sensor_hub
go test -tags integration -v -timeout 300s ./integration/
```

Run a specific integration test:

```bash
go test -tags integration -v -run TestCollection_CollectAll -timeout 300s ./integration/
```

Integration tests use the `//go:build integration` build tag and are excluded
from normal `go test ./...` runs. They run automatically in CI on pull requests,
main branch pushes, and release tags.