#!/usr/bin/env bash
set -euo pipefail

docker build .

image_id=$(docker build -q .)

args=(
  -ti
  --rm
  --mount "type=bind,source=$PWD,target=/code"
  --user "${DOCKER_USER:-$(id -u):$(id -g)}" 
)


docker run "${args[@]}" "$image_id" "$@"
