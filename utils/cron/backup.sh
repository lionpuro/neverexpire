#!/bin/bash

mkdir -p /pg_dump
DUMP_PATH="/pg_dump/${POSTGRES_DB}_$(date +%F).sql.gz"

pg_dump "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_DOCKER_PORT}/${POSTGRES_DB}?sslmode=disable" | gzip -c > ${DUMP_PATH}

if [ ! -f "$DUMP_PATH" ]; then
	echo "Error: pg_dump file not found"
	exit 1
fi

RCLONE_CONFIG_CONTENT=$(cat <<EOL
[r2]
type = s3
provider = Cloudflare
access_key_id = ${R2_ACCESS_KEY_ID}
secret_access_key = ${R2_SECRET_ACCESS_KEY}
region = auto
endpoint = ${R2_ENDPOINT}
acl = private
no_check_bucket = true
EOL
)

mkdir -p /opt/rclone
RCLONE_CONFIG="/opt/rclone/rclone.conf"
touch "$RCLONE_CONFIG"
echo "$RCLONE_CONFIG_CONTENT" > "$RCLONE_CONFIG"

rclone copy --config="$RCLONE_CONFIG" $DUMP_PATH r2:"${R2_BUCKET_NAME}/${POSTGRES_DB}"

if [ $? -eq 0 ]; then
	echo "Backup of database ${POSTGRES_DB} successfully uploaded to R2"
else
	echo "Failed to upload backup database ${POSTGRES_DB}"
	exit 1
fi
