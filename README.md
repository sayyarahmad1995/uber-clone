# Uber Clone

An MVP ride-hailing platform.

## Current status

The project is currently implementing its Deployment Foundation.

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

Redis, CI/CD, Kubernetes, and other future infrastructure are intentionally not part of the current milestone.
