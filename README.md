# survoyce

Voice-based customer feedback pipeline: audio upload -> queue -> transcription -> analysis -> dashboard.

## Quick start

```bash
go mod tidy

# Start Postgres, Redis, RustFS
make up

# Create tables
make migrate

# Run the API
make run-api

# Run the store integration test
make test-integration
```

## Architecture
_TODO: add diagram and design decisions._
