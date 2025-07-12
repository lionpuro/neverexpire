#!/bin/bash

cleanup() {
    echo "Cleaning up..."
    exit
}

trap cleanup EXIT

wgo -debounce 100ms -xdir assets/public \
	npm run build \
	:: go run ./cmd/app \
	:: wgo go run ./cmd/worker
