#!/bin/bash

wgo -debounce 100ms -xdir assets/public \
	npm run build \
	:: go run ./cmd/app \
	:: wgo go run ./cmd/service
