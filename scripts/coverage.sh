#!/usr/bin/env bash

set -eu
set -o pipefail

cd "$(dirname "$0")/.."

if [ $# -eq 0 ]; then
  target="$(go list ./... | fzf)"
  profile=.coverage/$target/coverage.txt
  mkdir -p .coverage/"$target"
elif [ $# -eq 1 ]; then
  target=$1
  mkdir -p .coverage/"$target"
  profile=.coverage/$target/coverage.txt
  target=./$target
else
  echo "too many arguments are given: $*" >&2
  exit 1
fi

go test "$target" -coverprofile="$profile" -covermode=atomic
go tool cover -html="$profile"
