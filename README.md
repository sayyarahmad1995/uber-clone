# Uber Clone

An MVP ride-hailing platform.

## Current status

The Go backend implements account authentication, Driver operations, Rider ride
requests, Driver offers, Rider-selected assignment, Trip execution, cancellation,
and history. One shared Flutter client is planned; it is not yet in this repository.

Drivers accepting the Rider's proposed fare create a pending offer. Only the
Rider's selection assigns a Trip. See [the marketplace decision](docs/ADR-0007-ride-request-marketplace-model.md)
and [the worklog](docs/WORKLOG.md) for scope and remaining work.

The [geographic marketplace API](docs/geographic-marketplace-api.md) ranks requests
by pickup distance and gives Riders vehicle, fare, distance, and availability
details. Drivers need fresh location updates to discover requests or make offers.

## Run locally

1. Start the stack:

   ```bash
   docker compose up -d --build
   ```

2. Verify the API:

   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:

   ```json
   {"status":"ok"}
   ```

3. Check the running services:

   ```bash
   docker compose ps
   ```

## Stop the stack

```bash
docker compose down
```

To also remove the PostgreSQL data volume:

```bash
docker compose down -v
```

## Current deployment scope

- Go API
- PostgreSQL
- Docker
- Docker Compose
- Ory Kratos and a local Mailpit email service

Redis, CI/CD, Kubernetes, and other future infrastructure are intentionally not part of the current milestone.

## Schema upgrades and tests

The API applies pending migrations at startup. Migration 016 removes the legacy
candidate table and booking-mode column; stop old API processes before upgrading.
See [the retirement rollout guide](docs/candidate-retirement-rollout.md) for data
handling, validation, and recovery.

Backend verification, using a dedicated PostgreSQL database whose name ends in
`_test`:

```bash
cd backend
export TEST_DATABASE_URL='postgres://USER:PASSWORD@localhost:5432/uber_clone_test?sslmode=disable'
go test -p 1 ./...
go vet ./...
```

Database integration tests skip when `TEST_DATABASE_URL` is absent. Run packages
serially on a fresh shared test database to avoid concurrent migration startup.
