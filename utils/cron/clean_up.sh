#!/bin/bash

psql "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_DOCKER_PORT}/${POSTGRES_DB}?sslmode=disable" -f clean_up.sql
