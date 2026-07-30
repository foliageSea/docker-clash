#!/bin/sh

set -eu

usage() {
  cat <<'EOF'
Usage: sh scripts/build-docker-image.sh [tag] [arch] [output-dir]

Build and export the Docker image as a gzip-compressed tar archive.

Arguments:
  tag         Image tag (default: local)
  arch        Target architecture: amd64 or arm64 (default: host architecture)
  output-dir  Archive output directory (default: dist)

Environment:
  IMAGE_NAME       Image repository (default: foliage-sea/docker-clash)
  MIHOMO_VERSION   mihomo version passed to the Docker build (default: v1.19.28)
EOF
}

case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
esac

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

TAG=${1:-local}
ARCH=${2:-}
OUTPUT_DIR=${3:-"$ROOT_DIR/dist"}
IMAGE_NAME=${IMAGE_NAME:-foliage-sea/docker-clash}
MIHOMO_VERSION=${MIHOMO_VERSION:-v1.19.28}

if ! printf '%s\n' "$TAG" | grep -Eq '^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$'; then
  echo "Invalid image tag: $TAG" >&2
  exit 1
fi

if [ -z "$ARCH" ]; then
  case "$(uname -m)" in
    x86_64|amd64) ARCH=amd64 ;;
    aarch64|arm64) ARCH=arm64 ;;
    *)
      echo "Cannot infer a supported architecture; pass amd64 or arm64 explicitly." >&2
      exit 1
      ;;
  esac
fi

case "$ARCH" in
  amd64|arm64) ;;
  *)
    echo "Unsupported architecture: $ARCH (expected amd64 or arm64)" >&2
    exit 1
    ;;
esac

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker CLI was not found in PATH." >&2
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

IMAGE_REF="$IMAGE_NAME:$TAG"
ARCHIVE="$OUTPUT_DIR/docker-clash-$TAG-linux-$ARCH.tar.gz"
TAR_TMP="$ARCHIVE.tar.tmp"
GZIP_TMP="$ARCHIVE.tmp"

cleanup() {
  rm -f "$TAR_TMP" "$GZIP_TMP"
}
trap cleanup EXIT HUP INT TERM

echo "Building $IMAGE_REF for linux/$ARCH..."
docker buildx build \
  --platform "linux/$ARCH" \
  --build-arg "MIHOMO_VERSION=$MIHOMO_VERSION" \
  --tag "$IMAGE_REF" \
  --load \
  "$ROOT_DIR"

echo "Exporting $ARCHIVE..."
docker image save --output "$TAR_TMP" "$IMAGE_REF"
gzip -9 -c "$TAR_TMP" > "$GZIP_TMP"
mv "$GZIP_TMP" "$ARCHIVE"

echo "Created $ARCHIVE"
