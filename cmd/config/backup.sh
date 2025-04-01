#!/bin/bash
DATE=$(date +%Y%m%d-%H%M%S)
docker exec cansu.dev-pg-primary pg_dumpall -U postgres | gzip -9c > $HOME/backups/db-$DATE.sql.gz
