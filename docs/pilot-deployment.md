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

The `pilot_postgres_data` volume is persistent. Back up both the HiGO
application database and the Kratos identity database with consistent dumps:

```bash
docker compose -f docker-compose.pilot.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > higo-app-backup.sql

docker compose -f docker-compose.pilot.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$ORY_KRATOS_DB" > higo-kratos-backup.sql
```

Before treating the backups as valid, restore both dumps into isolated
databases and verify them independently. Confirm that the application
migrations and schema load successfully, and that Kratos starts against the
restored identity database. Do not test restoration against the live pilot
databases or remove the persistent volume during routine updates.

Services use `restart: unless-stopped`. Inspect readiness with
`docker compose -f docker-compose.pilot.yml ps`, the API `/health` endpoint,
and `logs --tail=200 api kratos postgres`. Do not remove the live volume during
routine updates.
