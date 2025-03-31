# oblivion

Docker configurations for my server `*.cansu.dev/*` (WIP)
- [oblivion](#oblivion)
  - [usage](#usage)
    - [configuration](#configuration)
      - [1password](#1password)
      - [config](#config)
    - [networks](#networks)
    - [postgres](#postgres)
    - [static](#static)
    - [kuma](#kuma)
    - [firewall](#firewall)


## usage

### configuration
#### 1password
**1password is required**, at least for now.

export the `OP_SERVICE_ACCOUNT_TOKEN` env before running commands like `postgres`, `redis`, or anything that requires credentials. 
```bash
export OP_SERVICE_ACCOUNT_TOKEN=ops...
```

also read https://developer.1password.com/docs/sdks/concepts/#secret-references, for example, when I say `required secret "/Postgres/Replicator/username"` it means that
```
create a secure note called "Postgres" under the vault configured in oblivion.toml
create a section called "Replicator" under "Postgres"
create an item called "username" under "/Postgres/Replicator" and fill there
```
#### config
modify `.oblivion.toml` if you wwant and copy under home directory
```bash
cp .oblivion.toml $HOME/.oblivion.toml
``` 

### networks
```bash
oblivion networks up
```
creates all required networks

### postgres

required secrets
```
"/Postgres/Replicator/username"
"/Postgres/Replicator/password"
"/Postgres/Root/username"
"/Postgres/Root/password"
"/Postgres/Bouncer/username"
"/Postgres/Bouncer/password"
```

```bash
oblivion postgres up
```

execute the following script for databases that you want to connect with bouncer
```sql
CREATE ROLE pgbouncer LOGIN;
ALTER USER pgbouncer WITH PASSWORD 'i_look_cute_in_maid_outfit';
 
CREATE FUNCTION public.lookup (
   INOUT p_user     name,
   OUT   p_password text
) RETURNS record
   LANGUAGE sql SECURITY DEFINER SET search_path = pg_catalog AS
$$SELECT usename, passwd FROM pg_shadow WHERE usename = p_user$$;
 
-- make sure only 'pgbouncer' can use the function
REVOKE EXECUTE ON FUNCTION public.lookup(name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.lookup(name) TO pgbouncer;
```
### static
```bash
oblivion static up
```
run
```bash
oblivion static permissions
```
to setup required permissions/ownerships for both the uploader account and the `nginx` container.

if `fancyindex` is not there after initialization (`404 not found` on both header and footer at the index page), just move the contents of `cmd/config/static/fancyindex` to your static path, e.g `/var/www/servers/cansu.dev/static/fancyindex`.
### kuma
```bash
oblivion kuma up
```
`uptime` container is connected to all networks declared in config, so you can reference all services with their container names. like, `postgres://user:password@cansu.dev-pg-primary:5432` and so on. 
### firewall
```bash
sudo apt install ufw
# ensure that IPv6 is enabled
sudo nano /etc/default/ufw
# deny all
sudo ufw default deny incoming
# allow all
sudo ufw default allow outgoing
# allow ssh
sudo ufw allow ssh
# allow cloudflare tunnels
sudo ufw allow 443
# change to your reverse proxy, cloudflare zero trust is using 7844 and 443
sudo ufw allow 7844
# allow yourself  
sudo ufw allow from 203.0.113.101
sudo ufw allow enable
```

