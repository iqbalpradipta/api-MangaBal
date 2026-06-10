#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DEPLOY_DIR="$(dirname "$SCRIPT_DIR")"

cd "$DEPLOY_DIR"

if [ ! -f .env.production ]; then
  cp .env.production.example .env.production
  echo "ERROR: .env.production created from template."
  echo "Edit .env.production with real DB, admin, internal, and BalStorage secrets before continuing."
  exit 1
fi

ensure_secret() {
  local key="$1"
  local current
  current="$(grep -E "^${key}=" .env.production | tail -n 1 | cut -d= -f2- || true)"

  if [ "${#current}" -ge 32 ] && [ "${current#<}" = "${current}" ]; then
    return
  fi

  local generated
  if command -v openssl >/dev/null 2>&1; then
    generated="$(openssl rand -hex 32)"
  else
    generated="$(dd if=/dev/urandom bs=48 count=1 2>/dev/null | base64 | tr -dc 'A-Za-z0-9' | cut -c1-64)"
  fi

  if grep -qE "^${key}=" .env.production; then
    sed -i "s|^${key}=.*|${key}=${generated}|" .env.production
  else
    printf '\n%s=%s\n' "$key" "$generated" >> .env.production
  fi
  echo "Generated ${key} in .env.production"
}

ensure_secret "ADMIN_TOKEN"
ensure_secret "INGEST_INTERNAL_TOKEN"
ensure_secret "DASHBOARD_PASSWORD"
chmod 600 .env.production || true

echo "Building and starting services..."
docker compose build --pull manga-api
docker compose build --pull dashboard

legacy_project="$(docker inspect -f '{{ index .Config.Labels "com.docker.compose.project" }}' manga-api 2>/dev/null || true)"
if [ -n "$legacy_project" ] && [ "$legacy_project" != "api-mangabal" ]; then
  echo "Removing legacy manga-api container from compose project: $legacy_project"
  docker rm -f manga-api
fi

docker compose up -d --force-recreate --remove-orphans

echo "Waiting for service health..."
sleep 10

echo "=== Manga API health ==="
if curl -sf http://localhost:8001/api/v1/health; then
  echo " [OK]"
else
  echo " [FAIL]"
  echo ""
  echo "=== docker compose ps ==="
  docker compose ps
  echo ""
  echo "=== manga-api logs ==="
  docker logs --tail=200 manga-api || true
  exit 1
fi

echo ""
echo "=== Dashboard ==="
if curl -sf -o /dev/null http://localhost:3001; then
  echo "Dashboard [OK]"
else
  echo "Dashboard [FAIL]"
  echo ""
  echo "=== dashboard logs ==="
  docker logs --tail=50 manga-dashboard || true
fi

echo ""
echo "Deploy complete."
echo "API:       https://manga.iqbalpradipta.my.id"
echo "Dashboard: https://manga.iqbalpradipta.my.id/dashboard"
