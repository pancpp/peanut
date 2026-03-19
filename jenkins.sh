#!/bin/bash

set -euo pipefail

BUILDSH_DIR="$(realpath "$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)")"
cd ${BUILDSH_DIR}

VERSION_TAG=${VERSION_TAG:-"local-$(whoami)"}
GIT_HASH=${GIT_HASH:-"$(git rev-parse HEAD)"}
BUILD_TIME=${BUILD_TIME:-"$(TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ')"}
BUILD_NUMBER=${BUILD_NUMBER:-0}

# Build
go build -ldflags "\
  -s -w \
  -X github.com/pancpp/peanut/conf.gVersion=${VERSION_TAG} \
  -X github.com/pancpp/peanut/conf.gBuildTime=${BUILD_TIME} \
  -X github.com/pancpp/peanut/conf.gGitHash=${GIT_HASH} \
  -X github.com/pancpp/peanut/conf.gBuildNumber=${BUILD_NUMBER}" \
  -o peanut
