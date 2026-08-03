#!/bin/bash
# Regenerates the stock (reflection-based) protoc-gen-go output that the
# generator reads as its test input. Needs protoc; CI does not run this, the
# .pb.go is checked in.
#
# The embedpb side is NOT generated here — genpb's TestFixturesUpToDate emits
# ../embedded and ../embeddedrt, and fails if they drift:
#
#	go test ./genpb -run TestFixturesUpToDate -update
set -e
script_path=$(dirname "$(realpath "$0")")
bin="$script_path/.bin"
cd "$script_path"
GOBIN="$bin" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
PATH="$bin:$PATH" protoc -I ./ ./shapes.proto --go_out=. --go_opt=paths=source_relative
