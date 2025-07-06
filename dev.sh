#!/bin/bash

# Helpers

usage() {
	echo "Usage: $0 [COMMAND] [ARGUMENTS]"
	echo "Commands:"
	echo "  start     Start up the development containers"
	echo "  stop      Shut down the development containers"
	echo "  cmd       Run a command in the workspace container"
	echo "  shell     Open a shell into the workspace container"
	echo "  fmt       Format all code"
}

fn_exists() {
    type $1 2>/dev/null | grep -q 'is a function'
}

COMMAND=$1
shift
ARGUMENTS=${@}

# Commands

COMPOSE_FILE="compose.dev.yaml"

start() {
	docker compose -f $COMPOSE_FILE up
}

stop() {
	docker compose -f $COMPOSE_FILE stop
}

cmd() {
	docker compose -f $COMPOSE_FILE exec workspace ${@}
}

shell() {
	docker compose -f $COMPOSE_FILE exec workspace bash
}

fmt() {
	docker compose -f $COMPOSE_FILE exec workspace sh -c \
		"gofmt -l -s -w /app && npx prettier /app --write"
}

# Execute

fn_exists $COMMAND
if [ $? -eq 0 ]; then
	$COMMAND $ARGUMENTS
else
	usage
fi
