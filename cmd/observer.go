package cmd

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	observerUpCmd = &cobra.Command{
		Use: "up",
		Run: WrapCommandWithResources(observerUp, ResourceConfig{Resources: []ResourceType{ResourceDocker, ResourceOnePassword}, Networks: []Network{NetworkDatabase, NetworkUptime, NetworkGrafana, NetworkLoki}}),
	}
	observerCmd = &cobra.Command{
		Use: "observer",
		Long: `
Look into the night sky
Looking towards the big lights
Looking out to be free
Suddenly something passes by my window

I feel it in the darkness
I need to feel it sometimes
Following the street lamps
Warning that we're meant to leave behind

Going to take a spaceship
Fly back to the stars
Alien observer in a world that isn't mine

I feel it in the darkness
I need to feel it sometimes
Following the street lamps
Warning that we're meant to leave behind

Going to take a spaceship
Fly back to the stars
Alien observer in a world that isn't mine

Going to take a spaceship
Fly back to the stars
Alien observer in a world that isn't mine
You might also like
Going to take a spaceship
Fly back to the stars
Alien observer in a world that isn't mine		
`,
	}
)

func getObserverCmd() *cobra.Command {
	observerCmd.AddCommand(observerUpCmd)
	return observerCmd
}

func observerUp(cmd *cobra.Command, args []string) {
	app := GetApp(cmd)
	if err := app.createVolumeIfNotExists(cfg.Observer.Volumes.Grafana, nil); err != nil {
		log.Error().Err(err).Msg("failed to create grafana volume")
		return
	}
	if err := app.createVolumeIfNotExists(cfg.Observer.Volumes.Prometheus, nil); err != nil {
		log.Error().Err(err).Msg("failed to create prometheus volume")
		return
	}
	if err := app.cadvisorUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.alertmanagerUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.nodeExporterUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}

	if err := app.prometheusUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.grafanaUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.lokiUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
	if err := app.promtailUp(); err != nil {
		log.Error().Err(err).Send()
		return
	}
}

func (a *AppCtx) cadvisorUp() error {
	if err := a.pullImageIfNotExists(cfg.Observer.Images.Cadvisor); err != nil {
		return fmt.Errorf("failed to pull cadvisor image: %w", err)
	}
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Cadvisor)
	if err != nil {
		return fmt.Errorf("failed to check existence of cadvisor: %w", err)
	}
	if exists {
		color.Cyan("cadvisor running")
		return nil
	}
	a.Spinner.Prefix = "creating cadvisor"
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.Images.Cadvisor,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("8080/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.Cadvisor}},
			},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/",
					Target:   "/rootfs",
					ReadOnly: true,
					BindOptions: &mount.BindOptions{
						ReadOnlyForceRecursive: true,
					},
				},
				{
					Type:   mount.TypeBind,
					Source: "/var/run",
					Target: "/var/run",
				},
				{
					Type:     mount.TypeBind,
					Source:   "/sys",
					Target:   "/sys",
					ReadOnly: true,
					BindOptions: &mount.BindOptions{
						ReadOnlyForceRecursive: true,
					},
				},
				{
					Type:     mount.TypeBind,
					Source:   "/var/lib/docker",
					Target:   "/var/lib/docker",
					ReadOnly: true,
					BindOptions: &mount.BindOptions{
						ReadOnlyForceRecursive: true,
					},
				},
			},
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Cadvisor,
	)
	if err != nil {
		return fmt.Errorf("failed to create cadvisor container: %w", err)
	}
	a.Spinner.Prefix = "starting cadvisor"
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start cadvisor container: %w", err)
	}
	return nil
}

func (a *AppCtx) prometheusUp() error {
	if err := a.pullImageIfNotExists(cfg.Observer.Images.Prometheus); err != nil {
		return fmt.Errorf("failed to pull prometheus image: %w", err)
	}
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Prometheus)
	if err != nil {
		return fmt.Errorf("failed to check existence of prometheus container: %w", err)
	}
	if exists {
		color.Cyan("prometheus running")
		return nil
	}
	a.Spinner.Prefix = "creating prometheus container"
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.Images.Prometheus,
			Cmd: []string{
				"--config.file=/etc/prometheus/prometheus.yml",
				"--storage.tsdb.path=/prometheus",
				"--web.console.libraries=/usr/share/prometheus/console_libraries",
				"--web.console.templates=/usr/share/prometheus/consoles",
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			PortBindings: nat.PortMap{
				nat.Port("9090/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.Prometheus}},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: cfg.Observer.Volumes.Prometheus,
					Target: "/prometheus",
				},
				{
					Type:   mount.TypeBind,
					Source: cfg.Observer.Binds.Prometheus,
					Target: "/etc/prometheus/",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Prometheus)
	if err != nil {
		return fmt.Errorf("failed to create prometheus container: %w", err)
	}
	a.Spinner.Prefix = "starting prometheus"
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start prometheus: %w", err)
	}
	return nil
}

func (a *AppCtx) alertmanagerUp() error {
	if err := a.pullImageIfNotExists(cfg.Observer.Images.Alertmanager); err != nil {
		return fmt.Errorf("failed to pull alertmanager image: %w", err)
	}
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Alertmanager)
	if err != nil {
		return fmt.Errorf("failed to check existence of alertmanager container: %w", err)
	}
	if exists {
		color.Cyan("alertmanager running")
		return nil
	}
	a.Spinner.Prefix = "creating alertmanager container"
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.Images.Alertmanager,
			Cmd: []string{
				"--config.file=/etc/alertmanager/config.yml",
				"--storage.path=/alertmanager",
			},
		},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			PortBindings: nat.PortMap{
				nat.Port("9093/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.Alertmanager}},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: cfg.Observer.Binds.Alertmanager,
					Target: "/etc/alertmanager/",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Alertmanager)
	if err != nil {
		return fmt.Errorf("failed to create alrtmanager container: %w", err)
	}
	a.Spinner.Prefix = "starting alertmanager"
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start alertmanager: %w", err)
	}
	return nil
}

func (a *AppCtx) nodeExporterUp() error {
	if err := a.pullImageIfNotExists(cfg.Observer.Images.NodeExporter); err != nil {
		return fmt.Errorf("failed to pull node exporter image: %w", err)
	}
	exists, err := a.containerExists(cfg.Observer.ContainerNames.NodeExporter)
	if err != nil {
		return fmt.Errorf("failed to check existence of node exporter container: %w", err)
	}
	if exists {
		color.Cyan("node exporter running")
		return nil
	}
	a.Spinner.Prefix = "creating node exporter container"
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.Images.NodeExporter,
			Cmd: []string{
				"--path.rootfs=/host",
				"--collector.filesystem.ignored-mount-points",
				"^/(sys|proc|dev|host|etc|rootfs/var/lib/docker/containers|rootfs/var/lib/docker/overlay2|rootfs/run/docker/netns|rootfs/var/lib/docker/aufs)($$|/)",
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("9100/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.NodeExporter}},
			},
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/",
					Target:   "/host",
					ReadOnly: true,
					BindOptions: &mount.BindOptions{
						ReadOnlyForceRecursive: true,
						Propagation:            mount.PropagationRSlave,
					},
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.NodeExporter,
	)
	if err != nil {
		return fmt.Errorf("failed to create node exporter container: %w", err)
	}
	a.Spinner.Prefix = "starting node exporter"
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start node exporter: %w", err)
	}
	return nil
}

func (a *AppCtx) grafanaUp() error {
	if err := a.pullImageIfNotExists(cfg.Observer.Images.Grafana); err != nil {
		return fmt.Errorf("failed to pull grafana image: %w", err)
	}
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Grafana)
	if err != nil {
		return fmt.Errorf("failed to check existence of grafana container: %w", err)
	}
	if exists {
		color.Cyan("grafana running")
		return nil
	}
	admin_username, err := a.Vault.Client.Secrets().Resolve(a.Context, a.Vault.Prefix+"/Grafana/Admin/Username")
	if err != nil {
		return fmt.Errorf("failed to resolve grafana username password: %w", err)
	}
	admin_password, err := a.Vault.Client.Secrets().Resolve(a.Context, a.Vault.Prefix+"/Grafana/Admin/Password")
	if err != nil {
		return fmt.Errorf("failed to resolve grafana admin password: %w", err)
	}
	a.Spinner.Prefix = "creating grafana container"
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.Images.Grafana,
			Env: []string{
				"GF_USERS_ALLOW_SIGN_UP=false",
				"GF_SECURITY_ADMIN_USER=" + admin_username,
				"GF_SECURITY_ADMIN_PASSWORD=" + admin_password,
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("3000/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.Grafana}},
			},
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: cfg.Observer.Volumes.Grafana,
					Target: "/var/lib/grafana",
				},
				{
					Type:   mount.TypeBind,
					Source: cfg.Observer.Binds.Grafana,
					Target: "/etc/grafana/",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName, cfg.Networks.LokiNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Grafana,
	)
	if err != nil {
		return fmt.Errorf("failed to create grafana container: %w", err)
	}
	a.Spinner.Prefix = "starting grafana"
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start grafana: %w", err)
	}
	return nil
}

func (a *AppCtx) lokiUp() error {
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Loki)
	if err != nil {
		return fmt.Errorf("failed to check existence of loki container: %w", err)
	}
	if exists {
		color.Cyan("loki container running")
		return nil
	}
	if err := a.pullImageIfNotExists(cfg.Observer.Images.Loki); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image:        cfg.Observer.Images.Loki,
			ExposedPorts: nat.PortSet{nat.Port("3169/tcp"): struct{}{}},
			Cmd:          []string{"-config.file=/etc/loki/config.yaml"},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: cfg.Observer.Binds.Loki,
					Target: "/etc/loki/",
				},
			},
			PortBindings: nat.PortMap{
				nat.Port("3169/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Observer.Ports.Loki}},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.GrafanaNetworkName, cfg.Networks.LokiNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Loki,
	)
	if err != nil {
		return fmt.Errorf("failed to create loki container: %w", err)
	}
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start loki container: %w", err)
	}
	return nil
}

func (a *AppCtx) promtailUp() error {
	exists, err := a.containerExists(cfg.Observer.ContainerNames.Promtail)
	if err != nil {
		return fmt.Errorf("failed to check existence of promtail container: %w", err)
	}
	if exists {
		color.Cyan("promtail container running")
		return nil
	}
	if err = a.pullImageIfNotExists(cfg.Observer.Images.Promtail); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	resp, err := a.Docker.Client.ContainerCreate(a.Context,
		&container.Config{
			Image: cfg.Observer.ContainerNames.Promtail,
			Cmd:   []string{"-config.file=/etc/promtail/config.yml"},
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/var/log",
					Target: "/var/log",
				},
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: a.getNetworks(cfg.Networks.LokiNetworkName, cfg.Networks.GrafanaNetworkName),
		},
		nil,
		cfg.Observer.ContainerNames.Promtail,
	)
	if err != nil {
		return fmt.Errorf("failed to create promtail container: %w", err)
	}
	if err := a.Docker.Client.ContainerStart(a.Context, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start promtail container: %w", err)
	}
	return nil
}
