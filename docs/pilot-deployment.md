# Pilot deployment

Use `docker-compose.yml` for local development and
`docker-compose.pilot.yml` for the pilot behind an HTTPS reverse proxy. The
pilot file has no local credential fallbacks, does not publish PostgreSQL, does
not run Mailpit, and starts Kratos without `--dev`.

Provide `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `ORY_KRATOS_DB`,
`KRATOS_PUBLIC_URL`, `KRATOS_BROWSER_RETURN_URL`, `KRATOS_SMTP_CONNECTION_URI`,
`KRATOS_SMTP_FROM_ADDRESS`, `KRATOS_SMTP_FROM_NAME`, `ADMIN_REVIEW_USERNAME`,
`ADMIN_REVIEW_PASSWORD`, and `ADMIN_REVIEW_ORIGIN` externally. Use HTTPS URLs
for public origins and preserve the public `Host` header through the proxy.

The `pilot_postgres_data` volume is persistent. Back it up with a consistent
dump and verify restoration into an isolated database before relying on it:

```bash
docker compose -f docker-compose.pilot.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > backup.sql
```

Services use `restart: unless-stopped`. Inspect readiness with
`docker compose -f docker-compose.pilot.yml ps`, the API `/health` endpoint,
and `logs --tail=200 api kratos postgres`. Do not remove the live volume during
routine updates.
