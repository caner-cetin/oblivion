package config

type Root struct {
	Docker   DockerConfig      `toml:"Docker"`
	Networks NetworkConfig     `toml:"Networks"`
	Postgres PostgresConfig    `toml:"Postgres"`
	Onepass  OnepasswordConfig `toml:"Onepass"`
	Static   StaticConfig      `toml:"Static"`
	Kuma     KumaConfig        `toml:"Kuma"`
	Observer ObserverConfig    `toml:"Observer"`
}

type DockerConfig struct {
	Socket string `toml:"Socket"`
}

type NetworkConfig struct {
	DatabaseNetworkName string `toml:"database_network_name"`
	UptimeNetworkName   string `toml:"uptime_network_name"`
	GrafanaNetworkName  string `toml:"grafana_network_name"`
	LokiNetworkName     string `toml:"loki_network_name"`
}

type OnepasswordConfig struct {
	VaultName string `toml:"vault_name"`
}

type PostgresConfig struct {
	DB      string                 `toml:"db"`
	Primary PostgresInstanceConfig `toml:"Primary"`
	Replica PostgresInstanceConfig `toml:"Replica"`
	Bouncer PostgresInstanceConfig `toml:"Bouncer"`
}

type PostgresInstanceConfig struct {
	Port   string `toml:"port"`
	Name   string `toml:"name"`
	Image  string `toml:"image"`
	Volume string `toml:"volume"`
}

type StaticConfig struct {
	UploaderUser  string `toml:"uploader_user"`
	StaticPath    string `toml:"static_path"`
	Port          string `toml:"port"`
	ImageName     string `toml:"image_name"`
	ContainerName string `toml:"container_name"`
}

type KumaConfig struct {
	ContainerName string `toml:"container_name"`
	ImageName     string `toml:"image_name"`
	Port          string `toml:"port"`
	DataVolume    string `toml:"data_volume"`
}

type ObserverConfig struct {
	Volumes        ObserverInstanceConfig `toml:"Volumes"`
	Binds          ObserverInstanceConfig `toml:"Binds"`
	ContainerNames ObserverInstanceConfig `toml:"ContainerNames"`
	Ports          ObserverInstanceConfig `toml:"Ports"`
	Images         ObserverInstanceConfig `toml:"Images"`
}

type ObserverInstanceConfig struct {
	Grafana      string `toml:"grafana"`
	Prometheus   string `toml:"prometheus"`
	NodeExporter string `toml:"node_exporter"`
	Alertmanager string `toml:"alertmanager"`
	Cadvisor     string `toml:"cadvisor"`
	Loki         string `toml:"loki"`
	Promtail     string `toml:"promtail"`
}

var Config Root
