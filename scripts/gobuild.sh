#!/bin/bash
#
# This script builds and version stamps the output

VERBOSE=${VERBOSE:-"0"}
V=""
if [[ "${VERBOSE}" == "1" ]];then
    V="-x"
    set -x
fi

OUT=${1:?"output path"}
shift

set -e

BUILD_GOOS=${GOOS:-linux}
BUILD_GOARCH=${GOARCH:-amd64}
GOBINARY=${GOBINARY:-go}
BUILDINFO=${BUILDINFO:-""}
STATIC=${STATIC:-1}
GOBUILDFLAGS=${GOBUILDFLAGS:-}
GCFLAGS=${GCFLAGS:-}
LDFLAGS=${LDFLAGS:-"-extldflags -static"}
# Split GOBUILDFLAGS by spaces into an array called GOBUILDFLAGS_ARRAY.
IFS=' ' read -r -a GOBUILDFLAGS_ARRAY <<< "$GOBUILDFLAGS"

export CGO_ENABLED=0

if [[ "${STATIC}" !=  "1" ]];then
    LDFLAGS=""
fi

time GOOS=${BUILD_GOOS} GOARCH=${BUILD_GOARCH} ${GOBINARY} build \
        ${V} "${GOBUILDFLAGS_ARRAY[@]}" ${GCFLAGS:+-gcflags "${GCFLAGS}"} \
        -o "${OUT}" \
        -ldflags "${LDFLAGS}" "${@}"
