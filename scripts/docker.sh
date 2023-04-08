#!/usr/bin/env bash

set -eux

version=${1:-}
if [ "$version" = latest ]; then
	docker build -t aquaproj-aqua-dev -f Dockerfile-prebuilt .
else
	GOOS=linux go build -o dist/aqua-docker ./cmd/aqua
	docker build -t aquaproj-aqua-dev .
fi

docker run --rm -ti aquaproj-aqua-dev bash
