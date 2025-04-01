#!/bin/bash
# onepassword service token, DO NOT authenticate your main (root) account on 1password cli.
# by exporting this, cli will automatically login to the service account
export OP_SERVICE_ACCOUNT_TOKEN=...
# -E is required, i spent more time than im willing to admit for this issue. 
# op run creates environment for the command, but then sudo ignores that environment and creates its own env with minimal variables.
op run --env-file=".env" -- sudo -E wal-g backup-push /var/lib/postgresql/data 
