#!/bin/bash

set -xe

docker compose down
docker compose build

log_dir=/var/log/device-hub
log_path="$log_dir/app.log"

mkdir -p "$log_dir"

INFLUXDB_USERNAME="${INFLUXDB_USERNAME}" \
INFLUXDB_PASSWORD="${INFLUXDB_PASSWORD}" \
INFLUXDB_ORG="${INFLUXDB_ORG}" \
INFLUXDB_ADMIN_TOKEN="${INFLUXDB_ADMIN_TOKEN}" \
INFLUXDB_API_TOKEN="${INFLUXDB_API_TOKEN}" \
INFLUXDB_BUCKET="device_data" \
DEVICE_HUB_LOG_PATH="$log_path" \
docker compose up -d
