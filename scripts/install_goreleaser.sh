#!/bin/bash

VERSION=${VERSION:-2.3.2}
PLATFORM=${PLATFORM:-Linux_x86_64}
GO_BIN="$(go env GOPATH)/bin"

curl -L https://github.com/goreleaser/goreleaser/releases/download/v${VERSION}/goreleaser_${PLATFORM}.tar.gz | tar -xz -C /tmp/
mv /tmp/goreleaser ${GO_BIN}/