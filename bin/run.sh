#!/bin/sh
set -eu

echo "[run.sh] Starting service"

has_migrations=false
for migration in ./db/migrations/*.sql ./db/migrations/*.go; do
	if [ -f "$migration" ]; then
		has_migrations=true
		break
	fi
done

if [ "$has_migrations" = true ]; then
	echo "[run.sh] Running DB migrations"
	: "${DATABASE_URL:?DATABASE_URL is required}"
	goose -dir /app/db/migrations postgres "$DATABASE_URL" up
else
	echo "[run.sh] No DB migrations found, skipping"
fi

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app"
exec /app/bin/app
