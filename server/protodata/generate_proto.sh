#!/bin/bash

export PATH="$PATH:$(go env GOPATH)/bin"
export PATH="$PATH":"$HOME/.pub-cache/bin"

echo "Generating Go code..."
protoc --go_out=. --go_opt=paths=source_relative transit.proto

echo "Generating Dart code..."
protoc --dart_out=. transit.proto

echo "Done!"


# run with chmod +x gen_proto.sh