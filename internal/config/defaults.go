package config

// this wont override your config https://stackoverflow.com/a/30445480
func (c *Root) SetDefaults() {
	c.Docker.Socket = "unix:///var/run/docker.sock"
	c.Postgres.DB = "postgres"
	c.Postgres.Primary.Port = "5432"
	c.Postgres.Bouncer.Port = "6432"
	c.Postgres.Primary.Name = "cansu.dev-pg-primary"
	c.Postgres.Replica.Name = "cansu.dev-pg-replica"
	c.Postgres.Bouncer.Name = "cansu.dev-pg-bouncer"
	c.Postgres.Primary.Image = "postgres:17"
	c.Postgres.Replica.Image = "postgres:17"
	c.Postgres.Bouncer.Image = "edoburu/pgbouncer"
	c.Postgres.Primary.Volume = "pg_primary_data"
	c.Postgres.Replica.Volume = "pg_replica_data"
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
	c.Networks.GrafanaNetworkName = "grafana_bridge"
	c.Networks.LokiNetworkName = "loki_bridge"
	c.Observer.ContainerNames.Grafana = "cansu.dev-observer-grafana"
	c.Observer.ContainerNames.Prometheus = "cansu.dev-observer-prometheus"
	c.Observer.ContainerNames.Loki = "cansu.dev-observer-loki"
	c.Observer.ContainerNames.Cadvisor = "cansu.dev-observer-cadvisor"
	c.Observer.ContainerNames.NodeExporter = "cansu.dev-observer-node_exporter"
	c.Observer.ContainerNames.Alertmanager = "cansu.dev-observer-alertmanager"
	c.Observer.Ports.Grafana = "3000"
	c.Observer.Ports.Prometheus = "9090"
	c.Observer.Ports.NodeExporter = "9100"
	c.Observer.Ports.Alertmanager = "9093"
	c.Observer.Ports.Cadvisor = "8080"
	c.Observer.Ports.Loki = "3169"
	c.Observer.Volumes.Grafana = "grafana_data"
	c.Observer.Volumes.Prometheus = "prometheus_data"
	c.Observer.Images.Grafana = "grafana/grafana:latest"
	c.Observer.Images.Prometheus = "prom/prometheus:latest"
	c.Observer.Images.NodeExporter = "quay.io/prometheus/node-exporter:latest"
	c.Observer.Images.Alertmanager = "prom/alertmanager:latest"
	c.Observer.Images.Cadvisor = "gcr.io/cadvisor/cadvisor"
	c.Observer.Images.Loki = "grafana/loki:latest"
	c.Observer.Binds.Prometheus = "/Users/canercetin/Git/oblivion/cmd/config/prometheus/"
	c.Observer.Binds.Grafana = "/Users/canercetin/Git/oblivion/cmd/config/grafana/"
	c.Observer.Binds.Alertmanager = "/Users/canercetin/Git/oblivion/cmd/config/alertmanager/"
	c.Observer.Binds.Loki = "/Users/canercetin/Git/oblivion/cmd/config/loki"
}
