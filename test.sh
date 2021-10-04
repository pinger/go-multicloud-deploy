#!/bin/bash

cd src/
go build -o lambda lambda.go

cd ../tests/aws/
go test -v -timeout 30m

cd ../../
rm src/lambda