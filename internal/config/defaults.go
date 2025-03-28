package config

// this wont override your config https://stackoverflow.com/a/30445480
func (c *Root) SetDefaults() {
	c.Docker.Socket = "unix:///var/run/docker.sock"
	c.Postgres.DB = "postgres"
	c.Postgres.PrimaryPort = "5432"
	c.Postgres.BouncerPort = "6432"
	c.Postgres.PrimaryName = "cansu.dev-pg-primary"
	c.Postgres.ReplicaName = "cansu.dev-pg-replica"
	c.Postgres.BouncerName = "cansu.dev-pg-bouncer"
	c.Postgres.PrimaryImage = "postgres:17"
	c.Postgres.ReplicaImage = "postgres:17"
	c.Postgres.BouncerImage = "edoburu/pgbouncer"
	c.Postgres.PrimaryDataVol = "pg_primary_data"
	c.Postgres.ReplicaDataVol = "pg_replica_data"
	c.Networks.DatabaseNetworkName = "database_bridge"
	c.Networks.UptimeNetworkName = "uptime_bridge"
	c.Onepass.VaultName = "Server"
	c.Static.UploaderUser = "caner"
	c.Static.StaticPath = "/var/www/servers/cansu.dev/static"
	c.Static.Port = "44444"
	c.Static.ImageName = "cansu.dev-static-nginx"
	c.Static.ContainerName = "file-server"
	c.Kuma.ContainerName = "uptime"
	c.Kuma.ImageName = "louislam/uptime-kuma:1"
	c.Kuma.Port = "3001"
	c.Kuma.DataVolume = "kuma_kuma_data"
}
