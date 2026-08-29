#!/bin/sh
set -eu

if [ -n "${ORY_KRATOS_DB:-}" ] && [ "$ORY_KRATOS_DB" != "$POSTGRES_DB" ]; then
  if ! psql -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$ORY_KRATOS_DB'" | grep -q 1; then
    psql -U "$POSTGRES_USER" -d postgres -c "CREATE DATABASE \"$ORY_KRATOS_DB\""
  fi
fi
