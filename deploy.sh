#!/bin/bash
set -euo pipefail

REMOTE_HOST="${REMOTE_HOST:-192.168.2.124}"
REMOTE_PORT="${REMOTE_PORT:-2222}"
REMOTE_USER="${REMOTE_USER:-cburns}"
REMOTE_DOCKER_PATH="${REMOTE_DOCKER_PATH:-/home/cburns/docker/poker}"
REMOTE_ARCHIVE_PATH="${REMOTE_ARCHIVE_PATH:-/tmp/poker-images.tar}"

FRONTEND_IMAGE="${FRONTEND_IMAGE:-ghcr.io/burnsco/poker-frontend:latest}"
BACKEND_IMAGE="${BACKEND_IMAGE:-ghcr.io/burnsco/poker-backend:latest}"

ARCHIVE_PATH="$(mktemp -t poker-images.XXXXXX.tar)"
trap 'rm -f "$ARCHIVE_PATH"' EXIT

echo "🚀 Deploying Poker..."

echo "📦 Building Frontend..."
docker build --network=host -f Dockerfile.frontend -t "$FRONTEND_IMAGE" .

echo "📦 Building Backend..."
docker build --network=host -f Dockerfile.backend -t "$BACKEND_IMAGE" .

echo "📦 Packaging images for transfer..."
docker save -o "$ARCHIVE_PATH" "$FRONTEND_IMAGE" "$BACKEND_IMAGE"

echo "⬆️  Copying images to mordor..."
scp -P "$REMOTE_PORT" "$ARCHIVE_PATH" "$REMOTE_USER@$REMOTE_HOST:$REMOTE_ARCHIVE_PATH"

echo "🚢 Loading images and restarting the server stack..."
remote_cmd=$(printf 'set -euo pipefail; docker load -i %q >/dev/null; rm -f %q; cd %q && docker compose up -d --no-build' \
  "$REMOTE_ARCHIVE_PATH" \
  "$REMOTE_ARCHIVE_PATH" \
  "$REMOTE_DOCKER_PATH")
ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "bash -lc $(printf '%q' "$remote_cmd")"

echo "✅ Deployment complete!"
