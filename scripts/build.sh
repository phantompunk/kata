#!/bin/bash

## Build script for MacOS/Linux artifacts

set -eou pipefail

VERSION=${VERSION:-$(git describe --tags --abbrev=0)}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD)}
BINARY_NAME="kata"

GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}

LDFLAGS="-s -w -X github.com/phantompunk/kata/cmd.version=${VERSION} -X github.com/phantompunk/kata/cmd.commit=${COMMIT}"

OUTPUT_DIR="dist"
OUTPUT="${OUTPUT_DIR}/${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"
BINARY="${OUTPUT}/${BINARY_NAME}"

[ -n "$GOARCH" ] && export GOARCH
[ -n "${CC:-}" ] && export CC
export CGO_ENABLED=1

mkdir -p ${OUTPUT}
echo "Building ${BINARY_NAME} ${VERSION} (${COMMIT}) for ${GOOS}/${GOARCH}"
go build -ldflags "$LDFLAGS" -o "${BINARY}" .
echo "Built $BINARY"

tar -czf ${OUTPUT}.tar.gz -C ${OUTPUT} ${BINARY_NAME} 
echo "Compressed as ${OUTPUT}.tar.gz"

