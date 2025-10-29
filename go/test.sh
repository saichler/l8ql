#!/usr/bin/env bash

# Fail on errors and don't open cover file
set -e
# clean up
rm -rf go.sum
rm -rf go.mod
rm -rf vendor

# fetch dependencies
go mod init
GOPROXY=direct GOPRIVATE=github.com go mod tidy
go mod vendor

echo "******************************************************"
echo "* Make sure you built security before running this tests"
echo "* Shallow Security exist in https://github.com/saichler/l8utils/tree/main/go/utils/shallow_security/build.sh"
echo "******************************************************"
read -n 1 -s -r -p "Press any key to continue..."

# Run unit tests with coverage
go test -tags=unit -v -coverpkg=./gsql/... -coverprofile=cover.html ./... --failfast

rm -rf ./tests/loader.so

# Open the coverage report in a browser
go tool cover -html=cover.html
