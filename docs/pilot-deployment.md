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

## Suggested fare staged rollout

1. Deploy migration 031 and backend with `SUGGESTED_FARES_ENABLED=false` (default). Verify manual booking and existing Trip recovery. Backups include immutable pricing history and request snapshots.
2. Publish owner-approved PKR tariffs through Operations → Pricing for each intended active visible service. Enter all five values in rupees; verify history, reviewer and current identity. No example/test tariff is production approval. Publication does not enable suggestions.
3. Deploy the compatible Flutter client and complete the physical acceptance matrix in `pilot-google-booking-acceptance.md`. Deploy backend before testing Driver routes.
4. Explicitly set `SUGGESTED_FARES_ENABLED=true` and recreate the API container. Confirm catalog `pricing_required`, preview amount/version, edited booking and conflict recovery. Record the enable time and commits.
5. For rollback, set the flag false and recreate the API container. Existing requests retain proposals and snapshots; agreed Trips retain the selected offer. Disabling a current policy prevents new priced requests for that service, without deleting history or changing active Trips.

Do not enable from CI evidence alone. Physical Android validation and real tariff activation are pending.
