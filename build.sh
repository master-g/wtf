#!/usr/bin/env bash

# format
echo "formatting..."
goreturns -w $(find . -type f -name '*.go' -not -path "./vendor/*")

# mod
echo "go mod tidy and vendoring..."
go mod tidy
go mod vendor

# lint
echo "linting..."
gometalinter	--vendor \
							--fast \
							--enable-gc \
							--tests \
							--aggregate \
							--disable=gotype \
							${PWD}

# build
echo "building..."
go build

