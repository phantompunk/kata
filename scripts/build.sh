#!/bin/bash

## Build script for MacOS/Linux/Windows artifacts

set -eou pipefail

VERSION=${VERSION:-$(git describe --tags --abbrev=0)}
COMMIT=${COMMIT:-$(git rev-parse --short HEAD)}
BINARY_NAME="kata"

GOOS=${GOOS:-$(go env GOOS)}
GOARCH=${GOARCH:-$(go env GOARCH)}

LDFLAGS="-s -w"

OUTPUT_DIR="dist"
OUTPUT="${OUTPUT_DIR}/${BINARY_NAME}_${VERSION}_${GOOS}_${GOARCH}"

BINARY="${OUTPUT}/${BINARY_NAME}"
if [[ "$GOOS" = "windows" ]]; then
    BINARY="${OUTPUT}/${BINARY_NAME}.exe"
fi

# Export Go build variables for cross compiling 
[ -n "$GOARCH" ] && export GOARCH 
[ -n "${GOOS:-}" ] && export GOOS
[ -n "${CXX:-}" ] && export CXX
[ -n "${CC:-}" ] && export CC
export CGO_ENABLED=1

mkdir -p ${OUTPUT}
echo "Building ${BINARY_NAME} ${VERSION} (${COMMIT}) for ${GOOS}/${GOARCH}"
go build -ldflags "$LDFLAGS" -o "${BINARY}" .
echo "Built $BINARY"

if [[ "$GOOS" = "windows" ]]; then
    echo zip -q ${OUTPUT}.zip ${BINARY}
    zip -jq ${OUTPUT}.zip ${BINARY}
    echo "Compressed as ${OUTPUT}.zip"
else 
    tar -czf ${OUTPUT}.tar.gz -C ${OUTPUT} ${BINARY_NAME} 
    echo "Compressed as ${OUTPUT}.tar.gz"
fi

