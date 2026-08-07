#!/bin/bash
# Regenerates the stock protoc-gen-go output for the management network fixture.
set -e
script_path=$(dirname "$(realpath "$0")")
bin="$script_path/.bin"
cd "$script_path"
GOBIN="$bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
PATH="$bin:$PATH" protoc -I ./ ./management_net.proto --go_out=. --go_opt=paths=source_relative
