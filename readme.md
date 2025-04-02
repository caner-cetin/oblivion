# Oblivion

**A personalized toolkit for provisioning and managing Docker services on a server environment (`*.cansu.dev/*`), primarily built for the my specific needs.**

[![Go Report Card](https://goreportcard.com/badge/github.com/caner-cetin/oblivion)](https://goreportcard.com/report/github.com/caner-cetin/oblivion)

---

**⚠️ Disclaimer:** This project is highly tailored to the my server setup, configuration preferences, and specific services (`cansu.dev`). While parts might be adaptable, it relies on specific conventions (like 1Password secret structures) and may require significant modification for general use. It serves partly as infrastructure-as-code and partly as personal documentation.

---

- [Oblivion](#oblivion)
  - [Core Concepts](#core-concepts)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
    - [1. `.oblivion.toml`](#1-obliviontoml)
    - [2. 1Password Setup](#2-1password-setup)
  - [Usage](#usage)
    - [`networks`](#networks)
    - [`postgres`](#postgres)
    - [`static`](#static)
    - [`kuma`](#kuma)
    - [`observer`](#observer)
    - [`redis`](#redis)
    - [`playground`](#playground)
    - [Backup Scripts (`backup/`)](#backup-scripts-backup)
  - [Example System Configuration (my Setup)](#example-system-configuration-my-setup)
    - [Firewall (`ufw`)](#firewall-ufw)
  - [Development](#development)
  - [Adaptation / Contribution](#adaptation--contribution)
  - [License](#license)


## Core Concepts

*   **Docker-centric:** Manages Docker networks, volumes, containers, and images.
*   **1Password Integration:** Securely fetches credentials (database passwords, API keys, etc.) using a 1Password Service Account. **This is a hard dependency for commands requiring secrets.**
*   **Cobra CLI:** Provides a structured command-line interface.
*   **Configuration Driven:** Uses a TOML file (`~/.oblivion.toml`) for defining container names, ports, image tags, network names, etc.
*   **Service Provisioning:** Includes commands to set up:
    *   Docker Networks
    *   PostgreSQL (Primary + Replica + PgBouncer)
    *   Static File Server (Nginx with specific permissions)
    *   Uptime Kuma (Monitoring)
    *   Observer Stack (Grafana, Prometheus, Loki, cAdvisor, Node Exporter, Alertmanager)
    *   Redis (DragonflyDB)
    *   A custom "Playground" backend service.
*   **Backup Helpers:** Includes shell scripts using `wal-g` for PostgreSQL backups (requires specific setup).

## Prerequisites

*   **Go:** Version 1.24 or higher (see `go.mod`).
*   **Docker:** Docker Engine and Docker CLI installed and running. The user running `oblivion` needs permission to interact with the Docker socket.
*   **1Password CLI:** Installed and configured.
*   **1Password Service Account:** A 1Password Service Account token must be available via the `OP_SERVICE_ACCOUNT_TOKEN` environment variable for commands requiring secrets.
*   **`sudo`:** Required for some operations like `static permissions` (uses `setfacl`) and the `backup/link.sh` script.
*   **`jq`:** Required by the `backup/link.sh` script to parse Docker volume paths.
*   **`acl` package:** Required on Linux systems for `setfacl` used by `static permissions`. (e.g., `sudo apt install acl` on Debian/Ubuntu).
*   **(Optional) `ufw`:** Used in the example firewall setup documented below.

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/caner-cetin/oblivion.git
    cd oblivion
    ```
2.  Build the binary:
    ```bash
    go build -o oblivion .
    ```
3.  (Optional) Move the binary to a location in your `$PATH`:
    ```bash
    sudo mv oblivion /usr/local/bin/
    ```

## Configuration

### 1. `.oblivion.toml`

*   On the first run, if `~/.oblivion.toml` does not exist, Oblivion will create one with default values.
*   **Location:** Defaults to `$HOME/.oblivion.toml`. You can specify a different path using the `--config` flag.
*   **Review and Customize:** **It is crucial to review and customize this file.** Pay special attention to:
    *   `[Docker].Socket`: Ensure this points to your Docker socket (default is `unix:///var/run/docker.sock`).
    *   `[Networks]`: Verify or change network names if desired.
    *   `[Onepass].vault_name`: Set this to the name of the 1Password Vault where your secrets are stored.
    *   `[Static].uploader_user`: Set to the username that will upload files to the static server path.
    *   `[Static].static_path`: The **absolute path** on the host for static files.
    *   **`[Observer].Binds`**: **Critical:** These default to the my local paths. You **must** update these bind mount source paths to point to your actual configuration directories for Prometheus, Grafana, Alertmanager, and Loki.
    *   Other service-specific configurations (ports, container names, image tags, volumes).

### 2. 1Password Setup

*   **Service Account Token:** Export the token before running commands that need secrets:
    ```bash
    export OP_SERVICE_ACCOUNT_TOKEN="ops_your_token_here..."
    ```
*   **Vault:** Ensure the vault specified in `[Onepass].vault_name` exists.
*   **Secret Structure:** Oblivion expects secrets to be stored in 1Password using a specific convention. Secret references in the code and documentation follow this pattern: `op://<Vault Name>/<Item Title>/<Section Name>/<Field Name>` or `op://<Vault Name>/<Item Title>/<Field Name>` if no section is used.
    *   **Example:** A required secret `"/Postgres/Replicator/username"` means:
        1.  In the vault configured in `.oblivion.toml` (e.g., "Server")...
        2.  Find or create a Login or Secure Note item titled "Postgres".
        3.  Within that item, find or create a section named "Replicator".
        4.  Within that section, find or create a field named "username" and store the value there.

## Usage

The general command structure is:

```bash
oblivion [command] [subcommand] [flags]
```

### `networks`

Manages required Docker networks.

*   **`oblivion networks up`**
    *   Creates Docker bridge networks defined in the `[Networks]` section of the config (e.g., `database_bridge`, `uptime_bridge`, `grafana_bridge`, `loki_bridge`).
    *   This should typically be run first.

### `postgres`

Manages PostgreSQL primary, replica, and PgBouncer containers.

*   **Required Secrets:**
    *   `/Postgres/Replicator/username`
    *   `/Postgres/Replicator/password`
    *   `/Postgres/Root/username` (Superuser)
    *   `/Postgres/Root/password` (Superuser)
    *   `/Postgres/Bouncer/username`
    *   `/Postgres/Bouncer/password`
*   **`oblivion postgres up`**
    *   Pulls necessary images (`postgres:17`, `edoburu/pgbouncer` by default).
    *   Creates Docker volumes (`pg_primary_data`, `pg_replica_data` by default).
    *   Starts the primary PostgreSQL container, waits for it to be healthy.
    *   Starts the replica PostgreSQL container.
    *   Starts the PgBouncer container, configured to connect to the primary.
*   **PgBouncer SQL Setup:** For PgBouncer authentication using `AUTH_QUERY`, you need to execute the following SQL on your primary PostgreSQL instance **for each database** you intend to connect to via PgBouncer. Replace `'i_look_cute_in_maid_outfit'` with the actual password you stored in 1Password for `/Postgres/Bouncer/password`.
    ```sql
    -- Run this logged in as the Postgres superuser
    CREATE ROLE pgbouncer LOGIN PASSWORD 'i_look_cute_in_maid_outfit'; -- Use the actual password here!

    -- Create the lookup function in the 'public' schema (or desired schema)
    CREATE OR REPLACE FUNCTION public.lookup (
       INOUT p_user     name,
       OUT   p_password text
    ) RETURNS record
       LANGUAGE sql SECURITY DEFINER SET search_path = pg_catalog AS
    $$SELECT usename, passwd FROM pg_shadow WHERE usename = p_user$$;

    -- Restrict execution permissions to the 'pgbouncer' role
    REVOKE EXECUTE ON FUNCTION public.lookup(name) FROM PUBLIC;
    GRANT EXECUTE ON FUNCTION public.lookup(name) TO pgbouncer;
    ```

### `static`

Manages a static file server using Nginx.

*   **`oblivion static up`**
    *   Builds a custom Nginx image (`cansu.dev-static-nginx` by default) if it doesn't exist, including configuration for auto-indexing.
    *   Starts the Nginx container, binding the configured host port to container port 80.
    *   Mounts the `[Static].static_path` from the host into the container.
*   **`oblivion static permissions`**
    *   **Requires `sudo` and the `acl` package.**
    *   Sets complex ownership and permissions on the `[Static].static_path`.
    *   **Purpose:** Allows the specified `uploader_user` (and their group) to write files, while ensuring the Nginx process (running as a different user inside the container, typically `nginx` or `www-data`) can read them. Uses `chown`, `chmod`, and `setfacl` for fine-grained control and default ACLs for new files/directories. Review the `cmd/static.go:staticChmod` function for exact commands.
*   **`fancyindex` Note:** If the themed directory index (`fancyindex`) doesn't load correctly after `oblivion static up` (e.g., 404s for header/footer), you might need to manually ensure the contents of `cmd/config/static/fancyindex/` are present at `<your_static_path>/fancyindex/`. Ideally, these files should be copied into the custom Nginx image during the build.

### `kuma`

Manages an Uptime Kuma instance.

*   **`oblivion kuma up`**
    *   Pulls the Uptime Kuma image (`louislam/uptime-kuma:1` by default).
    *   Creates a data volume (`kuma_kuma_data` by default).
    *   Starts the container, exposing the configured port (default `3001`).
    *   Connects the container to the `database_network_name` and `uptime_network_name` (by default), allowing it to monitor services on those networks using their container names (e.g., `cansu.dev-pg-primary:5432`).

### `observer`

Manages a monitoring and observability stack.

*   **Required Secrets:**
    *   `/Grafana/Admin/Username`
    *   `/Grafana/Admin/Password`
*   **`oblivion observer up`**
    *   **Important:** Verify and update the host paths in `[Observer].Binds` in your `.oblivion.toml` before running!
    *   Pulls images for Grafana, Prometheus, Loki, cAdvisor, Node Exporter, and Alertmanager.
    *   Creates volumes for Grafana and Prometheus data.
    *   Starts all component containers with appropriate configurations, port bindings, and network attachments (`grafana_bridge`, `loki_bridge`).
    *   Configures Grafana admin credentials using secrets from 1Password.

### `redis`

Manages a Redis-compatible cache/database (using DragonflyDB).

*   **Required Secrets:**
    *   `/Redis/password`
*   **`oblivion redis up`**
    *   Pulls the DragonflyDB image (`docker.dragonflydb.io/dragonflydb/dragonfly` by default).
    *   Creates a data volume (`dragonflydata` by default).
    *   Starts the container, exposing the configured port (default `6379`) and requiring the password fetched from 1Password.
    *   Connects to the `database_network_name`.

### `playground`

Manages a custom backend service defined by the Cansu.

*   **Required Secrets:**
    *   `/Postgres/Playground/username`
    *   `/Postgres/Playground/password`
    *   `/Redis/password` (assumes same Redis instance as `redis up`)
    *   `/Hugging Face/API Key`
*   **`oblivion playground up`**
    *   Clones the repository specified in `[Playground.Backend].repository` into a temporary directory.
    *   Builds a Docker image (`playground-backend` by default) from the `backend` subdirectory of the cloned repo.
    *   Starts the container, injecting database URLs, Redis URLs, Hugging Face tokens, etc., as environment variables using secrets from 1Password.
    *   Connects to `database_network_name` and `loki_network_name`.
    *   **Note:** Starts the container with elevated privileges (`seccomp:unconfined`, `SYS_ADMIN`, host PID/Cgroup namespaces, Docker socket mount). This is likely required for the backend's specific function (e.g., running code, interacting with Docker) and implies security considerations.

### Backup Scripts (`backup/`)

These are helper scripts for backing up PostgreSQL using `wal-g` to an S3-compatible backend (like Cloudflare R2). **They require manual setup and execution.**

1.  **`backup/.env`:** This file defines environment variables needed by `wal-g` and `op run`. You **must** edit this file and ensure the `op://` references point to the correct secrets in your 1Password vault.
2.  **`backup/link.sh`:**
    *   **Purpose:** `wal-g` often needs direct access to the PostgreSQL data directory. Docker volume paths can be complex. This script finds the host path of the primary PostgreSQL container's data volume and creates a symbolic link at `/var/lib/postgresql/data` (a common default path) pointing to it.
    *   **Requires `sudo` and `jq`.**
    *   Run this script once (or whenever the volume path might change, though it's usually stable): `sudo ./backup/link.sh`
3.  **`backup/backup.sh`:**
    *   **Purpose:** Performs a `wal-g backup-push`. It uses `op run --env-file=.env` to inject secrets (like AWS keys, libsodium key, PG password) from 1Password into the `sudo -E wal-g ...` command environment. (`-E` preserves the environment for `sudo`).
    *   **Requires `sudo`.**
    *   Set your `OP_SERVICE_ACCOUNT_TOKEN` environment variable first.
    *   Run the script: `./backup/backup.sh` (Consider scheduling this with `cron`).

## Example System Configuration (my Setup)

This section contains notes relevant to the my specific server environment (Debian/Ubuntu). Adapt as needed for your OS/firewall.

### Firewall (`ufw`)

```bash
# Install ufw if needed
sudo apt update && sudo apt install ufw

# Ensure IPv6 is enabled (usually is by default)
# Check/edit /etc/default/ufw if necessary: IPV6=yes

# Set defaults (Deny incoming, Allow outgoing)
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Allow essential access
sudo ufw allow ssh # Or your custom SSH port

# Allow access for web traffic reverse proxy (e.g., Cloudflare Tunnel, Caddy, Nginx)
sudo ufw allow http  # Port 80
sudo ufw allow https # Port 443

# Allow specific ports if needed (e.g., Cloudflare Tunnel alternative port)
# sudo ufw allow 7844

# Allow access *from* your specific trusted IP address (Replace with your IP)
# sudo ufw allow from 203.0.113.101

# Enable the firewall
sudo ufw enable

# Check status
sudo ufw status verbose
```

## Development

*   **Linting:** Uses `golangci-lint`. Run `golangci-lint run` (configuration is in `.golangci.yml`).

## Adaptation / Contribution

This project is primarily for personal use. If you find parts useful, feel free to fork and adapt it. Contributions are unlikely to be merged unless they align with the my specific needs or improve the core tooling in a generally applicable way.

## License
GNU General Public License v3.0

Permissions of this strong copyleft license are conditioned on making available complete source code of licensed works and modifications, which include larger works using a licensed work, under the same license. Copyright and license notices must be preserved. Contributors provide an express grant of patent rights.
