#!/usr/bin/env bash
# This script is run by `cmdx docker`.

set -eux

if [ "$BUILD" = true ]; then
	GOOS=linux go build -o dist/aqua-docker ./cmd/aqua
	target=build
	image=aquaproj-aqua-build
else
	image=aquaproj-aqua-prebuilt
	target=prebuilt
fi

docker build -t "$image" --target "$target" .
docker run --rm -ti "$image" bash
