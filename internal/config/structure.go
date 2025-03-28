package config

type Root struct {
	Docker   DockerConfig
	Networks NetworkConfig
	Postgres PostgresConfig
	Onepass  OnepasswordConfig `toml:"onepass"`
	Static   StaticConfig
	Kuma     KumaConfig
}

type DockerConfig struct {
	Socket string
}

type NetworkConfig struct {
	DatabaseNetworkName string `toml:"database_network_name"`
	UptimeNetworkName   string `toml:"uptime_network_name"`
}

type OnepasswordConfig struct {
	VaultName string `toml:"vault_name"`
}

type PostgresConfig struct {
	DB             string `toml:"db"`
	PrimaryPort    string `toml:"primary_port"`
	BouncerPort    string `toml:"bouncer_port"`
	PrimaryName    string `toml:"primary_name"`
	ReplicaName    string `toml:"replica_name"`
	BouncerName    string `toml:"bouncer_name"`
	PrimaryImage   string `toml:"primary_image"`
	ReplicaImage   string `toml:"replica_image"`
	BouncerImage   string `toml:"bouncer_image"`
	PrimaryDataVol string `toml:"primary_data_vol"`
	ReplicaDataVol string `toml:"replica_data_vol"`
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

var Config Root
