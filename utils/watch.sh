#!/bin/bash

cleanup() {
    echo "Cleaning up..."
    exit
}

trap cleanup EXIT

wgo -debounce 100ms -xdir assets/public \
	npm run build \
	:: go run ./cmd/web \
	:: wgo go run ./cmd/worker
