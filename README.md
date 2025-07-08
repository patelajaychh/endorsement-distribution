# Endorsement Distribution Microservice

A standalone microservice for distributing endorsements and trust anchors to verifiers via REST API.

## Overview

This service provides a REST API endpoint that:
1. Accepts CoSERV queries in `application/coserv+cbor` format
2. Extracts query parameters and generates database keys
3. Fetches artefacts from PostgreSQL database
4. Packages results in CoSERV format and returns them

## API Endpoints

- `GET /endorsement-distribution/v1/coserv/:query` - Main endpoint for fetching endorsements
- `GET /.well-known/veraison/endorsement-distribution` - Service capability information

## Configuration

The service uses environment variables or config files for configuration:

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5432
  name: "endorsements"
  user: "postgres"
  password: "password"
  sslmode: "disable"

logging:
  level: "info"
```

## Database Schema

```sql
CREATE TABLE endorsements (
  kv_key text NOT NULL,
  kv_val text NOT NULL
);
```

## Key Format

Keys follow the format: `coserv://tenant/{profile}/{artifact-type}/{environment-selector-hash}`

## Build and Run

```bash
go build -o endorsement-distribution
./endorsement-distribution
```

## License

Apache-2.0 