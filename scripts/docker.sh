#!/usr/bin/env bash
# This script is run by `cmdx docker`.

set -eux

sha=""
if git diff --quiet; then
  sha=$(git rev-parse HEAD)
fi

case "$TARGET" in
    build)
       	GOOS=linux go build -ldflags "-X main.version=v1.0.0-local -X main.commit=$sha -X main.date=$(date +"%Y-%m-%dT%H:%M:%SZ%:z" | tr -d '+')" -o dist/aqua-docker ./cmd/aqua
       	image=aquaproj-aqua-build
		;;
	prebuilt)
       	image=aquaproj-aqua-prebuilt
		;;
	alpine-build)
    	GOOS=linux go build -ldflags "-X main.version=v1.0.0-local -X main.commit=$sha -X main.date=$(date +"%Y-%m-%dT%H:%M:%SZ%:z" | tr -d '+')" -o dist/aqua-docker ./cmd/aqua
    	image=aquaproj-aqua-alpine-build
		;;
	alpine-prebuilt)
    	image=aquaproj-aqua-alpine-prebuilt
		;;
	*)
		echo "[ERROR] Invalid target: $TARGET" >&2
		exit 1
		;;
esac

docker build -t "$image" --target "$TARGET" .
docker run --rm -ti -v "$PWD:/home/foo/workspace" "$image" bash
