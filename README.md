# Uber Clone

An MVP ride-hailing platform.

## Current status

The project is currently implementing its Deployment Foundation.

## Run locally

1. Create the environment file:

   ```bash
   cp .env.example .env
   ```

2. Start the stack:

   ```bash
   docker compose up -d --build
   ```

3. Verify the API:

   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:

   ```json
   {"status":"ok"}
   ```

4. Verify database readiness through the API:

   ```bash
   curl http://localhost:8080/ready
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
