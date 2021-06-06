#!/bin/bash

set -ux

(
	cd demo/client
	GOOS=js	GOARCH=wasm go build -o ../../static/main.wasm main.go
)

(
	cd demo/server
	go build -o ../../srv main.go
)

if [[ "$1" == "run" ]]; then
	./srv
fi
