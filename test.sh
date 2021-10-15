#!/bin/bash

cd src/aws/
./build.sh

cd ../../tests/common/
go mod tidy
go test -v -timeout 30m

cd ../../
rm -f src/bin/main
