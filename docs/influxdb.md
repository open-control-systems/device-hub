## Introduction

This guide shows how to set up and run influxdb in the docker container. For more advanced configuration, see the [documentation](https://hub.docker.com/_/influxdb) for the influxdb base image. See also the official setup [guide](https://docs.influxdata.com/influxdb/v2/install/?t=Docker).

## Configuration

The snippet below demonstrates the basic example of the influxdb service in the docker compose file:

```docker
services:
  influxdb:
    image: influxdb:latest
    container_name: influxdb
    ports:
      - "8086:8086"
    volumes:
      - influxdb-data:/var/lib/influxdb2
    environment:
      - DOCKER_INFLUXDB_INIT_ORG=${DEVICE_HUB_STORAGE_INFLUXDB_ORG}
      - DOCKER_INFLUXDB_INIT_USERNAME=${DEVICE_HUB_STORAGE_INFLUXDB_USERNAME}
      - DOCKER_INFLUXDB_INIT_PASSWORD=${DEVICE_HUB_STORAGE_INFLUXDB_PASSWORD}
      - DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=${DEVICE_HUB_STORAGE_INFLUXDB_ADMIN_TOKEN}
      - DOCKER_INFLUXDB_INIT_MODE=setup
      - DOCKER_INFLUXDB_INIT_BUCKET=${DEVICE_HUB_STORAGE_INFLUXDB_BUCKET}
```

**Run influxdb service**

```bash
# Replace <username>, <password>, <admin>, <bucket>, <org> with the required credentials.
# See also https://docs.influxdata.com/influxdb/cloud/reference/key-concepts/data-elements/.
DEVICE_HUB_STORAGE_INFLUXDB_USERNAME="<username>" \
DEVICE_HUB_STORAGE_INFLUXDB_PASSWORD="<password>" \
DEVICE_HUB_STORAGE_INFLUXDB_ADMIN_TOKEN="<admin_token>" \
DEVICE_HUB_STORAGE_INFLUXDB_BUCKET="<bucket>" \
DEVICE_HUB_STORAGE_INFLUXDB_ORG="<org>" \
docker compose up influxdb -d
```

**Verify influxdb service**

Go to `localhost:8086` in the web-browser, and enter the influxdb credentials, that were last used to run the docker compose service.

**Determine influxdb `org` identifier**

It will be used later to create API tokens.

```bash
curl http://localhost:8086/api/v2/orgs -H "Authorization: Token <admin_token>"
```

**Retrieve influxdb API token**

It's an access token for your application.

```bash
# See also: https://docs.influxdata.com/influxdb/cloud/admin/tokens/create-token/
curl http://localhost:8086/api/v2/authorizations \
  -H "Authorization: Token <admin_token>" \
  -H 'Content-type: application/json' \
  --data '{
  "status": "active",
  "description": "device-hub r/w API token",
  "orgId": "<org_id>",
  "permissions": [
    {
      "action": "write",
      "resource": {
        "type": "buckets"
      }
    },
    {
      "action": "read",
      "resource": {
        "type": "buckets"
      }
    }
  ]
}'
```

Save the "token" field of the response for future use. That's it. You can use this token in your application to communicate with the influxdb.
