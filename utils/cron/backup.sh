#!/bin/bash

mkdir -p /pg_dump
DUMP_PATH="/pg_dump/${POSTGRES_DB}_$(date +%F).sql.gz"

pg_dump "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_DOCKER_PORT}/${POSTGRES_DB}?sslmode=disable" | gzip -c > ${DUMP_PATH}

if [ ! -f "$DUMP_PATH" ]; then
	echo "Error: pg_dump file not found"
	exit 1
fi

rclone copy $DUMP_PATH r2:"${R2_BUCKET_NAME}/${POSTGRES_DB}"

if [ $? -eq 0 ]; then
	echo "Backup of database ${POSTGRES_DB} successfully uploaded to R2"
else
	echo "Failed to upload backup database ${POSTGRES_DB}"
	exit 1
fi
