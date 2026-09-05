#!/usr/bin/env sh
set -eu

docker build \
  -f deployment/Dockerfile.base \
  -t insurance-core-api-base:latest \
  .
