#!/bin/bash
set -e

REMOTE_HOST="192.168.2.124"
REMOTE_PORT="2222"
REMOTE_USER="cburns"
REMOTE_DOCKER_PATH="/home/cburns/docker/poker"

FRONTEND_IMAGE="ghcr.io/burnsco/poker-frontend:latest"
BACKEND_IMAGE="ghcr.io/burnsco/poker-backend:latest"

echo "🚀 Deploying Poker..."

echo "📦 Building Frontend..."
docker build --network=host -f Dockerfile.frontend -t "$FRONTEND_IMAGE" .

echo "📦 Building Backend..."
docker build --network=host -f Dockerfile.backend -t "$BACKEND_IMAGE" .

echo "⬆️  Pushing images to GHCR..."
docker push "$FRONTEND_IMAGE"
docker push "$BACKEND_IMAGE"

echo "🚢 Deploying to server..."
ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_DOCKER_PATH && docker compose pull && docker compose up -d"

echo "✅ Deployment complete!"
