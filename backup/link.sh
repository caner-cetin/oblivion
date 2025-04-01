#!/bin/bash
# https://github.com/wal-g/wal-g/issues/1953
CONTAINER="cansu.dev-pg-primary"
POSTGRES_VOLUME_PATH=$(docker inspect ${CONTAINER} | jq -r '.[0].Mounts[] | select(.Destination=="/var/lib/postgresql/data") | .Source')
sudo ln -s $POSTGRES_VOLUME_PATH /var/lib/postgresql/data
