#!/bin/bash

# Build and push Docker image to GitHub Container Registry (ghcr.io)
# Usage: ./scripts/build-and-push.sh [tag] [registry]
#   tag: Image tag (default: latest)
#   registry: ghcr.io or gcr.io (default: ghcr.io)

set -e

REGISTRY=${2:-"ghcr.io"}
TAG=${1:-"latest"}

if [ "$REGISTRY" = "ghcr.io" ]; then
  # GitHub Container Registry
  GITHUB_OWNER=${GITHUB_OWNER:-"your-github-username-or-org"}
  IMAGE_NAME="posthoot"  # Repository name
  FULL_IMAGE_NAME="${REGISTRY}/${GITHUB_OWNER}/${IMAGE_NAME}:${TAG}"
  
  echo "Building Docker image: ${FULL_IMAGE_NAME}"
  docker build -t "${FULL_IMAGE_NAME}" .
  
  # Login to GitHub Container Registry
  echo "Logging in to GitHub Container Registry..."
  echo "$GITHUB_TOKEN" | docker login "${REGISTRY}" -u "${GITHUB_USERNAME:-$GITHUB_OWNER}" --password-stdin
  
  echo "Pushing Docker image to ghcr.io..."
  docker push "${FULL_IMAGE_NAME}"
  
  echo "Successfully pushed ${FULL_IMAGE_NAME}"
  echo ""
  echo "Note: Make sure GITHUB_TOKEN is set in your environment"
  echo "You can create a token at: https://github.com/settings/tokens"
  echo "Required scope: write:packages"
  
elif [ "$REGISTRY" = "gcr.io" ]; then
  # Google Container Registry
  PROJECT_ID=${GCP_PROJECT_ID:-"your-project-id"}
  IMAGE_NAME="posthoot"
  FULL_IMAGE_NAME="${REGISTRY}/${PROJECT_ID}/${IMAGE_NAME}:${TAG}"
  
  echo "Building Docker image: ${FULL_IMAGE_NAME}"
  docker build -t "${FULL_IMAGE_NAME}" .
  
  # Configure Docker to use gcloud as a credential helper
  gcloud auth configure-docker "${REGISTRY}" --quiet
  
  echo "Pushing Docker image to GCR..."
  docker push "${FULL_IMAGE_NAME}"
  
  echo "Successfully pushed ${FULL_IMAGE_NAME}"
else
  echo "Error: Unsupported registry: $REGISTRY"
  echo "Supported registries: ghcr.io, gcr.io"
  exit 1
fi

